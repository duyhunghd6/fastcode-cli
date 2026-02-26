package parser

import (
	"strings"

	"github.com/duyhunghd6/fastcode-cli/internal/types"
	sitter "github.com/smacker/go-tree-sitter"
)

// parseGo extracts functions, methods, structs, interfaces, and imports from Go source.
func parseGo(root *sitter.Node, code []byte, result *types.FileParseResult) {
	for i := 0; i < int(root.ChildCount()); i++ {
		child := root.Child(i)
		switch child.Type() {
		case "package_clause":
			// Extract package comment (module docstring)
			if i == 0 || (i > 0 && root.Child(0).Type() == "comment") {
				result.ModuleDocstring = extractGoLeadingComment(root, code, int(child.StartPoint().Row))
			}

		case "import_declaration":
			result.Imports = append(result.Imports, extractGoImports(child, code)...)

		case "function_declaration":
			result.Functions = append(result.Functions, extractGoFunction(child, code, ""))

		case "method_declaration":
			fn := extractGoMethod(child, code)
			result.Functions = append(result.Functions, fn)

		case "type_declaration":
			for j := 0; j < int(child.ChildCount()); j++ {
				spec := child.Child(j)
				if spec.Type() == "type_spec" {
					ci := extractGoTypeSpec(spec, code)
					if ci != nil {
						result.Classes = append(result.Classes, *ci)
					}
				}
			}

		case "comment":
			// standalone comments, handled contextually
		}
	}
}

func extractGoImports(node *sitter.Node, code []byte) []types.ImportInfo {
	var imports []types.ImportInfo
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "import_spec" {
			imp := types.ImportInfo{
				Line: int(child.StartPoint().Row) + 1,
			}
			for j := 0; j < int(child.ChildCount()); j++ {
				c := child.Child(j)
				switch c.Type() {
				case "package_identifier", "dot":
					imp.Alias = c.Content(code)
				case "interpreted_string_literal":
					imp.Module = strings.Trim(c.Content(code), `"`)
				}
			}
			imports = append(imports, imp)
		} else if child.Type() == "import_spec_list" {
			for j := 0; j < int(child.ChildCount()); j++ {
				spec := child.Child(j)
				if spec.Type() == "import_spec" {
					imp := types.ImportInfo{
						Line: int(spec.StartPoint().Row) + 1,
					}
					for k := 0; k < int(spec.ChildCount()); k++ {
						c := spec.Child(k)
						switch c.Type() {
						case "package_identifier", "dot":
							imp.Alias = c.Content(code)
						case "interpreted_string_literal":
							imp.Module = strings.Trim(c.Content(code), `"`)
						}
					}
					imports = append(imports, imp)
				}
			}
		}
	}
	return imports
}

func extractGoFunction(node *sitter.Node, code []byte, className string) types.FunctionInfo {
	fn := types.FunctionInfo{
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		ClassName: className,
		IsMethod:  className != "",
	}
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		switch child.Type() {
		case "identifier":
			fn.Name = child.Content(code)
		case "parameter_list":
			fn.Parameters = extractGoParams(child, code)
		case "type_identifier", "pointer_type", "qualified_type", "slice_type", "map_type", "array_type":
			fn.ReturnType = child.Content(code)
		}
	}
	fn.Docstring = extractGoLeadingComment(node.Parent(), code, int(node.StartPoint().Row))
	return fn
}

func extractGoMethod(node *sitter.Node, code []byte) types.FunctionInfo {
	fn := types.FunctionInfo{
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		IsMethod:  true,
	}
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		switch child.Type() {
		case "parameter_list":
			if fn.Receiver == "" {
				// First parameter_list is the receiver
				fn.Receiver = child.Content(code)
				fn.ClassName = extractReceiverType(child, code)
			} else {
				fn.Parameters = extractGoParams(child, code)
			}
		case "field_identifier":
			fn.Name = child.Content(code)
		case "type_identifier", "pointer_type", "qualified_type", "slice_type", "map_type", "array_type":
			fn.ReturnType = child.Content(code)
		}
	}
	fn.Docstring = extractGoLeadingComment(node.Parent(), code, int(node.StartPoint().Row))
	return fn
}

func extractGoTypeSpec(node *sitter.Node, code []byte) *types.ClassInfo {
	ci := &types.ClassInfo{
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
	}
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		switch child.Type() {
		case "type_identifier":
			ci.Name = child.Content(code)
		case "struct_type":
			ci.Kind = "struct"
			ci.Bases = extractGoEmbeddedTypes(child, code)
		case "interface_type":
			ci.Kind = "interface"
			ci.Methods = extractGoInterfaceMethods(child, code, ci.Name)
		}
	}
	if ci.Name == "" {
		return nil
	}
	ci.Docstring = extractGoLeadingComment(node.Parent(), code, int(node.StartPoint().Row))
	return ci
}

func extractGoParams(node *sitter.Node, code []byte) []string {
	var params []string
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "parameter_declaration" {
			params = append(params, child.Content(code))
		}
	}
	return params
}

func extractReceiverType(node *sitter.Node, code []byte) string {
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "parameter_declaration" {
			for j := 0; j < int(child.ChildCount()); j++ {
				c := child.Child(j)
				switch c.Type() {
				case "type_identifier":
					return c.Content(code)
				case "pointer_type":
					for k := 0; k < int(c.ChildCount()); k++ {
						if c.Child(k).Type() == "type_identifier" {
							return c.Child(k).Content(code)
						}
					}
				}
			}
		}
	}
	return ""
}

func extractGoEmbeddedTypes(node *sitter.Node, code []byte) []string {
	var bases []string
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "field_declaration_list" {
			for j := 0; j < int(child.ChildCount()); j++ {
				field := child.Child(j)
				if field.Type() == "field_declaration" && field.ChildCount() == 1 {
					// Embedded type
					bases = append(bases, field.Child(0).Content(code))
				}
			}
		}
	}
	return bases
}

func extractGoInterfaceMethods(node *sitter.Node, code []byte, interfaceName string) []types.FunctionInfo {
	var methods []types.FunctionInfo
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "method_spec" {
			fn := types.FunctionInfo{
				StartLine: int(child.StartPoint().Row) + 1,
				EndLine:   int(child.EndPoint().Row) + 1,
				IsMethod:  true,
				ClassName: interfaceName,
			}
			for j := 0; j < int(child.ChildCount()); j++ {
				c := child.Child(j)
				switch c.Type() {
				case "field_identifier":
					fn.Name = c.Content(code)
				case "parameter_list":
					fn.Parameters = extractGoParams(c, code)
				}
			}
			methods = append(methods, fn)
		}
	}
	return methods
}

func extractGoLeadingComment(parent *sitter.Node, code []byte, targetRow int) string {
	if parent == nil {
		return ""
	}
	var comments []string
	for i := 0; i < int(parent.ChildCount()); i++ {
		child := parent.Child(i)
		if child.Type() == "comment" {
			endRow := int(child.EndPoint().Row)
			if endRow == targetRow-1 || endRow == targetRow-2 {
				text := child.Content(code)
				text = strings.TrimPrefix(text, "//")
				text = strings.TrimPrefix(text, " ")
				comments = append(comments, text)
			}
		}
	}
	if len(comments) == 0 {
		return ""
	}
	return strings.Join(comments, "\n")
}
