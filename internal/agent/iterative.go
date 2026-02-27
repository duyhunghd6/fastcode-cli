package agent

import (
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/duyhunghd6/fastcode-cli/internal/llm"
	"github.com/duyhunghd6/fastcode-cli/internal/types"
)

// IterativeAgent manages multi-round retrieval with confidence and cost control.
type IterativeAgent struct {
	client       *llm.Client
	toolExecutor *ToolExecutor
	config       AgentConfig

	// State tracked across rounds
	gatheredElements []types.CodeElement
	totalTokensUsed  int
	rounds           int

	// Adaptive parameters (set per query, mirroring Python)
	maxIterations       int
	confidenceThreshold int
	adaptiveLineBudget  int

	// History tracking (mirroring Python)
	toolCallHistory  []toolCallRecord
	iterationHistory []map[string]any
}

// toolCallRecord tracks a tool call for history display in prompts.
type toolCallRecord struct {
	Round      int
	ToolName   string
	Parameters map[string]any
}

// AgentConfig holds configuration for the iterative agent.
type AgentConfig struct {
	MaxRounds           int     // Base maximum retrieval rounds (default: 4)
	ConfidenceThreshold int     // Base confidence threshold (default: 95)
	MaxTokenBudget      int     // Maximum tokens to consume (default: 50000)
	MaxTotalLines       int     // Maximum total lines budget (default: 12000)
	Temperature         float64 // LLM temperature (default: 0.2)
	MaxTokensAgent      int     // Max tokens for agent LLM calls (default: 8000)
}

// DefaultAgentConfig returns sensible defaults matching Python.
func DefaultAgentConfig() AgentConfig {
	return AgentConfig{
		MaxRounds:           4,
		ConfidenceThreshold: 95,
		MaxTokenBudget:      50000,
		MaxTotalLines:       12000,
		Temperature:         0.2,
		MaxTokensAgent:      8000,
	}
}

// RoundResult holds the output of a single agent round.
type RoundResult struct {
	Round      int                 `json:"round"`
	Confidence int                 `json:"confidence"`
	Reasoning  string              `json:"reasoning"`
	ToolCalls  []ToolCall          `json:"tool_calls,omitempty"`
	KeepFiles  []string            `json:"keep_files,omitempty"`
	Elements   []types.CodeElement `json:"elements,omitempty"`

	// Round 1 specific fields
	QueryComplexity  int            `json:"query_complexity,omitempty"`
	QueryEnhancement map[string]any `json:"query_enhancement,omitempty"`
}

// ToolCall represents a tool the agent wants to invoke.
// Supports both simple (name+arg) and parameterized (name+parameters) formats.
type ToolCall struct {
	Name       string         `json:"name,omitempty"`
	Tool       string         `json:"tool,omitempty"` // Python uses "tool" field
	Arg        string         `json:"arg,omitempty"`
	Parameters map[string]any `json:"parameters,omitempty"` // Python-style parameters
}

// GetToolName returns the effective tool name.
func (tc ToolCall) GetToolName() string {
	if tc.Tool != "" {
		return tc.Tool
	}
	return tc.Name
}

// GetArg returns the effective argument string for tool execution.
func (tc ToolCall) GetArg() string {
	if tc.Arg != "" {
		return tc.Arg
	}
	// Build arg from parameters
	if st, ok := tc.Parameters["search_term"]; ok {
		return fmt.Sprintf("%v", st)
	}
	if p, ok := tc.Parameters["path"]; ok {
		return fmt.Sprintf("%v", p)
	}
	return ""
}

// RetrievalResult holds the final output of the iterative retrieval.
type RetrievalResult struct {
	Elements   []types.CodeElement `json:"elements"`
	Rounds     int                 `json:"rounds"`
	Confidence int                 `json:"confidence"`
	StopReason string              `json:"stop_reason"`
	Metadata   map[string]any      `json:"metadata,omitempty"`
}

// NewIterativeAgent creates a new iterative retrieval agent.
func NewIterativeAgent(client *llm.Client, toolExec *ToolExecutor, cfg AgentConfig) *IterativeAgent {
	if cfg.MaxRounds == 0 {
		cfg = DefaultAgentConfig()
	}
	return &IterativeAgent{
		client:       client,
		toolExecutor: toolExec,
		config:       cfg,
	}
}

// Retrieve performs iterative retrieval for the given query.
// Mirrors Python's retrieve_with_iteration method.
func (ia *IterativeAgent) Retrieve(query string, pq *ProcessedQuery) (*RetrievalResult, error) {
	ia.gatheredElements = nil
	ia.totalTokensUsed = 0
	ia.rounds = 0
	ia.toolCallHistory = nil
	ia.iterationHistory = nil

	// ─── Round 1: Initial assessment (no code context yet) ───
	round1Result, err := ia.executeRound1(query, pq)
	if err != nil {
		log.Printf("[agent] round 1 error: %v", err)
		return &RetrievalResult{StopReason: "error"}, err
	}

	// Record round 1 tool calls in history
	ia.recordToolCalls(1, round1Result.ToolCalls)

	// Initialize adaptive parameters based on query complexity from round 1
	queryComplexity := round1Result.QueryComplexity
	if queryComplexity == 0 {
		queryComplexity = pq.Complexity
	}
	ia.initializeAdaptiveParams(queryComplexity)

	// Execute round 1 tool calls using REAL filesystem search (matching Python)
	// Python's flow: tool calls → file candidates → LLM selects relevant files → indexed elements
	var allCandidates []FileCandidate
	if len(round1Result.ToolCalls) > 0 {
		for _, tc := range round1Result.ToolCalls {
			toolName := tc.GetToolName()
			params := tc.Parameters

			if toolName == "search_codebase" || toolName == "search_code" {
				searchTerm, _ := params["search_term"].(string)
				if searchTerm == "" {
					searchTerm = tc.GetArg()
				}
				filePattern, _ := params["file_pattern"].(string)
				if filePattern == "" {
					filePattern = "*"
				}
				useRegex, _ := params["use_regex"].(bool)

				candidates := ia.toolExecutor.ExecuteSearchCodebase(searchTerm, filePattern, useRegex)
				log.Printf("[agent] search_codebase(%q) returned %d candidates", searchTerm, len(candidates))
				allCandidates = append(allCandidates, candidates...)

			} else if toolName == "list_directory" || toolName == "list_files" {
				dirPath, _ := params["path"].(string)
				if dirPath == "" {
					dirPath = tc.GetArg()
				}
				candidates := ia.toolExecutor.ExecuteListDirectory(dirPath)
				log.Printf("[agent] list_directory(%q) returned %d candidates", dirPath, len(candidates))
				allCandidates = append(allCandidates, candidates...)
			}
		}
	}

	// If we have file candidates, use LLM to select the most relevant ones
	if len(allCandidates) > 0 {
		// Deduplicate candidates by file path
		seen := make(map[string]bool)
		var uniqueCandidates []FileCandidate
		for _, c := range allCandidates {
			if !seen[c.FilePath] {
				seen[c.FilePath] = true
				uniqueCandidates = append(uniqueCandidates, c)
			}
		}
		log.Printf("[agent] %d unique file candidates for LLM selection", len(uniqueCandidates))

		// LLM file selection (matching Python's _llm_select_elements_with_granularity)
		selectedFiles := ia.llmSelectFiles(query, uniqueCandidates)
		log.Printf("[agent] LLM selected %d files", len(selectedFiles))

		// Convert selected files to indexed elements
		for _, filePath := range selectedFiles {
			elements := ia.toolExecutor.FindElementsForFile(filePath)
			if len(elements) > 0 {
				ia.gatheredElements = append(ia.gatheredElements, elements...)
			}
		}
	}

	// Also perform standard BM25 search (baseline, like Python's _perform_standard_retrieval)
	if res, err := ia.toolExecutor.searchCode(query); err == nil && res != nil {
		ia.gatheredElements = append(ia.gatheredElements, res.Elements...)
	}

	// Deduplicate after round 1
	ia.gatheredElements = ia.removeDuplicatesWithContainment(ia.gatheredElements)

	// Record round 1 history
	totalLines := ia.calculateTotalLines(ia.gatheredElements)
	ia.iterationHistory = append(ia.iterationHistory, map[string]any{
		"round":        1,
		"confidence":   round1Result.Confidence,
		"elements":     len(ia.gatheredElements),
		"total_lines":  totalLines,
		"budget_usage": float64(totalLines) / float64(ia.adaptiveLineBudget) * 100,
	})

	ia.rounds = 1
	lastConfidence := round1Result.Confidence
	var stopReason string

	// ─── Rounds 2..N: Assessment with context ───
	for round := 2; round <= ia.maxIterations; round++ {
		ia.rounds = round

		roundResult, err := ia.executeRoundN(query, pq, round)
		if err != nil {
			log.Printf("[agent] round %d error: %v", round, err)
			stopReason = "error"
			break
		}

		// Record tool calls
		ia.recordToolCalls(round, roundResult.ToolCalls)

		// Filter elements based on keep_files
		if len(roundResult.KeepFiles) > 0 {
			ia.gatheredElements = ia.filterElementsByKeepFiles(ia.gatheredElements, roundResult.KeepFiles)
		}

		numBefore := len(ia.gatheredElements)
		lastConfidence = roundResult.Confidence

		// Log element filtering
		log.Printf("[agent] Round %d element filtering: %d -> %d elements",
			round, numBefore, len(ia.gatheredElements))
		log.Printf("[agent] Round %d confidence: %d", round, lastConfidence)

		// Calculate metrics
		totalLines = ia.calculateTotalLines(ia.gatheredElements)
		budgetUsage := float64(totalLines) / float64(ia.adaptiveLineBudget) * 100
		ia.iterationHistory = append(ia.iterationHistory, map[string]any{
			"round":        round,
			"confidence":   lastConfidence,
			"elements":     len(ia.gatheredElements),
			"total_lines":  totalLines,
			"budget_usage": budgetUsage,
		})

		// Check stopping conditions
		if lastConfidence >= ia.confidenceThreshold {
			stopReason = "confidence_threshold_reached"
			break
		}
		if ia.totalTokensUsed >= ia.config.MaxTokenBudget {
			stopReason = "budget_exhausted"
			break
		}

		// Execute round N tool calls
		if len(roundResult.ToolCalls) > 0 {
			for _, tc := range roundResult.ToolCalls {
				toolName := tc.GetToolName()
				result, err := ia.toolExecutor.Execute(toolName, tc.GetArg())
				if err != nil {
					log.Printf("[agent] tool %s error: %v", toolName, err)
					continue
				}
				ia.gatheredElements = append(ia.gatheredElements, result.Elements...)
			}
			// Deduplicate after each round
			ia.gatheredElements = ia.removeDuplicatesWithContainment(ia.gatheredElements)
		} else if lastConfidence < ia.confidenceThreshold {
			stopReason = "no_more_actions"
			break
		}
	}

	if stopReason == "" {
		stopReason = "max_rounds"
	}

	// Final deduplication
	elements := ia.removeDuplicatesWithContainment(ia.gatheredElements)

	return &RetrievalResult{
		Elements:   elements,
		Rounds:     ia.rounds,
		Confidence: lastConfidence,
		StopReason: stopReason,
		Metadata: map[string]any{
			"query_complexity": queryComplexity,
			"query_type":       pq.QueryType,
			"tokens_used":      ia.totalTokensUsed,
			"adaptive_params": map[string]any{
				"max_iterations":       ia.maxIterations,
				"confidence_threshold": ia.confidenceThreshold,
				"line_budget":          ia.adaptiveLineBudget,
			},
		},
	}, nil
}

// initializeAdaptiveParams sets dynamic thresholds matching Python's _initialize_adaptive_parameters.
func (ia *IterativeAgent) initializeAdaptiveParams(queryComplexity int) {
	// Adaptive max iterations
	ia.maxIterations = ia.config.MaxRounds
	if queryComplexity < 30 {
		ia.maxIterations = max(2, ia.config.MaxRounds)
	}

	// Adaptive confidence threshold
	if queryComplexity >= 80 {
		ia.confidenceThreshold = max(90, ia.config.ConfidenceThreshold-5)
	} else if queryComplexity >= 60 {
		ia.confidenceThreshold = max(92, ia.config.ConfidenceThreshold-3)
	} else {
		ia.confidenceThreshold = ia.config.ConfidenceThreshold
	}

	// Adaptive line budget
	maxLines := ia.config.MaxTotalLines
	if maxLines == 0 {
		maxLines = 12000
	}
	if queryComplexity <= 30 {
		ia.adaptiveLineBudget = int(float64(maxLines) * 0.6)
	} else if queryComplexity <= 60 {
		ia.adaptiveLineBudget = int(float64(maxLines) * 0.8)
	} else {
		ia.adaptiveLineBudget = maxLines
	}

	log.Printf("[agent] Adaptive params: max_iterations=%d, confidence_threshold=%d, line_budget=%d, query_complexity=%d",
		ia.maxIterations, ia.confidenceThreshold, ia.adaptiveLineBudget, queryComplexity)
}

// ─── Round 1: Initial assessment (no code context) ─────────────────

func (ia *IterativeAgent) executeRound1(query string, pq *ProcessedQuery) (*RoundResult, error) {
	prompt := ia.buildRound1Prompt(query, pq)

	response, err := ia.client.ChatCompletion([]llm.ChatMessage{
		{Role: "system", Content: "You are a precise code analysis agent. Respond in specified format only."},
		{Role: "user", Content: prompt},
	}, ia.config.Temperature, ia.config.MaxTokensAgent)
	if err != nil {
		return nil, fmt.Errorf("LLM call round 1: %w", err)
	}

	return ia.parseRound1Response(response)
}

func (ia *IterativeAgent) buildRound1Prompt(query string, pq *ProcessedQuery) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf(`You are a code analysis agent performing initial query assessment. You have NOT seen any code files yet.

**Current User Query**: %s

**Repository Structure**:
./%s

**Your Task**: Assess the query and decide on the retrieval strategy.

CONFIDENCE SCORING RULES (0-100):
- 95-100: You have complete knowledge to answer this question without needing any code files
- 80-94: You have good general knowledge but need to see specific implementation details
- 60-79: You understand the domain but need to examine the codebase structure and key files
- 40-59: The question requires detailed code inspection across multiple files
- 20-39: Complex cross-file analysis or deep implementation details needed
- 0-19: Highly specific question requiring comprehensive codebase examination

IMPORTANT: At this stage, you have NOT seen any code files yet. Base your confidence ONLY on:
1. Whether this is a general knowledge question vs specific implementation question
2. Whether the question asks about standard patterns vs custom implementation
3. Your general understanding of the technology/framework mentioned

`, query, ""))

	// Output format
	sb.WriteString(`**Output Format** (JSON only):

If confidence >= 95:
{
  "confidence": <0-100>,
  "reasoning": "Brief explanation"
}

If confidence < 95:
{
  "confidence": <0-100>,
  "query_complexity": <0-100>,
  "reasoning": "Brief explanation",
  "query_enhancement": {
    "needed": true/false,
    "refined_intent": "<intent>",
    "rewritten_query": "<optimized English query for semantic/BM25 retrieval, with key technical terms and concepts>",
    "selected_keywords": ["kw1", "kw2"],
    "pseudocode_hints": "<pseudocode or null>"
  },
  "tool_calls": [
    {"tool": "search_codebase", "parameters": {"search_term": "...", "file_pattern": "*.py", "use_regex": false}},
    {"tool": "list_directory", "parameters": {"path": "src/core"}}
  ]
}

**Query Complexity Scoring (0-100)**:
- 0-20: Simple lookup (find a function/class)
- 21-40: Single-file analysis (understand one component)
- 41-60: Multi-file analysis (trace logic across files)
- 61-80: Cross-module/architectural understanding
- 81-100: Complex debugging or system-wide refactoring questions

**Query Rewriting Guidelines**:
- Translate non-English queries to English for optimal retrieval accuracy
- Expand abbreviations and resolve references from dialogue context
- Include technical terms, class/function names, and domain-specific keywords
- Keep concise while preserving all essential meaning

**Tool Call Guidelines**:
- Use search_codebase for finding specific terms, classes, functions
  * search_term: literal text or regex pattern to find in file contents
  * file_pattern: SINGLE glob pattern per tool call to filter files (only one pattern allowed)
  * use_regex: true if search_term is regex, false for literal (default: false)

- Use list_directory to explore directory structure
  * path: directory path to list

- Maximum 10 tool calls
- Be strategic: target likely locations based on query and repo structure
- Do not use the model's native tool_calls format. Instead, include tool call instructions in your text response content in a parseable format

**CRITICAL**:
- Respond with valid JSON only
- No markdown code blocks
- No comments in JSON
- If confidence >= 95, ONLY output confidence and reasoning
`)

	return sb.String()
}

func (ia *IterativeAgent) parseRound1Response(response string) (*RoundResult, error) {
	result := &RoundResult{Round: 1}

	jsonStr := extractJSON(response)
	if jsonStr == "" {
		result.Confidence = 90
		result.Reasoning = response
		return result, nil
	}

	var parsed struct {
		Confidence       int            `json:"confidence"`
		QueryComplexity  int            `json:"query_complexity"`
		Reasoning        string         `json:"reasoning"`
		QueryEnhancement map[string]any `json:"query_enhancement"`
		ToolCalls        []ToolCall     `json:"tool_calls"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		result.Confidence = 90
		result.Reasoning = response
		return result, nil
	}

	result.Confidence = parsed.Confidence
	result.Reasoning = parsed.Reasoning
	result.QueryComplexity = parsed.QueryComplexity
	result.QueryEnhancement = parsed.QueryEnhancement
	result.ToolCalls = parsed.ToolCalls
	return result, nil
}

// ─── Round N (2+): Assessment with context ─────────────────────────

func (ia *IterativeAgent) executeRoundN(query string, pq *ProcessedQuery, round int) (*RoundResult, error) {
	prompt := ia.buildRoundNPrompt(query, pq, round)

	response, err := ia.client.ChatCompletion([]llm.ChatMessage{
		{Role: "system", Content: "You are a precise code analysis agent. Respond in specified format only."},
		{Role: "user", Content: prompt},
	}, ia.config.Temperature, ia.config.MaxTokensAgent)
	if err != nil {
		return nil, fmt.Errorf("LLM call round %d: %w", round, err)
	}

	return ia.parseRoundNResponse(response, round)
}

func (ia *IterativeAgent) buildRoundNPrompt(query string, pq *ProcessedQuery, round int) string {
	var sb strings.Builder

	// Calculate resource usage
	totalLines := ia.calculateTotalLines(ia.gatheredElements)
	remainingBudget := ia.adaptiveLineBudget - totalLines
	remainingIterations := ia.maxIterations - round
	budgetUsagePct := float64(totalLines) / float64(ia.adaptiveLineBudget) * 100

	sb.WriteString(fmt.Sprintf(`You are a cost-aware code analysis agent in round %d of iterative retrieval.

**Current User Query**: %s

**Repository Structure**:
Not available

`, round, query))

	// Resource status
	sb.WriteString(fmt.Sprintf(`
**Current Resource Usage**:
- Current code lines: %d / %d (%.1f%% used)
- Remaining budget: %d lines
- Current round: %d / %d
- Remaining iterations: %d

`, totalLines, ia.adaptiveLineBudget, budgetUsagePct, remainingBudget, round, ia.maxIterations, remainingIterations))

	// Current elements
	sb.WriteString(fmt.Sprintf("**Current Retrieved Elements**:\n%s\n", ia.formatElementsWithMetadata()))

	// Tool call history
	sb.WriteString(fmt.Sprintf("**Previous Tool Calls**:\n%s\n", ia.formatToolCallHistory(round)))

	// Confidence rules
	sb.WriteString(fmt.Sprintf(`
CONFIDENCE SCORING RULES (0-100) for Round %d:
- 95-100: Current files provide complete information to answer the query accurately
- 80-94: Files are mostly sufficient, minor details might be missing
- 60-79: Files provide good foundation but key implementations or connections are missing
- 40-59: Files are relevant but substantial information gaps exist
- 20-39: Files are only partially relevant, need significant additional context
- 0-19: Current files are insufficient or off-target

Base your confidence on:
1. Coverage of key concepts mentioned in the query
2. Presence of relevant signatures, classes, functions
3. Completeness of call chains or dependency relationships
4. Whether graph-related files fill important gaps

**IMPORTANT: Balance confidence with cost efficiency**

`, round))

	// Cost-aware guidelines
	sb.WriteString(fmt.Sprintf(`
**Cost-Aware Decision Making**:
1. **File Selection**:
   - Only remove irrelevant, redundant, or not useful files
   - Prefer class/function-level selections over entire files when possible, but select the entire file if multiple classes or functions within it are useful

2. **Confidence vs Cost Trade-off**:
   - If budget usage > 70%%: Be very selective, only keep essential files
   - If budget usage > 85%%: Only keep files critical for answering the query
   - If remaining_budget < 2000 lines: Do NOT request more tool calls unless critical gaps exist

3. **Stopping Criteria** (when to set confidence >= %d):
   - You have enough information to answer the query reasonably well
   - Additional files would provide diminishing returns
   - Budget is running low and current files are sufficient
   - Marginal benefit of more code < cost of retrieving it

4. **Tool Call Efficiency** (when confidence < %d):
   - Only request tool calls if they will find CRITICAL missing information
   - Be very specific to minimize noise
   - Do NOT repeat previous tool calls; use new terms/paths only
   - Consider if the information gap is worth the cost

`, ia.confidenceThreshold, ia.confidenceThreshold))

	// Output format
	sb.WriteString(fmt.Sprintf(`**Your Task**:
1. **Filter**: Keep files that are relevant to answering the query. If all files are potentially useful, keep all.
2. **Assess confidence**: Based on the kept files, how confident are you in answering the query?
3. **Decide on next action**:
   - If confidence >= %d OR budget is critical: STOP (set confidence >= %d)
   - If critical information is missing AND budget allows: Request targeted tool calls
   - Otherwise: STOP with current files

**Output Format** (JSON only):

If stopping (confidence >= %d or budget critical):
{
  "keep_files": ["file1.py", "file2.py"],
  "confidence": <0-100>,
  "reasoning": "Brief explanation of why these files are sufficient"
}

If continuing (confidence < %d and budget available):
{
  "keep_files": ["file1.py", "file2.py"],
  "confidence": <0-100>,
  "reasoning": "Brief explanation of what's missing",
  "tool_calls": [
    {"tool": "search_codebase", "parameters": {"search_term": "...", "file_pattern": "*.py", "use_regex": false}},
    {"tool": "list_directory", "parameters": {"path": "src/core"}}
  ]
}

**Keep Files Format**:
- Filename for file-level: "path/to/file.py"
- Class-level: "path/to/file.py:ClassName"
- Function-level: "path/to/file.py:function_name"

**Tool Call Guidelines**:
- Use search_codebase for finding specific terms, classes, functions
  * search_term: literal text or regex pattern to find in file contents
  * file_pattern: SINGLE glob pattern per tool call to filter files (only one pattern allowed)
  * use_regex: true if search_term is regex, false for literal (default: false)

- Use list_directory to explore directory structure
  * path: directory path to list

- Do NOT use the model's native tool_calls format. Instead, include tool call instructions in your text response content in a parseable format

**CRITICAL**:
- Respond with valid JSON only
- No markdown blocks
- No comments in JSON
- Be cost-conscious: fewer, more relevant files are better than many marginally useful files
`, ia.confidenceThreshold, ia.confidenceThreshold, ia.confidenceThreshold, ia.confidenceThreshold))

	return sb.String()
}

func (ia *IterativeAgent) parseRoundNResponse(response string, round int) (*RoundResult, error) {
	result := &RoundResult{Round: round}

	jsonStr := extractJSON(response)
	if jsonStr == "" {
		result.Confidence = 95
		result.Reasoning = response
		return result, nil
	}

	var parsed struct {
		Confidence int        `json:"confidence"`
		Reasoning  string     `json:"reasoning"`
		KeepFiles  []string   `json:"keep_files"`
		ToolCalls  []ToolCall `json:"tool_calls"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		result.Confidence = 95
		result.Reasoning = response
		return result, nil
	}

	result.Confidence = parsed.Confidence
	result.Reasoning = parsed.Reasoning
	result.KeepFiles = parsed.KeepFiles
	result.ToolCalls = parsed.ToolCalls
	return result, nil
}

// ─── Helpers ───────────────────────────────────────────────────────

// recordToolCalls records tool calls for prompt history (matching Python).
func (ia *IterativeAgent) recordToolCalls(round int, calls []ToolCall) {
	for _, tc := range calls {
		params := tc.Parameters
		if params == nil {
			params = map[string]any{}
			if tc.GetArg() != "" {
				params["search_term"] = tc.GetArg()
			}
		}
		ia.toolCallHistory = append(ia.toolCallHistory, toolCallRecord{
			Round:      round,
			ToolName:   tc.GetToolName(),
			Parameters: params,
		})
	}
}

// formatToolCallHistory formats tool call history for round N prompts.
func (ia *IterativeAgent) formatToolCallHistory(currentRound int) string {
	var sb strings.Builder
	for _, tc := range ia.toolCallHistory {
		if tc.Round < currentRound {
			paramsJSON, _ := json.Marshal(tc.Parameters)
			sb.WriteString(fmt.Sprintf("- Round %d: %s %s\n", tc.Round, tc.ToolName, string(paramsJSON)))
		}
	}
	if sb.Len() == 0 {
		return "None\n"
	}
	return sb.String()
}

// formatElementsWithMetadata formats gathered elements for round N prompt.
func (ia *IterativeAgent) formatElementsWithMetadata() string {
	var sb strings.Builder
	for i, elem := range ia.gatheredElements {
		if i >= 20 {
			sb.WriteString(fmt.Sprintf("\n... and %d more elements\n", len(ia.gatheredElements)-20))
			break
		}

		repoName := elem.RepoName
		if repoName == "" {
			repoName = "repo"
		}

		sb.WriteString(fmt.Sprintf("\n%d. %s/%s\n", i+1, repoName, elem.RelativePath))
		sb.WriteString(fmt.Sprintf("   Repo: %s\n", repoName))
		sb.WriteString(fmt.Sprintf("   Type: %s\n", elem.Type))

		// Source info
		source := "Retrieval"
		sb.WriteString(fmt.Sprintf("   Source: %s\n", source))

		lines := elem.EndLine - elem.StartLine + 1
		if lines <= 0 {
			lines = len(strings.Split(elem.Code, "\n"))
		}
		sb.WriteString(fmt.Sprintf("   Lines: %d\n", lines))

		if elem.Signature != "" {
			sb.WriteString(fmt.Sprintf("   - def %s\n", elem.Signature))
		}
	}
	return sb.String()
}

// calculateTotalLines calculates total lines across all elements.
func (ia *IterativeAgent) calculateTotalLines(elements []types.CodeElement) int {
	total := 0
	for _, elem := range elements {
		lines := elem.EndLine - elem.StartLine + 1
		if lines <= 0 {
			lines = len(strings.Split(elem.Code, "\n"))
		}
		total += lines
	}
	return total
}

// filterElementsByKeepFiles filters elements to only include those in the keep_files list.
func (ia *IterativeAgent) filterElementsByKeepFiles(elements []types.CodeElement, keepFiles []string) []types.CodeElement {
	if len(keepFiles) == 0 {
		return elements
	}

	keepSet := make(map[string]bool)
	for _, f := range keepFiles {
		keepSet[f] = true
		// Also add without leading repo prefix
		parts := strings.SplitN(f, "/", 2)
		if len(parts) > 1 {
			keepSet[parts[1]] = true
		}
	}

	var kept []types.CodeElement
	for _, elem := range elements {
		path := elem.RelativePath
		repoPath := ""
		if elem.RepoName != "" {
			repoPath = elem.RepoName + "/" + path
		}

		// Check various matching strategies
		if keepSet[path] || keepSet[repoPath] {
			kept = append(kept, elem)
			continue
		}

		// Check with element name suffix (path:ClassName or path:function_name)
		pathWithName := path + ":" + elem.Name
		repoPathWithName := repoPath + ":" + elem.Name
		if keepSet[pathWithName] || keepSet[repoPathWithName] {
			kept = append(kept, elem)
			continue
		}

		// Check if any keep_file is a prefix match
		for _, kf := range keepFiles {
			if strings.HasSuffix(path, kf) || strings.HasSuffix(repoPath, kf) {
				kept = append(kept, elem)
				break
			}
		}
	}

	// If filtering removed everything, return originals (safety fallback)
	if len(kept) == 0 && len(elements) > 0 {
		return elements
	}

	return kept
}

func extractJSON(s string) string {
	// Try to find JSON block in markdown code fence
	if idx := strings.Index(s, "```json"); idx >= 0 {
		start := idx + 7
		if end := strings.Index(s[start:], "```"); end >= 0 {
			return strings.TrimSpace(s[start : start+end])
		}
	}
	// Try to find raw JSON
	if idx := strings.Index(s, "{"); idx >= 0 {
		depth := 0
		for i := idx; i < len(s); i++ {
			if s[i] == '{' {
				depth++
			} else if s[i] == '}' {
				depth--
				if depth == 0 {
					return s[idx : i+1]
				}
			}
		}
	}
	return ""
}

// deduplicateElements was replaced by removeDuplicatesWithContainment to match Python's logic.
func (ia *IterativeAgent) removeDuplicatesWithContainment(elements []types.CodeElement) []types.CodeElement {
	// First remove exact ID duplicates
	seen := make(map[string]bool)
	var unique []types.CodeElement
	for _, elem := range elements {
		if !seen[elem.ID] {
			seen[elem.ID] = true
			unique = append(unique, elem)
		}
	}

	if len(unique) <= 1 {
		return unique
	}

	// Group by repo + file path
	type groupKey struct {
		repo string
		path string
	}
	groups := make(map[groupKey][]types.CodeElement)
	for _, elem := range unique {
		key := groupKey{repo: elem.RepoName, path: elem.RelativePath}
		groups[key] = append(groups[key], elem)
	}

	var final []types.CodeElement

	for _, group := range groups {
		if len(group) == 1 {
			final = append(final, group[0])
			continue
		}

		// Sort by priority (file > class > function, then line range size, then start line)
		sort.Slice(group, func(i, j int) bool {
			e1 := group[i]
			e2 := group[j]

			p1 := getTypePriority(e1.Type)
			p2 := getTypePriority(e2.Type)
			if p1 != p2 {
				return p1 > p2 // Higher priority first
			}

			s1 := e1.EndLine - e1.StartLine
			s2 := e2.EndLine - e2.StartLine
			if s1 != s2 {
				return s1 > s2 // Larger range first
			}

			return e1.StartLine < e2.StartLine
		})

		var kept []types.CodeElement
		for _, elem := range group {
			contained := false
			for _, k := range kept {
				// check if k contains elem
				// Python: kept_start <= start and end <= kept_end and (kept_start < start or end < kept_end)
				if k.StartLine <= elem.StartLine && elem.EndLine <= k.EndLine &&
					(k.StartLine < elem.StartLine || elem.EndLine < k.EndLine) {
					contained = true
					break
				}
			}
			if !contained {
				kept = append(kept, elem)
			}
		}
		final = append(final, kept...)
	}

	// Python's IterativeAgent seems to preserve original ordering (mostly), but we grouped them.
	// To preserve original order, we filter the original 'unique' list against 'final' IDs:
	finalSeen := make(map[string]bool)
	for _, f := range final {
		finalSeen[f.ID] = true
	}

	var orderedFinal []types.CodeElement
	for _, u := range unique {
		if finalSeen[u.ID] {
			orderedFinal = append(orderedFinal, u)
		}
	}

	return orderedFinal
}

func getTypePriority(t string) int {
	switch t {
	case "file":
		return 3
	case "class":
		return 2
	case "function":
		return 1
	}
	return 0
}

// ─── LLM File Selection (matching Python's _llm_select_elements_with_granularity) ───

// llmSelectFiles sends file candidates to the LLM and asks it to pick the most relevant files.
// This mirrors Python's _llm_select_elements_with_granularity() in iterative_agent.py.
func (ia *IterativeAgent) llmSelectFiles(query string, candidates []FileCandidate) []string {
	if len(candidates) == 0 {
		return nil
	}

	// Limit candidates to avoid huge prompts
	maxCandidates := 50
	if len(candidates) > maxCandidates {
		candidates = candidates[:maxCandidates]
	}

	prompt := ia.buildFileSelectionPrompt(query, candidates)

	messages := []llm.ChatMessage{
		{Role: "user", Content: prompt},
	}

	response, err := ia.client.ChatCompletion(messages, 0.2, 4000)
	if err != nil {
		log.Printf("[agent] LLM file selection error: %v", err)
		// Fallback: return top 10 files by match count
		return ia.fallbackFileSelection(candidates)
	}

	selectedFiles := ia.parseFileSelectionResponse(response)
	if len(selectedFiles) == 0 {
		// Fallback if LLM didn't return valid selections
		return ia.fallbackFileSelection(candidates)
	}

	return selectedFiles
}

// buildFileSelectionPrompt builds the prompt for LLM file selection.
// Mirrors Python's _build_element_selection_prompt.
func (ia *IterativeAgent) buildFileSelectionPrompt(query string, candidates []FileCandidate) string {
	var sb strings.Builder
	sb.WriteString("You are a code navigation assistant. Select only files from this repository that best address the query.\n")
	sb.WriteString(fmt.Sprintf("User Query: %q\n\n", query))
	sb.WriteString("**File Candidates**:\n")

	for i, c := range candidates {
		sb.WriteString(fmt.Sprintf("\n%d. %s", i+1, c.FilePath))
		if c.RepoName != "" {
			sb.WriteString(fmt.Sprintf("\n   Repo: %s", c.RepoName))
		}
		if c.MatchCount > 0 {
			sb.WriteString(fmt.Sprintf("\n   Matches: %d", c.MatchCount))
		}
	}

	sb.WriteString(`

**Your Task**: Select the MOST RELEVANT files to answer the query. Focus on actual source code files, NOT:
- Coverage reports or HTML files
- Configuration files (unless the query is about configuration)
- Test files (unless the query is about testing)
- Documentation/markdown files (unless the query is about docs)

**Response Format** (JSON only):
{
  "selected_elements": [
    {"file_path": "path/to/file.ts", "type": "file", "repo_name": "repo"}
  ]
}

**CRITICAL**:
- Respond with valid JSON only
- No markdown blocks, no comments
- Select 5-15 files maximum
- Prefer source files over generated files`)

	return sb.String()
}

// parseFileSelectionResponse parses the LLM response for file selections.
func (ia *IterativeAgent) parseFileSelectionResponse(response string) []string {
	jsonStr := extractJSON(response)
	if jsonStr == "" {
		return nil
	}

	var parsed map[string]any
	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		log.Printf("[agent] failed to parse file selection response: %v", err)
		return nil
	}

	selectedRaw, ok := parsed["selected_elements"]
	if !ok {
		return nil
	}

	elements, ok := selectedRaw.([]any)
	if !ok {
		return nil
	}

	var files []string
	for _, e := range elements {
		elemMap, ok := e.(map[string]any)
		if !ok {
			continue
		}
		filePath, ok := elemMap["file_path"].(string)
		if !ok || filePath == "" {
			continue
		}
		files = append(files, filePath)
	}

	return files
}

// fallbackFileSelection returns the top files sorted by match count when LLM selection fails.
func (ia *IterativeAgent) fallbackFileSelection(candidates []FileCandidate) []string {
	// Sort by match count descending (simple selection sort for small lists)
	sorted := make([]FileCandidate, len(candidates))
	copy(sorted, candidates)
	for i := 0; i < len(sorted)-1; i++ {
		maxIdx := i
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].MatchCount > sorted[maxIdx].MatchCount {
				maxIdx = j
			}
		}
		sorted[i], sorted[maxIdx] = sorted[maxIdx], sorted[i]
	}

	maxFiles := 10
	if len(sorted) < maxFiles {
		maxFiles = len(sorted)
	}

	files := make([]string, maxFiles)
	for i := 0; i < maxFiles; i++ {
		files[i] = sorted[i].FilePath
	}
	return files
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// deduplicateElements is a simple ID-based deduplication helper (backward-compat for tests).
func deduplicateElements(elements []types.CodeElement) []types.CodeElement {
	seen := make(map[string]bool)
	var unique []types.CodeElement
	for _, elem := range elements {
		if !seen[elem.ID] {
			seen[elem.ID] = true
			unique = append(unique, elem)
		}
	}
	return unique
}
