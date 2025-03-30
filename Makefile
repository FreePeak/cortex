# Makefile for cortex
.PHONY: build test test-race lint clean run example-sse example-stdio example-multi coverage deps update-deps help

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Example paths
STDIO_SERVER=examples/stdio-server/main.go
SSE_SERVER=examples/sse-server/main.go
MULTI_PROTOCOL_SERVER=examples/multi-protocol/main.go

# Binary paths
BIN_DIR=bin
STDIO_BIN=$(BIN_DIR)/stdio-server
SSE_BIN=$(BIN_DIR)/sse-server
MULTI_BIN=$(BIN_DIR)/multi-protocol-server

help:
	@echo "Available commands:"
	@echo "  make              - Run tests and build binaries"
	@echo "  make build        - Build the server binaries"
	@echo "  make test         - Run tests with race detection and coverage"
	@echo "  make test-race    - Run tests with race detection and coverage"
	@echo "  make coverage     - Generate test coverage report"
	@echo "  make lint         - Run linter"
	@echo "  make clean        - Clean build artifacts"
	@echo "  make deps         - Tidy up dependencies"
	@echo "  make update-deps  - Update dependencies"
	@echo "  make example-sse  - Run example SSE server"
	@echo "  make example-stdio - Run example stdio server"
	@echo "  make example-multi - Run example multi-protocol server"

all: test build lint

build: $(BIN_DIR)
	$(GOBUILD) -o $(STDIO_BIN) $(STDIO_SERVER)
	$(GOBUILD) -o $(SSE_BIN) $(SSE_SERVER)
	$(GOBUILD) -o $(MULTI_BIN) $(MULTI_PROTOCOL_SERVER)

$(BIN_DIR):
	mkdir -p $(BIN_DIR)

example-sse:
	$(GOCMD) run $(SSE_SERVER)

example-stdio:
	$(GOCMD) run $(STDIO_SERVER)

example-multi:
	$(GOCMD) run $(MULTI_PROTOCOL_SERVER) -protocol stdio

example-multi-http:
	$(GOCMD) run $(MULTI_PROTOCOL_SERVER) -protocol http -address localhost:8080

test:
	$(GOTEST) ./... -v -race -cover

test-race:
	$(GOTEST) ./... -v -race -cover

coverage:
	$(GOTEST) -cover -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out

lint:
	golangci-lint run ./...

clean:
	$(GOCLEAN)
	rm -rf $(BIN_DIR)
	rm -f coverage.out

deps:
	$(GOMOD) tidy

update-deps:
	$(GOMOD) tidy
	$(GOGET) -u ./... 