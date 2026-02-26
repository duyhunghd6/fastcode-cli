package agent

import (
	"fmt"
	"strings"

	"github.com/duyhunghd6/fastcode-cli/internal/llm"
	"github.com/duyhunghd6/fastcode-cli/internal/types"
)

// AnswerGenerator uses gathered context and an LLM to generate answers.
type AnswerGenerator struct {
	client *llm.Client
}

// NewAnswerGenerator creates a new answer generator.
func NewAnswerGenerator(client *llm.Client) *AnswerGenerator {
	return &AnswerGenerator{client: client}
}

// GenerateAnswer produces a natural-language answer given the query and retrieved context.
func (ag *AnswerGenerator) GenerateAnswer(query string, pq *ProcessedQuery, elements []types.CodeElement) (string, error) {
	prompt := ag.buildPrompt(query, pq, elements)

	answer, err := ag.client.ChatCompletion([]llm.ChatMessage{
		{Role: "system", Content: answerSystemPrompt()},
		{Role: "user", Content: prompt},
	}, 0.3, 4000)
	if err != nil {
		return "", fmt.Errorf("generate answer: %w", err)
	}

	return answer, nil
}

func (ag *AnswerGenerator) buildPrompt(query string, pq *ProcessedQuery, elements []types.CodeElement) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("## User Question\n%s\n\n", query))
	sb.WriteString(fmt.Sprintf("## Query Type: %s | Complexity: %d/100\n\n", pq.QueryType, pq.Complexity))

	sb.WriteString(fmt.Sprintf("## Retrieved Code Context (%d elements)\n\n", len(elements)))

	for i, elem := range elements {
		if i >= 15 { // Limit context to avoid token overflow
			sb.WriteString(fmt.Sprintf("\n... and %d more elements (omitted for brevity)\n", len(elements)-15))
			break
		}

		sb.WriteString(fmt.Sprintf("### [%s] %s\n", elem.Type, elem.Name))
		sb.WriteString(fmt.Sprintf("**File:** `%s` (L%d-%d) | **Language:** %s\n",
			elem.RelativePath, elem.StartLine, elem.EndLine, elem.Language))

		if elem.Signature != "" {
			sb.WriteString(fmt.Sprintf("**Signature:** `%s`\n", elem.Signature))
		}
		if elem.Docstring != "" {
			sb.WriteString(fmt.Sprintf("**Docstring:** %s\n", truncateStr(elem.Docstring, 200)))
		}
		if elem.Code != "" {
			code := truncateStr(elem.Code, 1000)
			sb.WriteString(fmt.Sprintf("```%s\n%s\n```\n", elem.Language, code))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func answerSystemPrompt() string {
	return `You are an expert code analyst. Answer the user's question about a codebase using 
ONLY the provided code context. Be specific and reference actual code elements, file paths, 
and line numbers when possible.

Structure your answer clearly:
1. Direct answer to the question
2. Key code references with file paths
3. Important details or caveats

If the context is insufficient, say so clearly and suggest what additional information would help.`
}

func truncateStr(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
