package parser

import (
	"strings"

	"github.com/duyhunghd6/fastcode-cli/internal/types"
	sitter "github.com/smacker/go-tree-sitter"
)

// parsePython extracts classes, functions, and imports from Python source.
func parsePython(root *sitter.Node, code []byte, result *types.FileParseResult) {
	// Extract module docstring
	if root.ChildCount() > 0 {
		first := root.Child(0)
		if first.Type() == "expression_statement" && first.ChildCount() > 0 {
			expr := first.Child(0)
			if expr.Type() == "string" {
				result.ModuleDocstring = cleanPythonDocstring(expr.Content(code))
			}
		}
	}

	for i := 0; i < int(root.ChildCount()); i++ {
		child := root.Child(i)
		switch child.Type() {
		case "import_statement":
			result.Imports = append(result.Imports, extractPythonImport(child, code))
		case "import_from_statement":
			result.Imports = append(result.Imports, extractPythonFromImport(child, code))
		case "class_definition":
			result.Classes = append(result.Classes, extractPythonClass(child, code))
		case "function_definition":
			fn := extractPythonFunction(child, code, "")
			if fn.Name != "" {
				result.Functions = append(result.Functions, fn)
			}
		case "decorated_definition":
			// A decorated_definition can wrap either a function or a class
			for j := 0; j < int(child.ChildCount()); j++ {
				inner := child.Child(j)
				if inner.Type() == "class_definition" {
					cls := extractPythonClass(child, code)
					result.Classes = append(result.Classes, cls)
					break
				} else if inner.Type() == "function_definition" {
					fn := extractPythonFunction(child, code, "")
					if fn.Name != "" {
						result.Functions = append(result.Functions, fn)
					}
					break
				}
			}
		}
	}
}

func extractPythonImport(node *sitter.Node, code []byte) types.ImportInfo {
	imp := types.ImportInfo{
		Line: int(node.StartPoint().Row) + 1,
	}
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "dotted_name" {
			imp.Module = child.Content(code)
		} else if child.Type() == "aliased_import" {
			for j := 0; j < int(child.ChildCount()); j++ {
				c := child.Child(j)
				if c.Type() == "dotted_name" {
					imp.Module = c.Content(code)
				} else if c.Type() == "identifier" {
					imp.Alias = c.Content(code)
				}
			}
		}
	}
	return imp
}

func extractPythonFromImport(node *sitter.Node, code []byte) types.ImportInfo {
	imp := types.ImportInfo{
		Line:   int(node.StartPoint().Row) + 1,
		IsFrom: true,
	}
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		switch child.Type() {
		case "dotted_name":
			imp.Module = child.Content(code)
		case "relative_import":
			imp.Module = child.Content(code)
			imp.Level = strings.Count(child.Content(code), ".")
		case "import_prefix":
			imp.Level = len(child.Content(code))
		case "identifier":
			imp.Names = append(imp.Names, child.Content(code))
		case "aliased_import":
			for j := 0; j < int(child.ChildCount()); j++ {
				c := child.Child(j)
				if c.Type() == "dotted_name" || c.Type() == "identifier" {
					imp.Names = append(imp.Names, c.Content(code))
				}
			}
		}
	}
	return imp
}

func extractPythonClass(node *sitter.Node, code []byte) types.ClassInfo {
	ci := types.ClassInfo{
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		Kind:      "class",
	}

	// Handle decorated definitions wrapping a class
	actual := node
	if node.Type() == "decorated_definition" {
		for i := 0; i < int(node.ChildCount()); i++ {
			child := node.Child(i)
			if child.Type() == "decorator" {
				ci.Decorators = append(ci.Decorators, child.Content(code))
			} else if child.Type() == "class_definition" {
				actual = child
			}
		}
	}

	for i := 0; i < int(actual.ChildCount()); i++ {
		child := actual.Child(i)
		switch child.Type() {
		case "identifier":
			ci.Name = child.Content(code)
		case "argument_list":
			ci.Bases = extractPythonBases(child, code)
		case "block":
			ci.Docstring = extractPythonBlockDocstring(child, code)
			ci.Methods = extractPythonMethods(child, code, ci.Name)
		case "decorator":
			ci.Decorators = append(ci.Decorators, child.Content(code))
		}
	}
	return ci
}

func extractPythonFunction(node *sitter.Node, code []byte, className string) types.FunctionInfo {
	fn := types.FunctionInfo{
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		ClassName: className,
		IsMethod:  className != "",
	}

	// Handle decorated definitions
	actual := node
	if node.Type() == "decorated_definition" {
		for i := 0; i < int(node.ChildCount()); i++ {
			child := node.Child(i)
			if child.Type() == "decorator" {
				fn.Decorators = append(fn.Decorators, child.Content(code))
			} else if child.Type() == "function_definition" {
				actual = child
			}
		}
	}

	for i := 0; i < int(actual.ChildCount()); i++ {
		child := actual.Child(i)
		switch child.Type() {
		case "identifier":
			fn.Name = child.Content(code)
		case "parameters":
			fn.Parameters = extractPythonParams(child, code)
		case "type":
			fn.ReturnType = child.Content(code)
		case "block":
			fn.Docstring = extractPythonBlockDocstring(child, code)
		}
	}

	// Check if async
	text := actual.Content(code)
	if strings.HasPrefix(text, "async ") {
		fn.IsAsync = true
	}

	return fn
}

func extractPythonBases(node *sitter.Node, code []byte) []string {
	var bases []string
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "identifier" || child.Type() == "attribute" {
			bases = append(bases, child.Content(code))
		}
	}
	return bases
}

func extractPythonParams(node *sitter.Node, code []byte) []string {
	var params []string
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		switch child.Type() {
		case "identifier":
			params = append(params, child.Content(code))
		case "typed_parameter", "default_parameter", "typed_default_parameter",
			"list_splat_pattern", "dictionary_splat_pattern":
			params = append(params, child.Content(code))
		}
	}
	return params
}

func extractPythonMethods(block *sitter.Node, code []byte, className string) []types.FunctionInfo {
	var methods []types.FunctionInfo
	for i := 0; i < int(block.ChildCount()); i++ {
		child := block.Child(i)
		if child.Type() == "function_definition" || child.Type() == "decorated_definition" {
			fn := extractPythonFunction(child, code, className)
			if fn.Name != "" {
				methods = append(methods, fn)
			}
		}
	}
	return methods
}

func extractPythonBlockDocstring(block *sitter.Node, code []byte) string {
	if block.ChildCount() == 0 {
		return ""
	}
	first := block.Child(0)
	if first.Type() == "expression_statement" && first.ChildCount() > 0 {
		expr := first.Child(0)
		if expr.Type() == "string" {
			return cleanPythonDocstring(expr.Content(code))
		}
	}
	return ""
}

func cleanPythonDocstring(raw string) string {
	// Remove triple quotes
	s := raw
	for _, q := range []string{`"""`, `'''`} {
		s = strings.TrimPrefix(s, q)
		s = strings.TrimSuffix(s, q)
	}
	return strings.TrimSpace(s)
}
