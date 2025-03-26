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
EXAMPLE_SERVER=cmd/example/main.go

all: test build

build:
	$(GOBUILD) -o bin/$(BINARY_NAME) $(EXAMPLE_SERVER)

example:
	$(GOCMD) run $(EXAMPLE_SERVER)

test:
	$(GOTEST) -v ./...

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