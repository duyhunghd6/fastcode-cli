package parser

import (
	"github.com/duyhunghd6/fastcode-cli/internal/types"
	sitter "github.com/smacker/go-tree-sitter"
)

// parseJava extracts classes, methods, and imports from Java source.
// Stub implementation — to be fully ported from Python parser.
func parseJava(root *sitter.Node, code []byte, result *types.FileParseResult) {
	visitGenericNode(root, code, result, "java")
}

// parseRust extracts structs, impl blocks, functions, and use statements from Rust source.
// Stub implementation — to be fully ported from Python parser.
func parseRust(root *sitter.Node, code []byte, result *types.FileParseResult) {
	visitGenericNode(root, code, result, "rust")
}

// parseC extracts structs, functions, and includes from C/C++ source.
// Stub implementation — to be fully ported from Python parser.
func parseC(root *sitter.Node, code []byte, result *types.FileParseResult) {
	visitGenericNode(root, code, result, "c")
}

// visitGenericNode is a basic recursive visitor that extracts functions via
// common tree-sitter node types. Used as a fallback for languages without
// dedicated parsers yet.
func visitGenericNode(node *sitter.Node, code []byte, result *types.FileParseResult, lang string) {
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		t := child.Type()

		switch {
		// Function-like declarations
		case t == "function_definition" || t == "function_item" || t == "method_declaration" ||
			t == "function_declaration":
			fn := types.FunctionInfo{
				StartLine: int(child.StartPoint().Row) + 1,
				EndLine:   int(child.EndPoint().Row) + 1,
			}
			// Try to find function name
			for j := 0; j < int(child.ChildCount()); j++ {
				c := child.Child(j)
				if c.Type() == "identifier" || c.Type() == "field_identifier" {
					fn.Name = c.Content(code)
					break
				}
			}
			if fn.Name != "" {
				result.Functions = append(result.Functions, fn)
			}

		// Class/struct-like declarations
		case t == "class_declaration" || t == "struct_item" || t == "struct_specifier" ||
			t == "impl_item":
			ci := types.ClassInfo{
				StartLine: int(child.StartPoint().Row) + 1,
				EndLine:   int(child.EndPoint().Row) + 1,
				Kind:      t,
			}
			for j := 0; j < int(child.ChildCount()); j++ {
				c := child.Child(j)
				if c.Type() == "identifier" || c.Type() == "type_identifier" {
					ci.Name = c.Content(code)
					break
				}
			}
			if ci.Name != "" {
				result.Classes = append(result.Classes, ci)
			}
			// Recurse into body for nested items
			visitGenericNode(child, code, result, lang)

		default:
			// Recurse into children
			if child.ChildCount() > 0 {
				visitGenericNode(child, code, result, lang)
			}
		}
	}
}
