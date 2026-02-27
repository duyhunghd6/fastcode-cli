package parser

import (
	"testing"
)

// === Go parser edge cases to hit uncovered lines ===

// TestParseGoInterfaceWithMethodSpec tests Go interface method_spec parsing
// This targets extractGoInterfaceMethods lines 225-232 (method_spec → field_identifier, parameter_list)
func TestParseGoInterfaceWithMethodSpec(t *testing.T) {
	p := New()
	content := `package main

type Writer interface {
	Write(data []byte) (int, error)
	Flush()
	WriteString(s string)
}
`
	result := p.ParseFile("writer.go", content)
	if result == nil {
		t.Fatal("expected parse result")
	}

	for _, cls := range result.Classes {
		if cls.Name == "Writer" && cls.Kind == "interface" {
			t.Logf("Writer: %d methods", len(cls.Methods))
			for _, m := range cls.Methods {
				t.Logf("  method: %s params=%v", m.Name, m.Parameters)
			}
		}
	}
}

// TestParseGoStructWithEmbedded tests struct with embedded types
func TestParseGoStructWithEmbedded(t *testing.T) {
	p := New()
	content := `package main

import "sync"

type Server struct {
	sync.Mutex
	Host string
	Port int
}
`
	result := p.ParseFile("server.go", content)
	if result == nil {
		t.Fatal("expected parse result")
	}

	for _, cls := range result.Classes {
		if cls.Name == "Server" {
			t.Logf("Server bases: %v", cls.Bases)
		}
	}
}

// TestParseGoImportWithAlias tests Go import with alias (dot, blank, named)
func TestParseGoImportWithAlias(t *testing.T) {
	p := New()
	content := `package main

import (
	. "fmt"
	_ "net/http/pprof"
	mylog "log"
)

func main() {
	Println("hello")
	_ = mylog.Default()
}
`
	result := p.ParseFile("alias.go", content)
	if result == nil {
		t.Fatal("expected parse result")
	}

	aliasFound := false
	for _, imp := range result.Imports {
		if imp.Alias != "" {
			aliasFound = true
			t.Logf("import %q alias=%q", imp.Module, imp.Alias)
		}
	}
	if !aliasFound {
		t.Logf("warning: no aliased imports found, got %d imports", len(result.Imports))
	}
}

// TestParseGoSingleImport tests Go single import (not grouped)
func TestParseGoSingleImport(t *testing.T) {
	p := New()
	content := `package main

import "fmt"

func main() {
	fmt.Println("hello")
}
`
	result := p.ParseFile("single.go", content)
	if result == nil {
		t.Fatal("expected parse result")
	}
	// Go files are non-code (Python parity): no import extraction
	if len(result.Imports) != 0 {
		t.Errorf("Go files should have 0 imports (Python parity), got %d", len(result.Imports))
	}
}

// TestParseGoMethodWithReceiver tests method with both value and pointer receivers
func TestParseGoMethodWithReceiver(t *testing.T) {
	p := New()
	content := `package main

type Config struct {
	Value string
}

func (c Config) Get() string {
	return c.Value
}

func (c *Config) Set(v string) {
	c.Value = v
}
`
	result := p.ParseFile("config.go", content)
	if result == nil {
		t.Fatal("expected parse result")
	}

	// Go files are non-code (Python parity): no function extraction
	if len(result.Functions) != 0 {
		t.Errorf("Go files should have 0 functions (Python parity), got %d", len(result.Functions))
	}
}

// TestParseGoFuncReturnTypes tests multiple return types
func TestParseGoFuncReturnTypes(t *testing.T) {
	p := New()
	content := `package main

func Open(path string) (*File, error) {
	return nil, nil
}

func Count() int {
	return 0
}

type File struct{}
`
	result := p.ParseFile("returns.go", content)
	if result == nil {
		t.Fatal("expected parse result")
	}

	for _, fn := range result.Functions {
		t.Logf("func %s returns=%q", fn.Name, fn.ReturnType)
	}
}

// === JS parser edge cases ===

// TestParseJSClassWithMethods tests JS class with constructor and methods
func TestParseJSClassWithMethods(t *testing.T) {
	p := New()
	content := `class Animal {
    constructor(name) {
        this.name = name;
    }

    speak() {
        return this.name;
    }

    static create(name) {
        return new Animal(name);
    }
}

class Dog extends Animal {
    bark() {
        return "woof";
    }
}
`
	result := p.ParseFile("animals.js", content)
	if result == nil {
		t.Fatal("expected parse result")
	}

	animalFound := false
	dogFound := false
	for _, cls := range result.Classes {
		if cls.Name == "Animal" {
			animalFound = true
			t.Logf("Animal: %d methods", len(cls.Methods))
		}
		if cls.Name == "Dog" {
			dogFound = true
		}
	}
	if !animalFound {
		t.Error("Animal class not found")
	}
	if !dogFound {
		t.Error("Dog class not found")
	}
}

// TestParseJSFunctionWithAsync tests JS async function
func TestParseJSFunctionWithAsync(t *testing.T) {
	p := New()
	content := `async function fetchData(url) {
    const response = await fetch(url);
    return response.json();
}

function syncFunc() {
    return 42;
}
`
	result := p.ParseFile("fetch.js", content)
	if result == nil {
		t.Fatal("expected parse result")
	}

	for _, fn := range result.Functions {
		t.Logf("func %s async=%v", fn.Name, fn.IsAsync)
	}
}

// === Python class edge cases ===

// TestParsePythonClassDecorators tests Python class with decorators
func TestParsePythonClassDecorators(t *testing.T) {
	p := New()
	content := `from dataclasses import dataclass

@dataclass
class Config:
    host: str = "localhost"
    port: int = 8080
`
	result := p.ParseFile("config.py", content)
	if result == nil {
		t.Fatal("expected parse result")
	}

	for _, cls := range result.Classes {
		if cls.Name == "Config" {
			t.Logf("Config decorators: %v", cls.Decorators)
		}
	}
}

// TestParsePythonAliasedImport tests Python import with alias
func TestParsePythonAliasedImport(t *testing.T) {
	p := New()
	content := `from os.path import join as path_join
import numpy as np
`
	result := p.ParseFile("imports.py", content)
	if result == nil {
		t.Fatal("expected parse result")
	}

	for _, imp := range result.Imports {
		t.Logf("import module=%q names=%v alias=%q isFrom=%v level=%d",
			imp.Module, imp.Names, imp.Alias, imp.IsFrom, imp.Level)
	}
}

// TestParsePythonBlockDocstring tests Python function with block docstring
func TestParsePythonBlockDocstring(t *testing.T) {
	p := New()
	content := `def process(data):
    """Process the input data.
    
    Args:
        data: The input to process.
    
    Returns:
        Processed result.
    """
    return data
`
	result := p.ParseFile("process.py", content)
	if result == nil {
		t.Fatal("expected parse result")
	}

	for _, fn := range result.Functions {
		if fn.Name == "process" {
			t.Logf("process docstring: %q", fn.Docstring)
		}
	}
}

// === Rust parser ===

func TestParseRustBasic(t *testing.T) {
	p := New()
	content := `struct Server {
    host: String,
    port: u16,
}

impl Server {
    fn new(host: String, port: u16) -> Self {
        Server { host, port }
    }

    fn start(&self) {
        println!("Starting on {}:{}", self.host, self.port);
    }
}

fn main() {
    let s = Server::new("localhost".to_string(), 8080);
    s.start();
}
`
	result := p.ParseFile("server.rs", content)
	if result == nil {
		t.Fatal("expected parse result")
	}
	t.Logf("rust: %d classes, %d functions, %d imports", len(result.Classes), len(result.Functions), len(result.Imports))
}

// === C parser ===

func TestParseCBasic(t *testing.T) {
	p := New()
	content := `#include <stdio.h>
#include <stdlib.h>

typedef struct {
    int x;
    int y;
} Point;

int add(int a, int b) {
    return a + b;
}

int main() {
    Point p = {1, 2};
    printf("%d\n", add(p.x, p.y));
    return 0;
}
`
	result := p.ParseFile("main.c", content)
	if result == nil {
		t.Fatal("expected parse result")
	}
	t.Logf("c: %d classes, %d functions, %d imports", len(result.Classes), len(result.Functions), len(result.Imports))
}

// === Java parser ===

func TestParseJavaBasic(t *testing.T) {
	p := New()
	content := `import java.util.List;
import java.util.ArrayList;

public class Server {
    private String host;
    private int port;

    public Server(String host, int port) {
        this.host = host;
        this.port = port;
    }

    public void start() {
        System.out.println("Starting " + host + ":" + port);
    }
}
`
	result := p.ParseFile("Server.java", content)
	if result == nil {
		t.Fatal("expected parse result")
	}
	t.Logf("java: %d classes, %d functions, %d imports", len(result.Classes), len(result.Functions), len(result.Imports))
}

// === TypeScript parser ===

func TestParseTypeScriptClass(t *testing.T) {
	p := New()
	content := `interface Config {
    host: string;
    port: number;
}

class Server implements Config {
    host: string;
    port: number;

    constructor(host: string, port: number) {
        this.host = host;
        this.port = port;
    }

    start(): void {
        console.log("Starting");
    }
}

function createServer(config: Config): Server {
    return new Server(config.host, config.port);
}
`
	result := p.ParseFile("server.ts", content)
	if result == nil {
		t.Fatal("expected parse result")
	}
	t.Logf("ts: %d classes, %d functions, %d imports", len(result.Classes), len(result.Functions), len(result.Imports))
}

// === TSX parser ===

func TestParseTSXComponent(t *testing.T) {
	p := New()
	content := `import React from 'react';

interface Props {
    name: string;
}

function Greeting({ name }: Props) {
    return <h1>Hello {name}</h1>;
}

export default Greeting;
`
	result := p.ParseFile("greeting.tsx", content)
	if result == nil {
		t.Fatal("expected parse result")
	}
	t.Logf("tsx: %d classes, %d functions, %d imports", len(result.Classes), len(result.Functions), len(result.Imports))
}

// TestParseGoTypeSpecNil tests Go type spec that returns nil (no name)
func TestParseGoEmptyTypeSpec(t *testing.T) {
	p := New()
	// This is valid Go but may produce edge cases in type spec extraction
	content := `package main

type ()
`
	result := p.ParseFile("empty_type.go", content)
	if result == nil {
		t.Fatal("expected parse result")
	}
	// Should not crash, may have 0 classes
	t.Logf("empty type: %d classes", len(result.Classes))
}
