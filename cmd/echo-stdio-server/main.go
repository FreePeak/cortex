package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/FreePeak/cortex/pkg/server"
	"github.com/FreePeak/cortex/pkg/tools"
)

// Record a timestamp for demo purposes
func getTimestamp() string {
	return fmt.Sprintf("%d", time.Now().Unix())
}

func main() {
	// Create the server with name and version
	mcpServer := server.NewMCPServer("Echo Stdio Server", "1.0.0")

	// Create the echo tool using the fluent API
	// When registered, the name will automatically be prefixed with "cortex_"
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

	// Add the tools with handler functions
	ctx := context.Background()
	err := mcpServer.AddTool(ctx, echoTool, handleEcho)
	if err != nil {
		log.Fatalf("Error adding echo tool: %v", err)
	}

	err = mcpServer.AddTool(ctx, weatherTool, handleWeather)
	if err != nil {
		log.Fatalf("Error adding weather tool: %v", err)
	}

	// Print server ready message
	fmt.Println("Server ready. You can now send JSON-RPC requests via stdin.")
	fmt.Println("Call the echo tool using:")
	fmt.Println("{\"jsonrpc\":\"2.0\",\"id\":1,\"method\":\"tools/call\",\"params\":{\"name\":\"cortex_echo\",\"parameters\":{\"message\":\"Hello, World!\"}}}")
	fmt.Println("Call the weather tool using:")
	fmt.Println("{\"jsonrpc\":\"2.0\",\"id\":1,\"method\":\"tools/call\",\"params\":{\"name\":\"cortex_weather\",\"parameters\":{\"location\":\"New York\"}}}")

	// Start the server
	if err := mcpServer.ServeStdio(); err != nil {
		fmt.Fprintf(os.Stderr, "Error serving stdio: %v\n", err)
		os.Exit(1)
	}
}

// Echo tool handler
func handleEcho(ctx context.Context, request server.ToolCallRequest) (interface{}, error) {
	// Extract the message parameter
	message, ok := request.Parameters["message"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid 'message' parameter")
	}

	// Add a timestamp to show we can process the message
	timestamp := getTimestamp()
	responseMessage := fmt.Sprintf("[%s] %s", timestamp, message)

	// Return the echo response in the format expected by the MCP protocol
	return map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": responseMessage,
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
