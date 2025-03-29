# Makefile for golang-mcp-server-sdk
.PHONY: build test lint clean run example

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

all: test build

build:
	$(GOBUILD) -o $(STDIO_BIN_SERVER) $(ECHO_STDIO_SERVER)
	$(GOBUILD) -o $(SSE_BIN_SERVER) $(ECHO_SSE_SERVER)

example:
	$(GOCMD) run $(ECHO_SSE_SERVER)

example-stdio:
	cd echo-stdio-test && go run main.go

test:
	$(GOTEST) ./... -cover

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