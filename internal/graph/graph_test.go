package graph

import (
	"testing"

	"github.com/duyhunghd6/fastcode-cli/internal/types"
)

func TestGraphAddEdge(t *testing.T) {
	g := NewGraph(DependencyGraph)
	g.AddEdge("a", "b")
	g.AddEdge("a", "c")
	g.AddEdge("b", "c")

	if got := g.EdgeCount(); got != 3 {
		t.Errorf("EdgeCount() = %d, want 3", got)
	}
	if got := g.NodeCount(); got != 3 {
		t.Errorf("NodeCount() = %d, want 3", got)
	}

	succ := g.Successors("a")
	if len(succ) != 2 {
		t.Errorf("Successors(a) = %d, want 2", len(succ))
	}

	pred := g.Predecessors("c")
	if len(pred) != 2 {
		t.Errorf("Predecessors(c) = %d, want 2", len(pred))
	}
}

func TestGraphNoDuplicateEdges(t *testing.T) {
	g := NewGraph(CallGraph)
	g.AddEdge("a", "b")
	g.AddEdge("a", "b") // duplicate
	if got := g.EdgeCount(); got != 1 {
		t.Errorf("EdgeCount() = %d after duplicate, want 1", got)
	}
}

func TestGraphNoSelfLoop(t *testing.T) {
	g := NewGraph(InheritanceGraph)
	g.AddEdge("a", "a")
	if got := g.EdgeCount(); got != 0 {
		t.Errorf("EdgeCount() = %d after self-loop, want 0", got)
	}
}

func TestGraphEmptySuccessors(t *testing.T) {
	g := NewGraph(DependencyGraph)
	succ := g.Successors("nonexistent")
	if len(succ) != 0 {
		t.Errorf("expected 0 successors for nonexistent node, got %d", len(succ))
	}
}

func TestGraphEmptyPredecessors(t *testing.T) {
	g := NewGraph(DependencyGraph)
	pred := g.Predecessors("nonexistent")
	if len(pred) != 0 {
		t.Errorf("expected 0 predecessors for nonexistent node, got %d", len(pred))
	}
}

func TestGraphEmptyNodeCount(t *testing.T) {
	g := NewGraph(DependencyGraph)
	if got := g.NodeCount(); got != 0 {
		t.Errorf("NodeCount() = %d for empty graph, want 0", got)
	}
}

func TestGraphEmptyEdgeCount(t *testing.T) {
	g := NewGraph(DependencyGraph)
	if got := g.EdgeCount(); got != 0 {
		t.Errorf("EdgeCount() = %d for empty graph, want 0", got)
	}
}

func TestNewCodeGraphs(t *testing.T) {
	cg := NewCodeGraphs()
	if cg == nil {
		t.Fatal("NewCodeGraphs returned nil")
	}
	if cg.Dependency == nil || cg.Inheritance == nil || cg.Call == nil {
		t.Error("all three graphs should be initialized")
	}
}

func TestGetRelatedElements(t *testing.T) {
	cg := NewCodeGraphs()
	cg.Dependency.AddEdge("file_a", "file_b")
	cg.Dependency.AddEdge("file_b", "file_c")
	cg.Dependency.AddEdge("file_c", "file_d")

	// 1 hop from file_a → should find file_b
	related1 := cg.GetRelatedElements("file_a", 1)
	if len(related1) != 1 {
		t.Errorf("1-hop related = %d, want 1", len(related1))
	}

	// 2 hops from file_a → should find file_b, file_c
	related2 := cg.GetRelatedElements("file_a", 2)
	if len(related2) != 2 {
		t.Errorf("2-hop related = %d, want 2", len(related2))
	}
}

func TestGetRelatedElementsMultipleGraphs(t *testing.T) {
	cg := NewCodeGraphs()
	cg.Dependency.AddEdge("a", "b")
	cg.Inheritance.AddEdge("a", "c")
	cg.Call.AddEdge("a", "d")

	related := cg.GetRelatedElements("a", 1)
	if len(related) != 3 {
		t.Errorf("expected 3 related across all graphs, got %d", len(related))
	}
}

func TestGetRelatedElementsZeroHops(t *testing.T) {
	cg := NewCodeGraphs()
	cg.Dependency.AddEdge("a", "b")

	related := cg.GetRelatedElements("a", 0)
	if len(related) != 0 {
		t.Errorf("0 hops should return 0 related, got %d", len(related))
	}
}

func TestGetRelatedElementsReverse(t *testing.T) {
	cg := NewCodeGraphs()
	cg.Dependency.AddEdge("a", "b")

	// From b, should find a via reverse edges
	related := cg.GetRelatedElements("b", 1)
	if len(related) != 1 {
		t.Errorf("reverse traversal: expected 1 related, got %d", len(related))
	}
}

func TestGenerateElementID(t *testing.T) {
	id1 := GenerateElementID("myrepo", "file", "internal/parser/parser.go")
	id2 := GenerateElementID("myrepo", "file", "internal/parser/parser.go")
	if id1 != id2 {
		t.Errorf("deterministic ID failed: %q != %q", id1, id2)
	}
	if id1 == "" {
		t.Error("ID should not be empty")
	}

	// Different inputs should produce different IDs
	id3 := GenerateElementID("myrepo", "function", "main")
	if id1 == id3 {
		t.Error("different inputs should produce different IDs")
	}
}

func TestBuildInheritanceGraph(t *testing.T) {
	cg := NewCodeGraphs()
	elements := []types.CodeElement{
		{ID: "cls_base", Type: "class", Name: "BaseHandler"},
		{ID: "cls_child", Type: "class", Name: "UserHandler", Metadata: map[string]any{
			"bases": []string{"BaseHandler"},
		}},
	}
	cg.BuildGraphs(elements)

	succ := cg.Inheritance.Successors("cls_child")
	if len(succ) != 1 || succ[0] != "cls_base" {
		t.Errorf("inheritance edge missing: got %v", succ)
	}
}

func TestBuildCallGraph(t *testing.T) {
	cg := NewCodeGraphs()
	elements := []types.CodeElement{
		{ID: "fn_main", Type: "function", Name: "main", Metadata: map[string]any{
			"calls": []string{"helper"},
		}},
		{ID: "fn_helper", Type: "function", Name: "helper"},
	}
	cg.BuildGraphs(elements)

	succ := cg.Call.Successors("fn_main")
	if len(succ) != 1 || succ[0] != "fn_helper" {
		t.Errorf("call edge missing: got %v", succ)
	}
}

func TestBuildDependencyGraph(t *testing.T) {
	cg := NewCodeGraphs()
	elements := []types.CodeElement{
		{
			ID: "file_a", Type: "file", RelativePath: "a.go",
			Metadata: map[string]any{
				"imports": []types.ImportInfo{
					{Module: "b"},
				},
			},
		},
		{
			ID: "file_b", Type: "file", RelativePath: "b.go",
		},
	}
	cg.BuildGraphs(elements)

	succ := cg.Dependency.Successors("file_a")
	if len(succ) != 1 || succ[0] != "file_b" {
		t.Errorf("dependency edge missing: got %v", succ)
	}
}

func TestStats(t *testing.T) {
	cg := NewCodeGraphs()
	cg.Dependency.AddEdge("a", "b")
	cg.Inheritance.AddEdge("c", "d")
	cg.Call.AddEdge("e", "f")

	stats := cg.Stats()
	if stats == nil {
		t.Fatal("Stats returned nil")
	}
	if _, ok := stats["dependency"]; !ok {
		t.Error("stats missing 'dependency'")
	}
	if _, ok := stats["inheritance"]; !ok {
		t.Error("stats missing 'inheritance'")
	}
	if _, ok := stats["call"]; !ok {
		t.Error("stats missing 'call'")
	}
}

func TestResolveImportModulePath(t *testing.T) {
	cg := NewCodeGraphs()
	elements := []types.CodeElement{
		{
			ID: "file_a", Type: "file", RelativePath: "internal/parser/parser.go",
			Metadata: map[string]any{
				"imports": []types.ImportInfo{
					{Module: "internal.util"},
				},
			},
		},
		{
			ID: "file_b", Type: "file", RelativePath: "internal/util/language.go",
		},
	}
	cg.BuildGraphs(elements)

	succ := cg.Dependency.Successors("file_a")
	if len(succ) != 1 {
		t.Errorf("module-style import resolution: expected 1 edge, got %d", len(succ))
	}
}

func TestResolveImportNoMatch(t *testing.T) {
	cg := NewCodeGraphs()
	elements := []types.CodeElement{
		{
			ID: "file_a", Type: "file", RelativePath: "a.go",
			Metadata: map[string]any{
				"imports": []types.ImportInfo{
					{Module: "external/pkg"},
				},
			},
		},
	}
	cg.BuildGraphs(elements)

	succ := cg.Dependency.Successors("file_a")
	if len(succ) != 0 {
		t.Errorf("unresolvable import should add 0 edges, got %d", len(succ))
	}
}

func TestResolveImportEmptyModule(t *testing.T) {
	cg := NewCodeGraphs()
	elements := []types.CodeElement{
		{
			ID: "file_a", Type: "file", RelativePath: "a.go",
			Metadata: map[string]any{
				"imports": []types.ImportInfo{
					{Module: ""},
				},
			},
		},
	}
	cg.BuildGraphs(elements)

	succ := cg.Dependency.Successors("file_a")
	if len(succ) != 0 {
		t.Errorf("empty module import should add 0 edges, got %d", len(succ))
	}
}

func TestBuildGraphsNoMetadata(t *testing.T) {
	cg := NewCodeGraphs()
	elements := []types.CodeElement{
		{ID: "file_a", Type: "file", RelativePath: "a.go"},
		{ID: "fn_main", Type: "function", Name: "main"},
		{ID: "cls_foo", Type: "class", Name: "Foo"},
	}
	cg.BuildGraphs(elements)
	// Should not panic, graphs should be empty
	if cg.Dependency.EdgeCount() != 0 {
		t.Errorf("expected 0 dependency edges, got %d", cg.Dependency.EdgeCount())
	}
	if cg.Call.EdgeCount() != 0 {
		t.Errorf("expected 0 call edges, got %d", cg.Call.EdgeCount())
	}
	if cg.Inheritance.EdgeCount() != 0 {
		t.Errorf("expected 0 inheritance edges, got %d", cg.Inheritance.EdgeCount())
	}
}
