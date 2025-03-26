package main

import (
	"log"
	"os"

	"github.com/FreePeak/golang-mcp-server-sdk/internal/interfaces/rest"
	"github.com/FreePeak/golang-mcp-server-sdk/internal/interfaces/stdio"
	"github.com/FreePeak/golang-mcp-server-sdk/internal/usecases"
)

func main() {
	// Create server service
	serviceConfig := usecases.ServerConfig{
		Name:         "GoMCP SDK",
		Version:      "0.1.0",
		Instructions: "Welcome to the GoMCP SDK stdio server!",
	}
	service := usecases.NewServerService(serviceConfig)

	// Create MCP server with a dummy address since we won't be using the HTTP server
	server := rest.NewMCPServer(service, ":0")

	// Set up the stdio server
	logger := log.New(os.Stderr, "[STDIO-MCP] ", log.LstdFlags)
	err := stdio.ServeStdio(server, stdio.WithErrorLogger(logger))
	if err != nil {
		logger.Printf("Error serving stdio: %v", err)
		os.Exit(1)
	}
}
