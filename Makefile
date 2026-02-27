BINARY_NAME := fastcode

VERSION ?= 0.1.0-dev
BUILD_TIME := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS := -X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME) -X main.gitCommit=$(GIT_COMMIT)

.PHONY: build install uninstall test test-e2e clean help

## Build the binary in the project directory
build:
	go build -ldflags="$(LDFLAGS)" -o $(BINARY_NAME) cmd/fastcode/*.go

## Install globally via go install (binary goes to GOPATH/bin)
install:
	go install -ldflags="$(LDFLAGS)" ./cmd/fastcode/
	@echo "✅ Installed $(BINARY_NAME) to $$(go env GOPATH)/bin/$(BINARY_NAME)"
	@echo ""
	@echo "👉 Make sure GOPATH/bin is in your PATH:"
	@echo '   echo '"'"'export PATH="$$GOPATH/bin:$$PATH"'"'"' >> ~/.zshrc && source ~/.zshrc'

## Remove the installed binary
uninstall:
	rm -f $$(go env GOPATH)/bin/$(BINARY_NAME)
	@echo "🗑  Removed $(BINARY_NAME) from $$(go env GOPATH)/bin"

## Run all tests
test:
	go test ./... -count=1 -v

## Run E2E tests only (no API key needed)
test-e2e:
	go test ./internal/orchestrator/ -v -run TestE2E -count=1

## Run shell E2E (requires OPENAI_API_KEY)
test-e2e-full:
	chmod +x run_e2e.sh && ./run_e2e.sh

## Clean build artifacts
clean:
	rm -f $(BINARY_NAME)
	@echo "🧹 Cleaned build artifacts"

## Show available targets
help:
	@echo "make build        — Build binary locally"
	@echo "make install      — Install to GOPATH/bin (call fastcode from anywhere)"
	@echo "make uninstall    — Remove installed binary"
	@echo "make test         — Run all tests"
	@echo "make test-e2e     — Run E2E tests (offline, no API key)"
	@echo "make test-e2e-full— Run shell E2E (needs OPENAI_API_KEY)"
	@echo "make clean        — Remove local binary"
