# Development Guide

This guide provides instructions on how to build, install, and start using `fastcode-cli` from source.

## 🚀 Quick Start

### Install from Source

```bash
git clone https://github.com/duyhunghd6/fastcode-cli.git
cd fastcode-cli
go build -o fastcode ./cmd/fastcode

# Configure your LLM endpoint
export OPENAI_API_KEY="your-key"
export MODEL="gpt-4o"
export BASE_URL="https://api.openai.com/v1"
```

### Usage

```bash
# Index a local repository
fastcode index /path/to/your/repo

# Query the indexed codebase
fastcode query "How does the authentication flow work?"

# Multi-repo query
fastcode query --repos /path/repo1,/path/repo2 "Where is the payment logic?"

# Start as MCP server (for Cursor / Claude Code)
fastcode serve-mcp --port 8080
```

---

## 🔨 Build and Install

We provide a `Makefile` to simplify building, installing, and testing the project.

### Building the Project

To compile the project and generate the binary locally in the project directory:

```bash
make build
# or explicitly:
go build -o fastcode cmd/fastcode/*.go
```

### Installation

To install the `fastcode` binary globally (into your `GOPATH/bin`), run:

```bash
make install
# or explicitly:
go install ./cmd/fastcode/
```

Ensure your `GOPATH/bin` is in your system's `PATH`:

```bash
echo 'export PATH="$GOPATH/bin:$PATH"' >> ~/.zshrc && source ~/.zshrc
```

### Uninstallation

To remove the globally installed binary:

```bash
make uninstall
```

### Testing

- **Run all standard tests**:
  ```bash
  make test
  ```
- **Run E2E tests** (offline, no API key needed):
  ```bash
  make test-e2e
  ```
- **Run full shell E2E tests** (requires `OPENAI_API_KEY`):
  ```bash
  make test-e2e-full
  ```

### Cleaning Up

To remove locally built binaries and artifacts:

```bash
make clean
```
