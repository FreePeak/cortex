# Makefile for golang-mcp-server-sdk
.PHONY: build test test-race lint clean run example coverage deps update-deps help

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=mcp-server
STDIO_BIN_SERVER=bin/echo-stdio-server
SSE_BIN_SERVER=bin/echo-sse-server
ECHO_SSE_SERVER=cmd/echo-sse-server/main.go
ECHO_STDIO_SERVER=cmd/echo-stdio-server/main.go

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
	@echo "  make example      - Run example SSE server"
	@echo "  make example-stdio - Run example stdio server"

all: test build

build:
	$(GOBUILD) -o $(STDIO_BIN_SERVER) $(ECHO_STDIO_SERVER)
	$(GOBUILD) -o $(SSE_BIN_SERVER) $(ECHO_SSE_SERVER)

example:
	$(GOCMD) run $(ECHO_SSE_SERVER)

example-stdio:
	cd echo-stdio-test && go run main.go

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
	rm -f bin/$(BINARY_NAME)
	rm -f coverage.out

deps:
	$(GOMOD) tidy

update-deps:
	$(GOMOD) tidy
	$(GOGET) -u ./... 