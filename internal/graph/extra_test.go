package graph

import (
	"testing"

	"github.com/duyhunghd6/fastcode-cli/internal/types"
)

// === buildDependencyGraph: imports metadata as ImportInfo ===

func TestBuildDependencyGraphWithImports(t *testing.T) {
	cg := NewCodeGraphs()

	elements := []types.CodeElement{
		{
			ID: "file1", Type: "file", Name: "main.go", RelativePath: "main.go",
			Metadata: map[string]any{
				"imports": []types.ImportInfo{
					{Module: "utils", Line: 3},
					{Module: "external/pkg", Line: 4},
				},
			},
		},
		{
			ID: "file2", Type: "file", Name: "utils.go", RelativePath: "utils.go",
			Metadata: map[string]any{},
		},
	}

	cg.BuildGraphs(elements)

	// file1 should depend on file2 since "utils" matches "utils.go"
	deps := cg.Dependency.Successors("file1")
	if len(deps) == 0 {
		t.Error("expected dependency edge from file1 to file2")
	}
}

func TestBuildDependencyGraphNoImports(t *testing.T) {
	cg := NewCodeGraphs()

	elements := []types.CodeElement{
		{ID: "file1", Type: "file", Name: "main.go", RelativePath: "main.go", Metadata: map[string]any{}},
	}

	cg.BuildGraphs(elements)

	deps := cg.Dependency.Successors("file1")
	if len(deps) != 0 {
		t.Errorf("no imports, expected 0 deps, got %d", len(deps))
	}
}

func TestBuildDependencyGraphWrongImportType(t *testing.T) {
	cg := NewCodeGraphs()

	elements := []types.CodeElement{
		{
			ID: "file1", Type: "file", Name: "main.go", RelativePath: "main.go",
			Metadata: map[string]any{
				"imports": "not a slice", // Wrong type — should be skipped
			},
		},
	}

	cg.BuildGraphs(elements)

	deps := cg.Dependency.Successors("file1")
	if len(deps) != 0 {
		t.Error("wrong type should be skipped")
	}
}

// === buildInheritanceGraph: bases metadata ===

func TestBuildInheritanceGraphWithBases(t *testing.T) {
	cg := NewCodeGraphs()

	elements := []types.CodeElement{
		{ID: "cls1", Type: "class", Name: "Animal", Metadata: map[string]any{}},
		{
			ID: "cls2", Type: "class", Name: "Dog",
			Metadata: map[string]any{
				"bases": []string{"Animal"},
			},
		},
	}

	cg.BuildGraphs(elements)

	// Dog should inherit from Animal
	succ := cg.Inheritance.Successors("cls2")
	if len(succ) != 1 || succ[0] != "cls1" {
		t.Errorf("expected Dog → Animal, got %v", succ)
	}
}

func TestBuildInheritanceGraphNoBases(t *testing.T) {
	cg := NewCodeGraphs()

	elements := []types.CodeElement{
		{ID: "cls1", Type: "class", Name: "Base", Metadata: map[string]any{}},
	}

	cg.BuildGraphs(elements)

	succ := cg.Inheritance.Successors("cls1")
	if len(succ) != 0 {
		t.Errorf("expected 0, got %d", len(succ))
	}
}

func TestBuildInheritanceGraphWrongBasesType(t *testing.T) {
	cg := NewCodeGraphs()

	elements := []types.CodeElement{
		{
			ID: "cls1", Type: "class", Name: "Broken",
			Metadata: map[string]any{
				"bases": 42, // Wrong type — should be skipped
			},
		},
	}

	cg.BuildGraphs(elements)
	succ := cg.Inheritance.Successors("cls1")
	if len(succ) != 0 {
		t.Error("wrong type should be skipped")
	}
}

// === buildCallGraph: calls metadata ===

func TestBuildCallGraphWithCalls(t *testing.T) {
	cg := NewCodeGraphs()

	elements := []types.CodeElement{
		{
			ID: "fn1", Type: "function", Name: "main",
			Metadata: map[string]any{
				"calls": []string{"helper"},
			},
		},
		{ID: "fn2", Type: "function", Name: "helper", Metadata: map[string]any{}},
	}

	cg.BuildGraphs(elements)

	callees := cg.Call.Successors("fn1")
	if len(callees) != 1 || callees[0] != "fn2" {
		t.Errorf("expected main → helper, got %v", callees)
	}
}

func TestBuildCallGraphNoCalls(t *testing.T) {
	cg := NewCodeGraphs()

	elements := []types.CodeElement{
		{ID: "fn1", Type: "function", Name: "main", Metadata: map[string]any{}},
	}

	cg.BuildGraphs(elements)

	callees := cg.Call.Successors("fn1")
	if len(callees) != 0 {
		t.Errorf("expected 0, got %d", len(callees))
	}
}

func TestBuildCallGraphWrongCallsType(t *testing.T) {
	cg := NewCodeGraphs()

	elements := []types.CodeElement{
		{
			ID: "fn1", Type: "function", Name: "broken",
			Metadata: map[string]any{
				"calls": 123, // Wrong type — should be skipped
			},
		},
	}

	cg.BuildGraphs(elements)
	callees := cg.Call.Successors("fn1")
	if len(callees) != 0 {
		t.Error("wrong type should be skipped")
	}
}

func TestBuildCallGraphUnresolvedCallee(t *testing.T) {
	cg := NewCodeGraphs()

	elements := []types.CodeElement{
		{
			ID: "fn1", Type: "function", Name: "main",
			Metadata: map[string]any{
				"calls": []string{"nonexistent_function"},
			},
		},
	}

	cg.BuildGraphs(elements)

	callees := cg.Call.Successors("fn1")
	if len(callees) != 0 {
		t.Errorf("expected 0 for unresolved callee, got %d", len(callees))
	}
}

// === resolveImport: dot-path resolution ===

func TestResolveImportDotPath(t *testing.T) {
	cg := NewCodeGraphs()
	cg.fileByPath["services/auth.py"] = "auth_id"

	// Module-style with dots
	imp := types.ImportInfo{Module: "services.auth"}
	source := &types.CodeElement{ID: "src", RelativePath: "main.py"}

	result := cg.resolveImport(imp, source)
	if result != "auth_id" {
		t.Errorf("expected auth_id, got %q", result)
	}
}

func TestResolveImportEmpty(t *testing.T) {
	cg := NewCodeGraphs()

	imp := types.ImportInfo{Module: ""}
	source := &types.CodeElement{ID: "src"}

	result := cg.resolveImport(imp, source)
	if result != "" {
		t.Errorf("empty module should return empty, got %q", result)
	}
}

func TestResolveImportNoMatchUnrelated(t *testing.T) {
	cg := NewCodeGraphs()
	cg.fileByPath["main.go"] = "main_id"

	imp := types.ImportInfo{Module: "completely_unrelated"}
	source := &types.CodeElement{ID: "src"}

	result := cg.resolveImport(imp, source)
	if result != "" {
		t.Errorf("no match should return empty, got %q", result)
	}
}

// === Non-file elements skipped in dependency graph ===

func TestBuildDependencyGraphSkipsNonFileElements(t *testing.T) {
	cg := NewCodeGraphs()

	elements := []types.CodeElement{
		{ID: "fn1", Type: "function", Name: "main", Metadata: map[string]any{
			"imports": []types.ImportInfo{{Module: "utils"}},
		}},
	}

	cg.BuildGraphs(elements)

	deps := cg.Dependency.Successors("fn1")
	if len(deps) != 0 {
		t.Error("non-file elements should be skipped in dependency graph")
	}
}
