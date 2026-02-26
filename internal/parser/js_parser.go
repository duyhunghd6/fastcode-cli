package parser

import (
	"github.com/duyhunghd6/fastcode-cli/internal/types"
	sitter "github.com/smacker/go-tree-sitter"
)

// parseJS extracts classes, functions, and imports from JavaScript/TypeScript source.
func parseJS(root *sitter.Node, code []byte, result *types.FileParseResult) {
	visitJSNode(root, code, result, "")
}

func visitJSNode(node *sitter.Node, code []byte, result *types.FileParseResult, className string) {
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		switch child.Type() {
		case "import_statement":
			result.Imports = append(result.Imports, extractJSImport(child, code))
		case "class_declaration":
			ci := extractJSClass(child, code)
			result.Classes = append(result.Classes, ci)
		case "function_declaration":
			fn := extractJSFunction(child, code, className)
			if fn.Name != "" {
				result.Functions = append(result.Functions, fn)
			}
		case "export_statement":
			visitJSNode(child, code, result, className)
		case "lexical_declaration":
			// Handle: const foo = () => {} or const foo = function() {}
			fns := extractJSArrowFunctions(child, code)
			result.Functions = append(result.Functions, fns...)
		}
	}
}

func extractJSImport(node *sitter.Node, code []byte) types.ImportInfo {
	imp := types.ImportInfo{
		Line:   int(node.StartPoint().Row) + 1,
		IsFrom: true,
	}
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		switch child.Type() {
		case "string", "template_string":
			imp.Module = trimQuotes(child.Content(code))
		case "import_clause":
			for j := 0; j < int(child.ChildCount()); j++ {
				c := child.Child(j)
				switch c.Type() {
				case "identifier":
					imp.Names = append(imp.Names, c.Content(code))
				case "named_imports":
					for k := 0; k < int(c.ChildCount()); k++ {
						spec := c.Child(k)
						if spec.Type() == "import_specifier" {
							for l := 0; l < int(spec.ChildCount()); l++ {
								s := spec.Child(l)
								if s.Type() == "identifier" {
									imp.Names = append(imp.Names, s.Content(code))
									break
								}
							}
						}
					}
				case "namespace_import":
					imp.Alias = c.Content(code)
				}
			}
		}
	}
	return imp
}

func extractJSClass(node *sitter.Node, code []byte) types.ClassInfo {
	ci := types.ClassInfo{
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
		Kind:      "class",
	}
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		switch child.Type() {
		case "identifier":
			ci.Name = child.Content(code)
		case "class_heritage":
			for j := 0; j < int(child.ChildCount()); j++ {
				c := child.Child(j)
				if c.Type() == "identifier" || c.Type() == "member_expression" {
					ci.Bases = append(ci.Bases, c.Content(code))
				}
			}
		case "class_body":
			ci.Methods = extractJSClassMethods(child, code, ci.Name)
		}
	}
	return ci
}

func extractJSClassMethods(body *sitter.Node, code []byte, className string) []types.FunctionInfo {
	var methods []types.FunctionInfo
	for i := 0; i < int(body.ChildCount()); i++ {
		child := body.Child(i)
		if child.Type() == "method_definition" {
			fn := types.FunctionInfo{
				StartLine: int(child.StartPoint().Row) + 1,
				EndLine:   int(child.EndPoint().Row) + 1,
				IsMethod:  true,
				ClassName: className,
			}
			for j := 0; j < int(child.ChildCount()); j++ {
				c := child.Child(j)
				switch c.Type() {
				case "property_identifier":
					fn.Name = c.Content(code)
				case "formal_parameters":
					fn.Parameters = extractJSParams(c, code)
				}
			}
			// Check async
			text := child.Content(code)
			if len(text) > 5 && text[:5] == "async" {
				fn.IsAsync = true
			}
			methods = append(methods, fn)
		}
	}
	return methods
}

func extractJSFunction(node *sitter.Node, code []byte, className string) types.FunctionInfo {
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
		case "formal_parameters":
			fn.Parameters = extractJSParams(child, code)
		case "type_annotation":
			fn.ReturnType = child.Content(code)
		}
	}
	text := node.Content(code)
	if len(text) > 5 && text[:5] == "async" {
		fn.IsAsync = true
	}
	return fn
}

func extractJSArrowFunctions(node *sitter.Node, code []byte) []types.FunctionInfo {
	var fns []types.FunctionInfo
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "variable_declarator" {
			fn := types.FunctionInfo{
				StartLine: int(child.StartPoint().Row) + 1,
				EndLine:   int(child.EndPoint().Row) + 1,
			}
			for j := 0; j < int(child.ChildCount()); j++ {
				c := child.Child(j)
				switch c.Type() {
				case "identifier":
					fn.Name = c.Content(code)
				case "arrow_function", "function":
					for k := 0; k < int(c.ChildCount()); k++ {
						p := c.Child(k)
						if p.Type() == "formal_parameters" {
							fn.Parameters = extractJSParams(p, code)
						}
					}
				}
			}
			if fn.Name != "" {
				fns = append(fns, fn)
			}
		}
	}
	return fns
}

func extractJSParams(node *sitter.Node, code []byte) []string {
	var params []string
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		switch child.Type() {
		case "identifier", "assignment_pattern", "rest_pattern",
			"object_pattern", "array_pattern", "required_parameter",
			"optional_parameter":
			params = append(params, child.Content(code))
		}
	}
	return params
}

func trimQuotes(s string) string {
	if len(s) >= 2 {
		if (s[0] == '\'' && s[len(s)-1] == '\'') ||
			(s[0] == '"' && s[len(s)-1] == '"') ||
			(s[0] == '`' && s[len(s)-1] == '`') {
			return s[1 : len(s)-1]
		}
	}
	return s
}
