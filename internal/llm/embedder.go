package llm

import (
	"fmt"
	"log"
	"strings"
)

// Embedder generates embedding vectors for code elements via an LLM API.
type Embedder struct {
	client    *Client
	model     string
	batchSize int
}

// NewEmbedder creates a new embedder using the given client.
func NewEmbedder(client *Client, embeddingModel string, batchSize int) *Embedder {
	if embeddingModel == "" {
		embeddingModel = "text-embedding-3-small"
	}
	if batchSize <= 0 {
		batchSize = 32
	}
	return &Embedder{
		client:    client,
		model:     embeddingModel,
		batchSize: batchSize,
	}
}

// EmbedTexts generates embeddings for a list of texts, batching as needed.
func (e *Embedder) EmbedTexts(texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, nil
	}

	allEmbeddings := make([][]float32, len(texts))

	for start := 0; start < len(texts); start += e.batchSize {
		end := start + e.batchSize
		if end > len(texts) {
			end = len(texts)
		}
		batch := texts[start:end]

		embeddings, err := e.client.Embed(batch, e.model)
		if err != nil {
			return nil, fmt.Errorf("embed batch [%d:%d]: %w", start, end, err)
		}

		for i, emb := range embeddings {
			allEmbeddings[start+i] = emb
		}

		if end < len(texts) {
			log.Printf("[embedder] embedded %d/%d texts", end, len(texts))
		}
	}

	return allEmbeddings, nil
}

// EmbedText generates an embedding for a single text.
func (e *Embedder) EmbedText(text string) ([]float32, error) {
	results, err := e.EmbedTexts([]string{text})
	if err != nil {
		return nil, err
	}
	if len(results) == 0 || results[0] == nil {
		return nil, fmt.Errorf("no embedding returned")
	}
	return results[0], nil
}

// BuildSearchText creates a searchable text representation for a code element.
func BuildSearchText(name, docstring, signature, code string) string {
	var parts []string
	if name != "" {
		parts = append(parts, name)
	}
	if docstring != "" {
		parts = append(parts, docstring)
	}
	if signature != "" {
		parts = append(parts, signature)
	}
	// Truncate code to avoid exceeding token limits
	if code != "" {
		if len(code) > 2000 {
			code = code[:2000]
		}
		parts = append(parts, code)
	}
	return strings.Join(parts, "\n")
}
