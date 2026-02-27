package parser

import (
	"strings"

	"github.com/duyhunghd6/fastcode-cli/internal/types"
	sitter "github.com/smacker/go-tree-sitter"
)

// parseJava extracts classes, methods, and imports from Java source.
// Stub implementation — to be fully ported from Python parser.
func parseJava(root *sitter.Node, code []byte, result *types.FileParseResult) {
	visitGenericNode(root, code, result, "java")
}

// parseRust extracts structs, impl blocks, functions, and use statements from Rust source.
// Matches Python's _parse_rust/_extract_rust_items behavior exactly.
func parseRust(root *sitter.Node, code []byte, result *types.FileParseResult) {
	// Extract module-level docstring: scan ALL root children for first comment
	// (Python's _extract_rust_module_docstring does NOT stop at non-comment nodes)
	for i := 0; i < int(root.ChildCount()); i++ {
		child := root.Child(i)
		t := child.Type()
		if t == "line_comment" || t == "block_comment" {
			commentText := child.Content(code)
			if strings.HasPrefix(commentText, "//!") {
				result.ModuleDocstring = strings.TrimSpace(commentText[3:])
				break
			} else if strings.HasPrefix(commentText, "///") {
				result.ModuleDocstring = strings.TrimSpace(commentText[3:])
				break
			} else if strings.HasPrefix(commentText, "/*") && strings.HasSuffix(commentText, "*/") {
				result.ModuleDocstring = strings.TrimSpace(commentText[2 : len(commentText)-2])
				break
			}
			// Regular // comment — skip and continue scanning (matches Python)
		}
		// Do NOT break on non-comment nodes — Python scans all root children
	}
	// Use Rust-specific visitor matching Python's _extract_rust_items
	visitRustNode(root, code, result)
}

// visitRustNode matches Python's _extract_rust_items visit_node exactly:
// - struct_item/trait_item/impl_item → class (do NOT recurse for nested functions)
// - function_item → function
// - else → recurse into children
func visitRustNode(node *sitter.Node, code []byte, result *types.FileParseResult) {
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		t := child.Type()

		if t == "struct_item" || t == "trait_item" || t == "impl_item" {
			// Extract as class (with methods embedded)
			ci := extractRustType(child, code)
			if ci != nil {
				result.Classes = append(result.Classes, *ci)
			}
			// Python does NOT recurse into these — elif chain stops here
		} else if t == "function_item" {
			// Top-level function
			fn := extractRustFunction(child, code, "")
			if fn != nil {
				result.Functions = append(result.Functions, *fn)
			}
		} else {
			// Recurse into children
			if child.ChildCount() > 0 {
				visitRustNode(child, code, result)
			}
		}
	}
}

// extractRustType extracts struct/trait/impl info including methods.
func extractRustType(node *sitter.Node, code []byte) *types.ClassInfo {
	t := node.Type()
	var name string

	// Find name
	for j := 0; j < int(node.ChildCount()); j++ {
		c := node.Child(j)
		if c.Type() == "type_identifier" || c.Type() == "identifier" {
			name = c.Content(code)
			break
		}
	}
	if name == "" {
		return nil
	}

	ci := &types.ClassInfo{
		Name:      name,
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		Kind:      t,
	}

	// Extract methods from impl/trait body
	if t == "impl_item" || t == "trait_item" {
		for j := 0; j < int(node.ChildCount()); j++ {
			c := node.Child(j)
			if c.Type() == "declaration_list" {
				for k := 0; k < int(c.ChildCount()); k++ {
					member := c.Child(k)
					if member.Type() == "function_item" {
						fn := extractRustFunction(member, code, name)
						if fn != nil {
							ci.Methods = append(ci.Methods, *fn)
						}
					}
				}
			}
		}
	}

	return ci
}

// extractRustFunction extracts function info from a Rust function_item node.
func extractRustFunction(node *sitter.Node, code []byte, className string) *types.FunctionInfo {
	var funcName string
	for j := 0; j < int(node.ChildCount()); j++ {
		c := node.Child(j)
		if c.Type() == "identifier" {
			funcName = c.Content(code)
			break
		}
	}
	if funcName == "" {
		return nil
	}

	fn := &types.FunctionInfo{
		Name:      funcName,
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		ClassName: className,
		IsMethod:  className != "",
	}
	return fn
}

// parseC extracts structs, functions, and includes from C/C++ source.
// Matches Python's _parse_c_cpp: extracts module docstrings from leading
// comments and function docstrings from preceding comments.
func parseC(root *sitter.Node, code []byte, result *types.FileParseResult, lang string) {
	// Extract module-level docstring from leading comment
	// (matches Python's _extract_c_module_docstring)
	for i := 0; i < int(root.ChildCount()); i++ {
		child := root.Child(i)
		if child.Type() == "comment" {
			commentText := child.Content(code)
			// Clean up comment markers
			if len(commentText) >= 4 && commentText[:2] == "/*" && commentText[len(commentText)-2:] == "*/" {
				result.ModuleDocstring = strings.TrimSpace(commentText[2 : len(commentText)-2])
			} else if len(commentText) >= 2 && commentText[:2] == "//" {
				result.ModuleDocstring = strings.TrimSpace(commentText[2:])
			} else {
				result.ModuleDocstring = commentText
			}
			break
		}
		// Stop if we hit a non-comment node (only look at leading comments)
		if child.Type() != "comment" {
			break
		}
	}

	// Extract classes/structs and functions with docstrings
	visitCNode(root, code, result, lang)
}

// visitCNode extracts functions and classes from C/C++ AST,
// matching Python's visit_node behavior exactly:
// - function_definition: requires function_declarator child (else skip)
// - class_specifier/struct_specifier: extract methods from field_declaration_list
// - else: recurse into children
// Note: when lang is "c", class_specifier is skipped because Python's C grammar
// does not include class_specifier (it's a C++ construct), but Go's go-tree-sitter
// C grammar erroneously includes it.
func visitCNode(node *sitter.Node, code []byte, result *types.FileParseResult, lang string) {
	visitCNodeAtDepth(node, code, result, lang, 0)
}

func visitCNodeAtDepth(node *sitter.Node, code []byte, result *types.FileParseResult, lang string, depth int) {
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		t := child.Type()

		if t == "class_specifier" || t == "struct_specifier" {
			// Extract class/struct info (matches Python's _extract_c_class)
			ci := extractCClass(child, code, result)
			if ci != nil {
				// Extract docstring from preceding comment
				if i > 0 {
					prev := node.Child(i - 1)
					if prev.Type() == "comment" {
						ci.Docstring = cleanCComment(prev.Content(code))
					}
				}
				result.Classes = append(result.Classes, *ci)
			}
			// Python does NOT recurse into class/struct bodies beyond method extraction
		} else if t == "function_definition" {
			// Extract function (matches Python's _extract_c_function)
			fn := extractCFunction(child, code, "")
			if fn != nil {
				// Extract docstring from preceding comment
				if i > 0 {
					prev := node.Child(i - 1)
					if prev.Type() == "comment" {
						fn.Docstring = cleanCComment(prev.Content(code))
					}
				}
				result.Functions = append(result.Functions, *fn)
			}
		} else if t == "ERROR" && depth == 0 {
			// Skip root-level ERROR nodes only. Go's C grammar produces ERROR nodes
			// at root where Python's grammar wraps the same content in function_definition
			// (which stops recursion due to elif). Python's ERROR nodes at root have
			// 0 parseable children (verified). At deeper levels, ERROR nodes may contain
			// valid elements that both grammars would find.
		} else {
			// Recurse into children (matches Python's else clause)
			if child.ChildCount() > 0 {
				visitCNodeAtDepth(child, code, result, lang, depth+1)
			}
		}
	}
}

// extractCClass extracts class/struct info from a C/C++ AST node,
// including methods from field_declaration_list (matching Python's _extract_c_class).
func extractCClass(node *sitter.Node, code []byte, result *types.FileParseResult) *types.ClassInfo {
	// Find name
	var name string
	for j := 0; j < int(node.ChildCount()); j++ {
		c := node.Child(j)
		if c.Type() == "type_identifier" || c.Type() == "identifier" {
			name = c.Content(code)
			break
		}
	}
	if name == "" {
		return nil
	}

	ci := &types.ClassInfo{
		Name:      name,
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		Kind:      node.Type(),
	}

	// Extract methods from field_declaration_list (matches Python)
	for j := 0; j < int(node.ChildCount()); j++ {
		c := node.Child(j)
		if c.Type() == "field_declaration_list" {
			for k := 0; k < int(c.ChildCount()); k++ {
				member := c.Child(k)
				if member.Type() == "function_definition" {
					fn := extractCFunction(member, code, name)
					if fn != nil {
						ci.Methods = append(ci.Methods, *fn)
					}
				}
			}
		}
	}

	return ci
}

// extractCFunction extracts function info from a C/C++ function_definition node.
// Matches Python's _extract_c_function: requires function_declarator child.
func extractCFunction(node *sitter.Node, code []byte, className string) *types.FunctionInfo {
	// Find function_declarator (Python returns None without it)
	var declarator *sitter.Node
	for j := 0; j < int(node.ChildCount()); j++ {
		c := node.Child(j)
		if c.Type() == "function_declarator" {
			declarator = c
			break
		}
	}
	if declarator == nil {
		return nil
	}

	// Get function name from declarator
	var funcName string
	for j := 0; j < int(declarator.ChildCount()); j++ {
		c := declarator.Child(j)
		if c.Type() == "identifier" || c.Type() == "field_identifier" {
			funcName = c.Content(code)
			break
		}
	}
	if funcName == "" {
		return nil
	}

	fn := &types.FunctionInfo{
		Name:      funcName,
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		ClassName: className,
		IsMethod:  className != "",
	}

	return fn
}

// cleanCComment removes comment markers from C-style comments.
func cleanCComment(comment string) string {
	if len(comment) >= 5 && comment[:3] == "/**" && comment[len(comment)-2:] == "*/" {
		return strings.TrimSpace(comment[3 : len(comment)-2])
	}
	if len(comment) >= 4 && comment[:2] == "/*" && comment[len(comment)-2:] == "*/" {
		return strings.TrimSpace(comment[2 : len(comment)-2])
	}
	if len(comment) >= 3 && comment[:3] == "///" {
		return strings.TrimSpace(comment[3:])
	}
	if len(comment) >= 2 && comment[:2] == "//" {
		return strings.TrimSpace(comment[2:])
	}
	return comment
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
