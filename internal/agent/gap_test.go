package agent

import (
	"testing"

	"github.com/duyhunghd6/fastcode-cli/internal/types"
)

// === parseRound1Response Tests ===

func TestParseRound1ResponseNoJSON(t *testing.T) {
	ia := &IterativeAgent{}
	result, err := ia.parseRound1Response("This has no JSON at all, just text.")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Confidence != 90 {
		t.Errorf("confidence = %d, want 90 (fallback)", result.Confidence)
	}
	if result.Reasoning != "This has no JSON at all, just text." {
		t.Errorf("reasoning should be the full response")
	}
}

func TestParseRound1ResponseInvalidJSONFallback(t *testing.T) {
	ia := &IterativeAgent{}
	result, err := ia.parseRound1Response(`{"confidence": "not_a_number"}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Confidence != 90 {
		t.Errorf("confidence = %d, want 90 (fallback)", result.Confidence)
	}
}

func TestParseRound1ResponseValidJSONWithToolCalls(t *testing.T) {
	ia := &IterativeAgent{}
	result, err := ia.parseRound1Response(`{"confidence": 75, "reasoning": "Need more", "tool_calls": [{"tool": "search_codebase", "parameters": {"search_term": "main"}}]}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Confidence != 75 {
		t.Errorf("confidence = %d, want 75", result.Confidence)
	}
	if len(result.ToolCalls) != 1 {
		t.Errorf("tool_calls = %d, want 1", len(result.ToolCalls))
	}
}

func TestParseRound1ResponseHighConfidence(t *testing.T) {
	ia := &IterativeAgent{}
	result, err := ia.parseRound1Response(`{"confidence": 95, "reasoning": "Fully answered"}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Confidence != 95 {
		t.Errorf("confidence = %d, want 95", result.Confidence)
	}
	if len(result.ToolCalls) != 0 {
		t.Error("no tool calls expected")
	}
}

func TestParseRoundNResponseKeepFiles(t *testing.T) {
	ia := &IterativeAgent{}
	result, err := ia.parseRoundNResponse(`{"confidence": 97, "reasoning": "sufficient", "keep_files": ["file1.go", "file2.go"]}`, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Confidence != 97 {
		t.Errorf("confidence = %d, want 97", result.Confidence)
	}
	if len(result.KeepFiles) != 2 {
		t.Errorf("keep_files = %d, want 2", len(result.KeepFiles))
	}
}

// === extractJSON Tests ===

func TestExtractJSONFromMarkdown(t *testing.T) {
	input := "Some text\n```json\n{\"key\": \"value\"}\n```\nMore text"
	result := extractJSON(input)
	if result == "" {
		t.Error("should extract JSON from markdown code block")
	}
}

func TestExtractJSONBareJSON(t *testing.T) {
	input := `{"confidence": 50}`
	result := extractJSON(input)
	if result == "" {
		t.Error("should extract bare JSON object")
	}
}

func TestExtractJSONNoJSON(t *testing.T) {
	input := "This is just plain text with no JSON"
	result := extractJSON(input)
	if result != "" {
		t.Errorf("expected empty, got %q", result)
	}
}

// === scoreComplexity edge cases ===

func TestScoreComplexitySimpleQuery(t *testing.T) {
	score := scoreComplexity("what is main?", []string{"main"})
	if score >= 60 {
		t.Errorf("simple query complexity = %d, want < 60", score)
	}
}

func TestScoreComplexityMediumQuery(t *testing.T) {
	keywords := []string{"authentication", "middleware", "database", "connection"}
	score := scoreComplexity("How does the authentication middleware connect to the database", keywords)
	if score <= 20 {
		t.Errorf("medium query complexity = %d, want > 20", score)
	}
}

func TestScoreComplexityComplex(t *testing.T) {
	keywords := []string{"authentication", "middleware", "interact", "database", "connection", "pool", "transactions", "service", "layers", "distributed", "architecture", "error", "recovery", "retry"}
	score := scoreComplexity("How does the authentication middleware interact with the database connection pool and how are transactions handled across multiple service layers in the distributed architecture including error recovery and retry patterns?", keywords)
	if score <= 40 {
		t.Errorf("complex query complexity = %d, want > 40", score)
	}
}

func TestScoreComplexityWithTechIndicator(t *testing.T) {
	score := scoreComplexity("explain the architecture of this system", []string{"architecture", "system"})
	// Should get tech indicator bonus (+15)
	if score < 25 {
		t.Errorf("tech query complexity = %d, want >= 25", score)
	}
}

func TestScoreComplexityMultiIndicator(t *testing.T) {
	score := scoreComplexity("compare the two modules and explain their relationship", []string{"compare", "modules", "relationship"})
	// Should get multi-concept indicator bonus (+10)
	if score < 20 {
		t.Errorf("multi-concept complexity = %d, want >= 20", score)
	}
}

func TestScoreComplexityCap100(t *testing.T) {
	// Very complex query with many indicators to push past 100
	keywords := make([]string, 10)
	for i := range keywords {
		keywords[i] = "word"
	}
	score := scoreComplexity("How does the architecture interact between the concurrent threads and the async pipeline algorithm design pattern with dependency injection?", keywords)
	if score > 100 {
		t.Errorf("complexity = %d, should be capped at 100", score)
	}
}

// === deduplicateElements Tests ===

func TestDeduplicateElementsWithDupes(t *testing.T) {
	elements := []types.CodeElement{
		{ID: "e1", Name: "main"},
		{ID: "e2", Name: "helper"},
		{ID: "e1", Name: "main"}, // duplicate
	}
	result := deduplicateElements(elements)
	if len(result) != 2 {
		t.Errorf("expected 2 unique, got %d", len(result))
	}
}

func TestDeduplicateElementsNoDupes(t *testing.T) {
	elements := []types.CodeElement{
		{ID: "e1", Name: "main"},
		{ID: "e2", Name: "helper"},
	}
	result := deduplicateElements(elements)
	if len(result) != 2 {
		t.Errorf("expected 2, got %d", len(result))
	}
}

func TestDeduplicateElementsEmpty(t *testing.T) {
	result := deduplicateElements(nil)
	if len(result) != 0 {
		t.Errorf("expected 0, got %d", len(result))
	}
}

// === classifyQuery Tests ===

func TestClassifyQueryLocateType(t *testing.T) {
	result := classifyQuery("where is the main function defined?")
	if result != "locate" {
		t.Errorf("classify = %q, want locate", result)
	}
}

func TestClassifyQueryDebugType(t *testing.T) {
	result := classifyQuery("there is a bug in the parser")
	if result != "debug" {
		t.Errorf("classify = %q, want debug", result)
	}
}

func TestClassifyQueryHowtoType(t *testing.T) {
	result := classifyQuery("how to add a new endpoint")
	if result != "howto" {
		t.Errorf("classify = %q, want howto", result)
	}
}

func TestClassifyQueryOverviewType(t *testing.T) {
	result := classifyQuery("give me an overview of the codebase")
	if result != "overview" {
		t.Errorf("classify = %q, want overview", result)
	}
}

func TestClassifyQueryUnderstandType(t *testing.T) {
	result := classifyQuery("explain the config struct")
	if result != "understand" {
		t.Errorf("classify = %q, want understand", result)
	}
}

// === Min/Max helper ===

func TestMinHelperFunc(t *testing.T) {
	if min(3, 5) != 3 {
		t.Error("min(3,5) should be 3")
	}
	if min(7, 2) != 2 {
		t.Error("min(7,2) should be 2")
	}
	if min(4, 4) != 4 {
		t.Error("min(4,4) should be 4")
	}
}

func TestMaxHelperFunc(t *testing.T) {
	if max(3, 5) != 5 {
		t.Error("max(3,5) should be 5")
	}
	if max(7, 2) != 7 {
		t.Error("max(7,2) should be 7")
	}
}

// === filterElementsByKeepFiles ===

func TestFilterElementsByKeepFilesBasic(t *testing.T) {
	ia := &IterativeAgent{}
	elements := []types.CodeElement{
		{ID: "e1", RelativePath: "src/audio.ts", RepoName: "repo"},
		{ID: "e2", RelativePath: "src/other.ts", RepoName: "repo"},
		{ID: "e3", RelativePath: "src/engine.ts", RepoName: "repo"},
	}
	keepFiles := []string{"src/audio.ts", "src/engine.ts"}
	result := ia.filterElementsByKeepFiles(elements, keepFiles)
	if len(result) != 2 {
		t.Errorf("expected 2 kept, got %d", len(result))
	}
}

func TestFilterElementsByKeepFilesEmpty(t *testing.T) {
	ia := &IterativeAgent{}
	elements := []types.CodeElement{
		{ID: "e1", RelativePath: "src/audio.ts"},
	}
	// Empty keep_files should return all elements
	result := ia.filterElementsByKeepFiles(elements, nil)
	if len(result) != 1 {
		t.Errorf("expected all 1, got %d", len(result))
	}
}

// === Adaptive parameters ===

func TestInitializeAdaptiveParamsSimple(t *testing.T) {
	ia := &IterativeAgent{config: DefaultAgentConfig()}
	ia.initializeAdaptiveParams(20) // simple query
	if ia.confidenceThreshold != 95 {
		t.Errorf("simple query threshold = %d, want 95", ia.confidenceThreshold)
	}
	if ia.adaptiveLineBudget > 8000 {
		t.Errorf("simple query budget = %d, want <= 8000", ia.adaptiveLineBudget)
	}
}

func TestInitializeAdaptiveParamsComplex(t *testing.T) {
	ia := &IterativeAgent{config: DefaultAgentConfig()}
	ia.initializeAdaptiveParams(85) // complex query
	if ia.confidenceThreshold > 92 {
		t.Errorf("complex query threshold = %d, want <= 92", ia.confidenceThreshold)
	}
	if ia.adaptiveLineBudget < 10000 {
		t.Errorf("complex query budget = %d, want >= 10000", ia.adaptiveLineBudget)
	}
}
