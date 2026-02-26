package parser

import (
	"strings"

	"github.com/duyhunghd6/fastcode-cli/internal/types"
	sitter "github.com/smacker/go-tree-sitter"
)

// parseJS extracts classes, functions, and imports from JavaScript/TypeScript source.
func parseJS(root *sitter.Node, code []byte, result *types.FileParseResult) {
	// Extract module-level docstring from first comment (matches Python behavior).
	for i := 0; i < int(root.ChildCount()); i++ {
		child := root.Child(i)
		if child.Type() == "comment" {
			text := child.Content(code)
			if strings.HasPrefix(text, "//") {
				result.ModuleDocstring = strings.TrimSpace(text[2:])
			} else if strings.HasPrefix(text, "/*") && strings.HasSuffix(text, "*/") {
				result.ModuleDocstring = strings.TrimSpace(text[2 : len(text)-2])
			}
			break
		}
		// Stop at first non-comment node
		if child.Type() != "comment" {
			break
		}
	}
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
		case "interface_declaration", "type_alias_declaration":
			// Python treats interfaces and type aliases as classes
			ci := extractJSInterface(child, code)
			if ci.Name != "" {
				result.Classes = append(result.Classes, ci)
			}
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

// extractJSInterface extracts interface/type alias information as a ClassInfo.
// Python's parser treats these as classes, so we match that behavior.
func extractJSInterface(node *sitter.Node, code []byte) types.ClassInfo {
	ci := types.ClassInfo{
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
	}
	// Determine kind
	if node.Type() == "interface_declaration" {
		ci.Kind = "interface"
	} else {
		ci.Kind = "type"
	}
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		switch child.Type() {
		case "type_identifier", "identifier":
			if ci.Name == "" {
				ci.Name = child.Content(code)
			}
		case "extends_clause", "extends_type_clause":
			for j := 0; j < int(child.ChildCount()); j++ {
				c := child.Child(j)
				if c.Type() == "type_identifier" || c.Type() == "identifier" {
					ci.Bases = append(ci.Bases, c.Content(code))
				}
			}
		case "interface_body", "object_type":
			// Extract method signatures from interface body
			ci.Methods = extractJSInterfaceMethods(child, code, ci.Name)
		}
	}
	return ci
}

// extractJSInterfaceMethods extracts method signatures from interface bodies.
func extractJSInterfaceMethods(body *sitter.Node, code []byte, className string) []types.FunctionInfo {
	var methods []types.FunctionInfo
	for i := 0; i < int(body.ChildCount()); i++ {
		child := body.Child(i)
		if child.Type() == "method_signature" || child.Type() == "method_definition" {
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
			if fn.Name != "" {
				methods = append(methods, fn)
			}
		}
	}
	return methods
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
				case "statement_block":
					fn.Calls = extractJSCalls(c, code)
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
		case "statement_block":
			fn.Calls = extractJSCalls(child, code)
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
						if p.Type() == "statement_block" {
							fn.Calls = extractJSCalls(p, code)
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

// jsBuiltins contains JS/TS built-in names to filter out of call graphs.
var jsBuiltins = map[string]bool{
	"console": true, "setTimeout": true, "setInterval": true,
	"clearTimeout": true, "clearInterval": true, "requestAnimationFrame": true,
	"cancelAnimationFrame": true, "fetch": true, "require": true,
	"parseInt": true, "parseFloat": true, "isNaN": true, "isFinite": true,
	"encodeURIComponent": true, "decodeURIComponent": true,
	"encodeURI": true, "decodeURI": true, "alert": true, "confirm": true,
	"JSON": true, "Object": true, "Array": true, "Math": true,
	"String": true, "Number": true, "Boolean": true, "Symbol": true,
	"Map": true, "Set": true, "WeakMap": true, "WeakSet": true,
	"Date": true, "Error": true, "RegExp": true, "Promise": true,
	"Proxy": true, "Reflect": true, "WeakRef": true,
	"Uint8Array": true, "Int8Array": true, "Float32Array": true, "Float64Array": true,
	"ArrayBuffer": true, "DataView": true, "BigInt": true,
	// React built-ins
	"React": true, "createElement": true, "Fragment": true,
}

// extractJSCalls recursively walks a function body to extract call_expression nodes.
// Returns a deduplicated list of callee names (function/method names being called).
func extractJSCalls(node *sitter.Node, code []byte) []string {
	seen := make(map[string]bool)
	collectJSCalls(node, code, seen)

	if len(seen) == 0 {
		return nil
	}

	calls := make([]string, 0, len(seen))
	for name := range seen {
		calls = append(calls, name)
	}
	return calls
}

// collectJSCalls recursively walks the AST and collects function call names.
func collectJSCalls(node *sitter.Node, code []byte, seen map[string]bool) {
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "call_expression" {
			name := extractJSCalleeName(child, code)
			if name != "" && !jsBuiltins[name] {
				seen[name] = true
			}
		}
		// Recurse into all children (call_expression can be nested)
		collectJSCalls(child, code, seen)
	}
}

// extractJSCalleeName extracts the function/method name from a call_expression node.
// For simple calls like foo(), returns "foo".
// For method calls like obj.bar(), returns "bar".
// For chained calls like a.b.c(), returns "c".
func extractJSCalleeName(callNode *sitter.Node, code []byte) string {
	if callNode.ChildCount() == 0 {
		return ""
	}

	funcNode := callNode.Child(0)
	switch funcNode.Type() {
	case "identifier":
		// Simple call: foo()
		return funcNode.Content(code)
	case "member_expression":
		// Method call: obj.bar() or a.b.c()
		// Extract the rightmost property (the method name)
		for j := int(funcNode.ChildCount()) - 1; j >= 0; j-- {
			prop := funcNode.Child(j)
			if prop.Type() == "property_identifier" {
				return prop.Content(code)
			}
		}
	}
	return ""
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
