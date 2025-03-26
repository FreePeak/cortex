package echo_client

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// JSONRPCRequest represents a JSON-RPC request.
type JSONRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// JSONRPCResponse represents a JSON-RPC response.
type JSONRPCResponse struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      interface{}   `json:"id"`
	Result  interface{}   `json:"result,omitempty"`
	Error   *JSONRPCError `json:"error,omitempty"`
}

// JSONRPCError represents a JSON-RPC error.
type JSONRPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func main() {
	// Parse command line flags
	serverURL := flag.String("server", "http://localhost:8080/jsonrpc", "MCP server URL")
	message := flag.String("message", "Hello from echo client!", "Message to echo")
	flag.Parse()

	// Call the echo tool
	result, err := callEchoTool(*serverURL, *message)
	if err != nil {
		log.Fatalf("Failed to call echo tool: %v", err)
	}

	// Extract the echoed message
	resultMap, ok := result["result"].(map[string]interface{})
	if !ok {
		log.Fatalf("Unexpected result format: %v", result)
	}

	echoedMessage, ok := resultMap["message"].(string)
	if !ok {
		log.Fatalf("Echoed message not found in response: %v", resultMap)
	}

	// Print the result
	fmt.Printf("Server echoed: %s\n", echoedMessage)
}

func callEchoTool(serverURL, message string) (map[string]interface{}, error) {
	// Create tools/call request
	echoReq := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name": "echo",
			"parameters": map[string]interface{}{
				"message": message,
			},
		},
	}

	// Marshal request to JSON
	reqBytes, err := json.Marshal(echoReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", serverURL, bytes.NewBuffer(reqBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	// Send request
	httpResp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer httpResp.Body.Close()

	// Read response body
	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", httpResp.StatusCode, string(body))
	}

	// Unmarshal response
	var resp JSONRPCResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w, body: %s", err, string(body))
	}

	// Check for errors
	if resp.Error != nil {
		return nil, fmt.Errorf("RPC error: %s (code: %d)", resp.Error.Message, resp.Error.Code)
	}

	// Return the result as a map
	result, ok := resp.Result.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected result type: %T", resp.Result)
	}

	return result, nil
}
