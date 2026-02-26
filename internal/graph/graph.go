package graph

import (
	"crypto/sha256"
	"fmt"
	"strings"

	"github.com/duyhunghd6/fastcode-cli/internal/types"
)

// GraphType represents the type of relationship graph.
type GraphType string

const (
	DependencyGraph  GraphType = "dependency"
	InheritanceGraph GraphType = "inheritance"
	CallGraph        GraphType = "call"
)

// Edge represents a directed edge in a code relationship graph.
type Edge struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Label  string `json:"label,omitempty"`
}

// Graph holds a single graph's adjacency list.
type Graph struct {
	Type    GraphType           `json:"type"`
	Forward map[string][]string // node → outgoing edges
	Reverse map[string][]string // node → incoming edges
}

// NewGraph creates a new empty graph.
func NewGraph(t GraphType) *Graph {
	return &Graph{
		Type:    t,
		Forward: make(map[string][]string),
		Reverse: make(map[string][]string),
	}
}

// AddEdge adds a directed edge from source to target.
func (g *Graph) AddEdge(source, target string) {
	if source == target {
		return
	}
	// Avoid duplicates
	for _, t := range g.Forward[source] {
		if t == target {
			return
		}
	}
	g.Forward[source] = append(g.Forward[source], target)
	g.Reverse[target] = append(g.Reverse[target], source)
}

// Successors returns all direct successors of a node.
func (g *Graph) Successors(nodeID string) []string {
	return g.Forward[nodeID]
}

// Predecessors returns all direct predecessors of a node.
func (g *Graph) Predecessors(nodeID string) []string {
	return g.Reverse[nodeID]
}

// NodeCount returns the number of unique nodes.
func (g *Graph) NodeCount() int {
	nodes := make(map[string]bool)
	for k, vs := range g.Forward {
		nodes[k] = true
		for _, v := range vs {
			nodes[v] = true
		}
	}
	return len(nodes)
}

// EdgeCount returns the total number of edges.
func (g *Graph) EdgeCount() int {
	count := 0
	for _, vs := range g.Forward {
		count += len(vs)
	}
	return count
}

// CodeGraphs holds all three relationship graphs.
type CodeGraphs struct {
	Dependency  *Graph
	Inheritance *Graph
	Call        *Graph

	// Lookup maps
	elementByID map[string]*types.CodeElement
	fileByPath  map[string]string // relativePath → elementID
}

// NewCodeGraphs creates a new set of code relationship graphs.
func NewCodeGraphs() *CodeGraphs {
	return &CodeGraphs{
		Dependency:  NewGraph(DependencyGraph),
		Inheritance: NewGraph(InheritanceGraph),
		Call:        NewGraph(CallGraph),
		elementByID: make(map[string]*types.CodeElement),
		fileByPath:  make(map[string]string),
	}
}

// BuildGraphs constructs all three graphs from the given code elements.
func (cg *CodeGraphs) BuildGraphs(elements []types.CodeElement) {
	// Build lookup maps
	for i := range elements {
		elem := &elements[i]
		cg.elementByID[elem.ID] = elem
		if elem.Type == "file" {
			cg.fileByPath[elem.RelativePath] = elem.ID
		}
	}

	// Build each graph
	cg.buildDependencyGraph(elements)
	cg.buildInheritanceGraph(elements)
	cg.buildCallGraph(elements)
}

// GetRelatedElements returns all elements within maxHops of the given element.
func (cg *CodeGraphs) GetRelatedElements(elementID string, maxHops int) []string {
	visited := make(map[string]bool)
	queue := []string{elementID}
	visited[elementID] = true

	for hop := 0; hop < maxHops && len(queue) > 0; hop++ {
		var next []string
		for _, id := range queue {
			// Check all three graphs
			for _, graph := range []*Graph{cg.Dependency, cg.Inheritance, cg.Call} {
				for _, neighbor := range graph.Successors(id) {
					if !visited[neighbor] {
						visited[neighbor] = true
						next = append(next, neighbor)
					}
				}
				for _, neighbor := range graph.Predecessors(id) {
					if !visited[neighbor] {
						visited[neighbor] = true
						next = append(next, neighbor)
					}
				}
			}
		}
		queue = next
	}

	// Collect all visited except the starting element
	var related []string
	for id := range visited {
		if id != elementID {
			related = append(related, id)
		}
	}
	return related
}

// Stats returns statistics about all graphs.
func (cg *CodeGraphs) Stats() map[string]any {
	return map[string]any{
		"dependency":  map[string]int{"nodes": cg.Dependency.NodeCount(), "edges": cg.Dependency.EdgeCount()},
		"inheritance": map[string]int{"nodes": cg.Inheritance.NodeCount(), "edges": cg.Inheritance.EdgeCount()},
		"call":        map[string]int{"nodes": cg.Call.NodeCount(), "edges": cg.Call.EdgeCount()},
	}
}

// --- Graph building logic ---

// buildDependencyGraph creates file-level dependency edges based on imports.
func (cg *CodeGraphs) buildDependencyGraph(elements []types.CodeElement) {
	for i := range elements {
		elem := &elements[i]
		if elem.Type != "file" {
			continue
		}

		// Get imports from metadata
		imports, ok := elem.Metadata["imports"]
		if !ok {
			continue
		}
		importList, ok := imports.([]types.ImportInfo)
		if !ok {
			continue
		}

		for _, imp := range importList {
			// Try to resolve the import to a file in the repo
			targetID := cg.resolveImport(imp, elem)
			if targetID != "" {
				cg.Dependency.AddEdge(elem.ID, targetID)
			}
		}
	}
}

// buildInheritanceGraph creates class inheritance edges.
func (cg *CodeGraphs) buildInheritanceGraph(elements []types.CodeElement) {
	// Build class name → ID map
	classMap := make(map[string]string) // "ClassName" → elementID
	for i := range elements {
		elem := &elements[i]
		if elem.Type == "class" {
			classMap[elem.Name] = elem.ID
		}
	}

	for i := range elements {
		elem := &elements[i]
		if elem.Type != "class" {
			continue
		}
		bases, ok := elem.Metadata["bases"]
		if !ok {
			continue
		}
		baseList, ok := bases.([]string)
		if !ok {
			continue
		}
		for _, base := range baseList {
			if targetID, found := classMap[base]; found {
				cg.Inheritance.AddEdge(elem.ID, targetID)
			}
		}
	}
}

// buildCallGraph creates function call edges.
func (cg *CodeGraphs) buildCallGraph(elements []types.CodeElement) {
	// Build function name → ID map
	funcMap := make(map[string]string)
	for i := range elements {
		elem := &elements[i]
		if elem.Type == "function" {
			funcMap[elem.Name] = elem.ID
		}
	}

	for i := range elements {
		elem := &elements[i]
		if elem.Type != "function" {
			continue
		}
		calls, ok := elem.Metadata["calls"]
		if !ok {
			continue
		}
		callList, ok := calls.([]string)
		if !ok {
			continue
		}
		for _, callee := range callList {
			if targetID, found := funcMap[callee]; found {
				cg.Call.AddEdge(elem.ID, targetID)
			}
		}
	}
}

// resolveImport tries to map an import to a file element ID.
func (cg *CodeGraphs) resolveImport(imp types.ImportInfo, source *types.CodeElement) string {
	module := imp.Module
	if module == "" {
		return ""
	}

	// Try direct path match (e.g., Go imports)
	for path, id := range cg.fileByPath {
		if strings.Contains(path, module) || strings.HasSuffix(path, module) {
			return id
		}
	}

	// Try module-style resolution (dots to slashes)
	modulePath := strings.ReplaceAll(module, ".", "/")
	for path, id := range cg.fileByPath {
		if strings.Contains(path, modulePath) {
			return id
		}
	}

	return ""
}

// GenerateElementID generates a deterministic ID for a code element.
func GenerateElementID(repoName, elemType string, parts ...string) string {
	h := sha256.New()
	for _, p := range parts {
		h.Write([]byte(p))
	}
	hash := fmt.Sprintf("%x", h.Sum(nil))[:12]
	return fmt.Sprintf("%s_%s_%s", repoName, elemType, hash)
}
