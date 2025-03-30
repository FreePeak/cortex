package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/FreePeak/cortex/pkg/server"
	"github.com/FreePeak/cortex/pkg/tools"
)

const (
	serverName       = "Example SSE MCP Server"
	serverVersion    = "1.0.0"
	serverAddr       = ":8080"
	shutdownTimeout  = 10 * time.Second
	shutdownGraceful = 2 * time.Second
)

func main() {
	// Create a new server using the SDK
	mcpServer := server.NewMCPServer(serverName, serverVersion)

	// Set the server address
	mcpServer.SetAddress(serverAddr)

	// Create tools with the fluent API
	echoTool := tools.NewTool("echo",
		tools.WithDescription("Echoes back the input message"),
		tools.WithString("message",
			tools.Description("The message to echo back"),
			tools.Required(),
		),
	)

	// Create the weather tool
	weatherTool := tools.NewTool("weather",
		tools.WithDescription("Gets today's weather forecast"),
		tools.WithString("location",
			tools.Description("The location to get weather for"),
			tools.Required(),
		),
	)

	// Add tools with handlers
	ctx := context.Background()
	err := mcpServer.AddTool(ctx, echoTool, handleEcho)
	if err != nil {
		log.Fatalf("Error adding echo tool: %v", err)
	}

	err = mcpServer.AddTool(ctx, weatherTool, handleWeather)
	if err != nil {
		log.Fatalf("Error adding weather tool: %v", err)
	}

	// Handle graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		fmt.Printf("Server is running on %s\n", serverAddr)
		fmt.Printf("You can connect to this server from Cursor by going to Settings > Extensions > Model Context Protocol and entering 'http://localhost%s' as the server URL.\n", serverAddr)
		fmt.Println("Available tools: echo, weather")
		fmt.Println("Press Ctrl+C to stop")

		// Use the SDK's built-in HTTP server functionality
		if err := mcpServer.ServeHTTP(); err != nil {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for shutdown signal
	<-shutdown
	fmt.Println("Shutting down server...")

	// Create a context with timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	// Shutdown server
	if err := mcpServer.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	// Small delay to allow final cleanup
	time.Sleep(shutdownGraceful)
	fmt.Println("Server stopped gracefully")
}

// Echo tool handler
func handleEcho(ctx context.Context, request server.ToolCallRequest) (interface{}, error) {
	// Extract the message parameter
	message, ok := request.Parameters["message"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid 'message' parameter")
	}

	// Return the echo response in the format expected by the MCP protocol
	return map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": message,
			},
		},
	}, nil
}

// Weather tool handler
func handleWeather(ctx context.Context, request server.ToolCallRequest) (interface{}, error) {
	// Extract the location parameter
	location, ok := request.Parameters["location"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid 'location' parameter")
	}

	// Generate random weather data for testing
	conditions := []string{"Sunny", "Partly Cloudy", "Cloudy", "Rainy", "Thunderstorms", "Snowy", "Foggy", "Windy"}
	tempF := rand.Intn(50) + 30 // Random temperature between 30°F and 80°F
	tempC := (tempF - 32) * 5 / 9
	humidity := rand.Intn(60) + 30 // Random humidity between 30% and 90%
	windSpeed := rand.Intn(20) + 5 // Random wind speed between 5-25mph

	// Select a random condition
	condition := conditions[rand.Intn(len(conditions))]

	// Format today's date
	today := time.Now().Format("Monday, January 2, 2006")

	// Format the weather response
	weatherInfo := fmt.Sprintf("Weather for %s on %s:\n"+
		"Condition: %s\n"+
		"Temperature: %d°F (%d°C)\n"+
		"Humidity: %d%%\n"+
		"Wind Speed: %d mph",
		location, today, condition, tempF, tempC, humidity, windSpeed)

	// Return the weather response in the format expected by the MCP protocol
	return map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": weatherInfo,
			},
		},
	}, nil
}
