package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/FreePeak/cortex/examples/providers/database"
	"github.com/FreePeak/cortex/examples/providers/weather"
	"github.com/FreePeak/cortex/pkg/server"
)

func main() {
	// Parse command-line flags
	protocol := flag.String("protocol", "stdio", "Communication protocol (stdio or http)")
	address := flag.String("address", "localhost:8080", "HTTP server address (when using http protocol)")
	flag.Parse()

	// Create a logger
	logger := log.New(os.Stdout, "[cortex] ", log.LstdFlags)

	// Create the MCP server
	mcpServer := server.NewMCPServer("Multi-Protocol Server", "1.0.0", logger)

	// Set HTTP address if using HTTP protocol
	if *protocol == "http" {
		mcpServer.SetAddress(*address)
	}

	// Create the weather provider
	weatherProvider, err := weather.NewWeatherProvider(logger)
	if err != nil {
		logger.Fatalf("Failed to create weather provider: %v", err)
	}

	// Create the database provider
	dbProvider, err := database.NewDBProvider(logger)
	if err != nil {
		logger.Fatalf("Failed to create database provider: %v", err)
	}

	// Register providers with the server
	ctx := context.Background()
	err = mcpServer.RegisterProvider(ctx, weatherProvider)
	if err != nil {
		logger.Fatalf("Failed to register weather provider: %v", err)
	}

	err = mcpServer.RegisterProvider(ctx, dbProvider)
	if err != nil {
		logger.Fatalf("Failed to register database provider: %v", err)
	}

	// Set up graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Start the server based on the specified protocol
	if *protocol == "stdio" {
		// Print server ready message for stdio
		fmt.Println("Server ready. You can now send JSON-RPC requests via stdin.")
		fmt.Println("Example weather tool request:")
		fmt.Println(`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"weather","parameters":{"location":"New York"}}}`)
		fmt.Println("Example database set tool request:")
		fmt.Println(`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"db.set","parameters":{"key":"user1","value":{"name":"John","age":30}}}}`)

		// Start the stdio server (this will block until terminated)
		go func() {
			if err := mcpServer.ServeStdio(); err != nil {
				logger.Fatalf("Error serving stdio: %v", err)
			}
		}()
	} else if *protocol == "http" {
		// Print server ready message for HTTP
		fmt.Printf("HTTP server starting on %s\n", *address)
		fmt.Println("You can query tools using HTTP POST requests to /tools/call")

		// Start the HTTP server (this will block until terminated)
		go func() {
			if err := mcpServer.ServeHTTP(); err != nil {
				logger.Fatalf("Error serving HTTP: %v", err)
			}
		}()
	} else {
		logger.Fatalf("Unknown protocol: %s (must be 'stdio' or 'http')", *protocol)
	}

	// Wait for shutdown signal
	<-stop
	logger.Println("Shutting down server...")

	// Shutdown the server gracefully
	if *protocol == "http" {
		shutdownCtx := context.Background()
		if err := mcpServer.Shutdown(shutdownCtx); err != nil {
			logger.Fatalf("Error shutting down server: %v", err)
		}
	}

	logger.Println("Server stopped")
}
