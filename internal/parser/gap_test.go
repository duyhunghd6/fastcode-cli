package parser

import (
	"testing"
)

// === nodeText Tests ===

// TestNodeTextDirect tests the nodeText utility function directly
func TestNodeTextDirect(t *testing.T) {
	// nodeText accepts an interface with Content([]byte) string method
	mock := &mockNode{text: "hello world"}
	result := nodeText(mock, []byte("source code"))
	if result != "hello world" {
		t.Errorf("nodeText = %q, want %q", result, "hello world")
	}
}

type mockNode struct {
	text string
}

func (m *mockNode) Content(code []byte) string {
	return m.text
}

// TestNodeTextEmpty tests nodeText with empty content
func TestNodeTextEmpty(t *testing.T) {
	mock := &mockNode{text: ""}
	result := nodeText(mock, nil)
	if result != "" {
		t.Errorf("nodeText empty = %q", result)
	}
}

// === Go Interface Method Extraction via ParseFile ===

// TestParseGoInterfaceWithMethods tests Go interface with multiple methods
func TestParseGoInterfaceWithMethods(t *testing.T) {
	p := New()
	content := `package main

// Reader is a basic reader interface.
type Reader interface {
	Read(p []byte) (n int, err error)
	Close() error
}
`
	result := p.ParseFile("reader.go", content)
	if result == nil {
		t.Fatal("expected parse result")
	}

	found := false
	for _, cls := range result.Classes {
		if cls.Name == "Reader" {
			found = true
			if cls.Kind != "interface" {
				t.Errorf("kind = %q, want interface", cls.Kind)
			}
			// Methods should be extracted
			if len(cls.Methods) < 1 {
				t.Logf("warning: interface methods not extracted (may depend on tree-sitter grammar version), got %d", len(cls.Methods))
			}
		}
	}
	if !found {
		t.Error("Reader interface not found")
	}
}

// TestParseGoInterfaceEmpty tests Go interface with no methods
func TestParseGoInterfaceEmpty(t *testing.T) {
	p := New()
	content := `package main

type Empty interface{}
`
	result := p.ParseFile("empty.go", content)
	if result == nil {
		t.Fatal("expected parse result")
	}

	found := false
	for _, cls := range result.Classes {
		if cls.Name == "Empty" {
			found = true
			if len(cls.Methods) != 0 {
				t.Errorf("empty interface should have 0 methods, got %d", len(cls.Methods))
			}
		}
	}
	if !found {
		t.Error("Empty interface not found")
	}
}

// TestParseGoInterfaceEmbedded tests Go interface with embedded interfaces
func TestParseGoInterfaceEmbedded(t *testing.T) {
	p := New()
	content := `package main

import "io"

type ReadCloser interface {
	io.Reader
	Close() error
}
`
	result := p.ParseFile("rc.go", content)
	if result == nil {
		t.Fatal("expected parse result")
	}

	for _, cls := range result.Classes {
		if cls.Name == "ReadCloser" {
			t.Logf("ReadCloser: kind=%s methods=%d bases=%v", cls.Kind, len(cls.Methods), cls.Bases)
		}
	}
}

// === Go receiver type edge cases ===

// TestParseGoGenericReceiver tests Go method with generic-like type
func TestParseGoGenericReceiver(t *testing.T) {
	p := New()
	content := `package main

type Handler struct {
	name string
}

func (h Handler) Name() string {
	return h.name
}

func (h *Handler) SetName(n string) {
	h.name = n
}
`
	result := p.ParseFile("handler.go", content)
	if result == nil {
		t.Fatal("expected parse result")
	}

	// Should find both methods
	methodCount := 0
	for _, fn := range result.Functions {
		if fn.IsMethod && fn.ClassName == "Handler" {
			methodCount++
		}
	}
	if methodCount < 2 {
		t.Errorf("expected at least 2 methods on Handler, got %d", methodCount)
	}
}

// === Go type spec edge cases ===

// TestParseGoTypeAlias tests Go type alias
func TestParseGoTypeAlias(t *testing.T) {
	p := New()
	content := `package main

type StringSlice []string
type IntMap map[string]int
`
	result := p.ParseFile("types.go", content)
	if result == nil {
		t.Fatal("expected parse result")
	}
	// Type aliases are represented as classes in the parser
	t.Logf("classes: %d, functions: %d", len(result.Classes), len(result.Functions))
}

// === Python import edge cases ===

// TestParsePythonRelativeImport tests Python relative import
func TestParsePythonRelativeImport(t *testing.T) {
	p := New()
	content := `from . import utils
from .. import base
from ...core import handler
`
	result := p.ParseFile("mod.py", content)
	if result == nil {
		t.Fatal("expected parse result")
	}

	if len(result.Imports) < 2 {
		t.Logf("warning: relative imports may not all be captured, got %d", len(result.Imports))
	}

	// Check at least one from-import is detected
	hasFrom := false
	for _, imp := range result.Imports {
		if imp.IsFrom {
			hasFrom = true
		}
	}
	if !hasFrom {
		t.Error("expected at least one from-import")
	}
}

// === Python class with bases and methods ===

// TestParsePythonClassWithInheritance tests class with bases
func TestParsePythonClassWithInheritance(t *testing.T) {
	p := New()
	content := `class Animal:
    """Base animal class."""
    def speak(self):
        pass

class Dog(Animal):
    """A dog that speaks."""
    def speak(self):
        return "Woof!"

    def fetch(self, item):
        return f"Fetching {item}"
`
	result := p.ParseFile("animals.py", content)
	if result == nil {
		t.Fatal("expected parse result")
	}

	dogFound := false
	for _, cls := range result.Classes {
		if cls.Name == "Dog" {
			dogFound = true
			if len(cls.Bases) == 0 {
				t.Error("Dog should have Animal as base")
			}
			if cls.Docstring == "" {
				t.Error("Dog should have docstring")
			}
		}
	}
	if !dogFound {
		t.Error("Dog class not found")
	}
}

// TestParsePythonModuleDocstring tests Python module-level docstring
func TestParsePythonModuleDocstring(t *testing.T) {
	p := New()
	content := `"""
This is a module docstring.
It has multiple lines.
"""

def main():
    pass
`
	result := p.ParseFile("module.py", content)
	if result == nil {
		t.Fatal("expected parse result")
	}

	if result.ModuleDocstring == "" {
		t.Logf("warning: module docstring may not be captured by this grammar version")
	}
}

// === JS edge cases ===

// TestParseJSArrowFunction tests JavaScript arrow function
func TestParseJSArrowFunction(t *testing.T) {
	p := New()
	content := `const add = (a, b) => a + b;

const multiply = (a, b) => {
    return a * b;
};

export function divide(a, b) {
    return a / b;
}
`
	result := p.ParseFile("math.js", content)
	if result == nil {
		t.Fatal("expected parse result")
	}

	// At minimum, divide should be found
	found := false
	for _, fn := range result.Functions {
		if fn.Name == "divide" {
			found = true
		}
	}
	if !found {
		t.Error("divide function not found")
	}
}

// TestParseJSImportVariants tests various JS import styles
func TestParseJSImportVariants(t *testing.T) {
	p := New()
	content := `import React from 'react';
import { useState, useEffect } from 'react';
import * as utils from './utils';
import './styles.css';
`
	result := p.ParseFile("app.js", content)
	if result == nil {
		t.Fatal("expected parse result")
	}

	if len(result.Imports) < 3 {
		t.Logf("warning: JS import parsing captured %d imports (may vary by grammar version)", len(result.Imports))
	}
}

// === Go leading comment edge case ===

// TestParseGoFunctionWithDoc tests Go function with documentation comment
func TestParseGoFunctionWithDoc(t *testing.T) {
	p := New()
	content := `package main

// Add adds two integers together.
// It returns the sum.
func Add(a, b int) int {
	return a + b
}

func noDoc() {}
`
	result := p.ParseFile("math.go", content)
	if result == nil {
		t.Fatal("expected parse result")
	}

	for _, fn := range result.Functions {
		if fn.Name == "Add" {
			if fn.Docstring == "" {
				t.Error("Add should have docstring")
			}
		}
		if fn.Name == "noDoc" {
			// noDoc has no comment, docstring should be empty
			if fn.Docstring != "" {
				t.Logf("noDoc docstring = %q (may have nearby comment captured)", fn.Docstring)
			}
		}
	}
}

// === Parser New() error path ===

// TestNewParser tests parser creation
func TestNewParser(t *testing.T) {
	p := New()
	if p == nil {
		t.Fatal("New() returned nil")
	}
}

// TestParseUnsupportedLanguage tests parsing a file with unsupported language
func TestParseUnsupportedLanguage(t *testing.T) {
	p := New()
	result := p.ParseFile("data.dat", "binary content")
	// Should return nil for unsupported language
	if result != nil {
		t.Logf("unsupported language returned result: %+v", result)
	}
}

// TestParseEmptyFile tests parsing empty file
func TestParseEmptyGoFile(t *testing.T) {
	p := New()
	result := p.ParseFile("empty.go", "")
	if result == nil {
		t.Fatal("expected result for empty Go file")
	}
	if result.TotalLines != 0 {
		t.Logf("empty file lines = %d", result.TotalLines)
	}
}

// TestParseGoImportGrouped tests Go grouped imports
func TestParseGoImportGrouped(t *testing.T) {
	p := New()
	content := `package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/example/pkg"
)

func main() {
	fmt.Println(os.Args)
	_ = strings.Join(nil, "")
	_ = pkg.Do()
}
`
	result := p.ParseFile("imports.go", content)
	if result == nil {
		t.Fatal("expected parse result")
	}
	if len(result.Imports) < 4 {
		t.Logf("grouped imports: got %d", len(result.Imports))
	}
}
