package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/duyhunghd6/fastcode-cli/internal/orchestrator"
)

// serveMCP starts a JSON-RPC server implementing the Model Context Protocol.
func serveMCP(cfg orchestrator.Config, port int) error {
	engine := orchestrator.NewEngine(cfg)
	mux := buildMCPMux(engine)

	addr := fmt.Sprintf(":%d", port)
	log.Printf("🚀 FastCode MCP server listening on http://localhost%s", addr)
	log.Printf("   MCP endpoint: http://localhost%s/mcp/", addr)
	return http.ListenAndServe(addr, mux)
}

// buildMCPMux creates the HTTP handler mux with all MCP endpoints.
func buildMCPMux(engine *orchestrator.Engine) *http.ServeMux {
	mux := http.NewServeMux()

	// MCP initialize
	mux.HandleFunc("/mcp/initialize", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"protocolVersion": "2024-11-05",
			"serverInfo": map[string]string{
				"name":    "fastcode-cli",
				"version": version,
			},
			"capabilities": map[string]any{
				"tools": map[string]bool{
					"listChanged": false,
				},
			},
		}
		writeJSON(w, resp)
	})

	// MCP tools/list
	mux.HandleFunc("/mcp/tools/list", func(w http.ResponseWriter, r *http.Request) {
		tools := []map[string]any{
			{
				"name":        "index_repository",
				"description": "Index a local code repository for querying",
				"inputSchema": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"path":  map[string]string{"type": "string", "description": "Path to the repository"},
						"force": map[string]string{"type": "boolean", "description": "Force re-indexing"},
					},
					"required": []string{"path"},
				},
			},
			{
				"name":        "query_codebase",
				"description": "Ask a question about an indexed codebase",
				"inputSchema": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"question": map[string]string{"type": "string", "description": "The question to ask"},
						"repo":     map[string]string{"type": "string", "description": "Repository path (optional if already indexed)"},
					},
					"required": []string{"question"},
				},
			},
			{
				"name":        "search_code",
				"description": "Search for code elements matching a query",
				"inputSchema": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"query": map[string]string{"type": "string", "description": "Search query"},
						"top_k": map[string]string{"type": "integer", "description": "Number of results (default: 10)"},
					},
					"required": []string{"query"},
				},
			},
		}
		writeJSON(w, map[string]any{"tools": tools})
	})

	// MCP tools/call
	mux.HandleFunc("/mcp/tools/call", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Name   string         `json:"name"`
			Params map[string]any `json:"arguments"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, "Invalid request body", 400)
			return
		}

		switch req.Name {
		case "index_repository":
			path, _ := req.Params["path"].(string)
			force, _ := req.Params["force"].(bool)
			if path == "" {
				writeError(w, "path is required", 400)
				return
			}
			result, err := engine.Index(path, force)
			if err != nil {
				writeError(w, err.Error(), 500)
				return
			}
			writeToolResult(w, result)

		case "query_codebase":
			question, _ := req.Params["question"].(string)
			repo, _ := req.Params["repo"].(string)
			if question == "" {
				writeError(w, "question is required", 400)
				return
			}
			if repo != "" {
				if _, err := engine.Index(repo, false); err != nil {
					writeError(w, err.Error(), 500)
					return
				}
			}
			result, err := engine.Query(question)
			if err != nil {
				writeError(w, err.Error(), 500)
				return
			}
			writeToolResult(w, result)

		default:
			writeError(w, fmt.Sprintf("Unknown tool: %s", req.Name), 404)
		}
	})

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]string{"status": "ok", "version": version})
	})

	return mux
}

func writeJSON(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{"message": msg},
	})
}

func writeToolResult(w http.ResponseWriter, data any) {
	content, _ := json.Marshal(data)
	writeJSON(w, map[string]any{
		"content": []map[string]any{
			{"type": "text", "text": string(content)},
		},
	})
}
