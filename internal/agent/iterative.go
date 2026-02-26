package agent

import (
	"encoding/json"
	"fmt"
	"log"
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
}

// AgentConfig holds configuration for the iterative agent.
type AgentConfig struct {
	MaxRounds           int     // Maximum retrieval rounds (default: 5)
	ConfidenceThreshold int     // Stop when confidence >= this (default: 85)
	MaxTokenBudget      int     // Maximum tokens to consume (default: 50000)
	Temperature         float64 // LLM temperature (default: 0.1)
}

// DefaultAgentConfig returns sensible defaults.
func DefaultAgentConfig() AgentConfig {
	return AgentConfig{
		MaxRounds:           5,
		ConfidenceThreshold: 85,
		MaxTokenBudget:      50000,
		Temperature:         0.1,
	}
}

// RoundResult holds the output of a single agent round.
type RoundResult struct {
	Round      int                 `json:"round"`
	Confidence int                 `json:"confidence"`
	Reasoning  string              `json:"reasoning"`
	ToolCalls  []ToolCall          `json:"tool_calls,omitempty"`
	Elements   []types.CodeElement `json:"elements,omitempty"`
}

// ToolCall represents a tool the agent wants to invoke.
type ToolCall struct {
	Name string `json:"name"`
	Arg  string `json:"arg"`
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
func (ia *IterativeAgent) Retrieve(query string, pq *ProcessedQuery) (*RetrievalResult, error) {
	ia.gatheredElements = nil
	ia.totalTokensUsed = 0
	ia.rounds = 0

	// Adaptive parameters based on query complexity
	maxRounds := ia.config.MaxRounds
	if pq.Complexity < 30 {
		maxRounds = min(maxRounds, 2) // Simple queries need fewer rounds
	}

	var lastConfidence int
	var stopReason string

	for round := 1; round <= maxRounds; round++ {
		ia.rounds = round

		roundResult, err := ia.executeRound(query, pq, round)
		if err != nil {
			log.Printf("[agent] round %d error: %v", round, err)
			stopReason = "error"
			break
		}

		lastConfidence = roundResult.Confidence

		// Execute tool calls from this round
		if len(roundResult.ToolCalls) > 0 {
			for _, tc := range roundResult.ToolCalls {
				result, err := ia.toolExecutor.Execute(tc.Name, tc.Arg)
				if err != nil {
					log.Printf("[agent] tool %s error: %v", tc.Name, err)
					continue
				}
				ia.gatheredElements = append(ia.gatheredElements, result.Elements...)
			}
		}

		// Check stopping conditions
		if roundResult.Confidence >= ia.config.ConfidenceThreshold {
			stopReason = "confidence_reached"
			break
		}
		if ia.totalTokensUsed >= ia.config.MaxTokenBudget {
			stopReason = "budget_exhausted"
			break
		}
		if len(roundResult.ToolCalls) == 0 {
			stopReason = "no_more_actions"
			break
		}
	}

	if stopReason == "" {
		stopReason = "max_rounds"
	}

	// Deduplicate elements
	elements := deduplicateElements(ia.gatheredElements)

	return &RetrievalResult{
		Elements:   elements,
		Rounds:     ia.rounds,
		Confidence: lastConfidence,
		StopReason: stopReason,
		Metadata: map[string]any{
			"query_complexity": pq.Complexity,
			"query_type":       pq.QueryType,
			"tokens_used":      ia.totalTokensUsed,
		},
	}, nil
}

// executeRound runs a single round of the iterative retrieval.
func (ia *IterativeAgent) executeRound(query string, pq *ProcessedQuery, round int) (*RoundResult, error) {
	prompt := ia.buildRoundPrompt(query, pq, round)

	response, err := ia.client.ChatCompletion([]llm.ChatMessage{
		{Role: "system", Content: systemPrompt()},
		{Role: "user", Content: prompt},
	}, ia.config.Temperature, 2000)
	if err != nil {
		return nil, fmt.Errorf("LLM call: %w", err)
	}

	return ia.parseRoundResponse(response, round)
}

func (ia *IterativeAgent) buildRoundPrompt(query string, pq *ProcessedQuery, round int) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## Query\n%s\n\n", query))
	sb.WriteString(fmt.Sprintf("## Query Analysis\n- Type: %s\n- Complexity: %d/100\n- Keywords: %s\n\n",
		pq.QueryType, pq.Complexity, strings.Join(pq.Keywords, ", ")))
	sb.WriteString(fmt.Sprintf("## Round %d of %d\n\n", round, ia.config.MaxRounds))

	if len(ia.gatheredElements) > 0 {
		sb.WriteString(fmt.Sprintf("## Context Gathered So Far (%d elements)\n", len(ia.gatheredElements)))
		for i, elem := range ia.gatheredElements {
			if i >= 20 { // Limit context window
				sb.WriteString(fmt.Sprintf("... and %d more elements\n", len(ia.gatheredElements)-20))
				break
			}
			sb.WriteString(fmt.Sprintf("- [%s] %s (%s) L%d-%d\n",
				elem.Type, elem.Name, elem.RelativePath, elem.StartLine, elem.EndLine))
			if elem.Signature != "" {
				sb.WriteString(fmt.Sprintf("  Signature: %s\n", elem.Signature))
			}
		}
		sb.WriteString("\n")
	}

	sb.WriteString("## Available Tools\n")
	for _, tool := range AvailableTools() {
		sb.WriteString(fmt.Sprintf("- `%s`: %s\n", tool.Name, tool.Description))
	}

	sb.WriteString("\n## Your Task\nAssess your confidence (0-100) that you have enough context to answer the query.\n")
	sb.WriteString("If confidence < 85, specify tool calls to gather more information.\n")
	sb.WriteString("\nRespond in JSON:\n```json\n{\"confidence\": <0-100>, \"reasoning\": \"...\", \"tool_calls\": [{\"name\": \"...\", \"arg\": \"...\"}]}\n```\n")

	return sb.String()
}

func (ia *IterativeAgent) parseRoundResponse(response string, round int) (*RoundResult, error) {
	result := &RoundResult{Round: round}

	// Try to extract JSON from the response
	jsonStr := extractJSON(response)
	if jsonStr == "" {
		// Fallback: treat as high-confidence with no tool calls
		result.Confidence = 90
		result.Reasoning = response
		return result, nil
	}

	var parsed struct {
		Confidence int        `json:"confidence"`
		Reasoning  string     `json:"reasoning"`
		ToolCalls  []ToolCall `json:"tool_calls"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		result.Confidence = 90
		result.Reasoning = response
		return result, nil
	}

	result.Confidence = parsed.Confidence
	result.Reasoning = parsed.Reasoning
	result.ToolCalls = parsed.ToolCalls
	return result, nil
}

func systemPrompt() string {
	return `You are a code retrieval agent. Your job is to assess whether you have enough context 
to answer a user's question about a codebase, and if not, to use tools to gather more information.

Be efficient with tokens — prefer skim_file over browse_file when possible.
Focus on the most relevant files first based on the query keywords.
Stop gathering context once you are confident enough to answer (confidence >= 85).

Always respond with valid JSON containing: confidence (0-100), reasoning, and tool_calls.`
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

func deduplicateElements(elements []types.CodeElement) []types.CodeElement {
	seen := make(map[string]bool)
	var result []types.CodeElement
	for _, elem := range elements {
		if !seen[elem.ID] {
			seen[elem.ID] = true
			result = append(result, elem)
		}
	}
	return result
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
