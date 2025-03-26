package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

// JSONRPCRequest represents a JSON-RPC request.
type JSONRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// JSONRPCResponse represents a JSON-RPC response
type JSONRPCResponse struct {
	JSONRPC string                 `json:"jsonrpc"`
	ID      interface{}            `json:"id"`
	Result  map[string]interface{} `json:"result,omitempty"`
	Error   *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func main() {
	// Default to the echo-stdio-server if no argument is provided
	serverPath := "../echo-stdio-server/main.go"
	if len(os.Args) > 1 {
		serverPath = os.Args[1]
	}

	fmt.Printf("Testing server at path: %s\n", serverPath)

	// Command to start the echo stdio server
	cmd := exec.Command("go", "run", serverPath)

	// Get pipes for stdin/stdout
	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Fatalf("Failed to get stdin pipe: %v", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatalf("Failed to get stdout pipe: %v", err)
	}

	// Capture stderr for debugging
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		log.Fatalf("Failed to get stderr pipe: %v", err)
	}

	// Start the server
	if err := cmd.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	// Read server logs in background
	go func() {
		scanner := bufio.NewScanner(stderrPipe)
		for scanner.Scan() {
			log.Printf("SERVER: %s", scanner.Text())
		}
	}()

	// Wait for server to initialize
	time.Sleep(500 * time.Millisecond)

	// Reader for stdout
	reader := bufio.NewReader(stdout)

	// Initialize the server
	sendRequest(stdin, "initialize", map[string]interface{}{}, 1)
	initResp := readResponse(reader)
	fmt.Printf("Initialize response: %s\n", initResp)

	// Call the echo tool
	echoParams := map[string]interface{}{
		"name": "echo",
		"parameters": map[string]interface{}{
			"message": "Hello, MCP Server!",
		},
	}

	sendRequest(stdin, "tools/call", echoParams, 2)
	echoResp := readResponse(reader)
	fmt.Printf("Echo response: %s\n", echoResp)

	// Clean up
	if err := cmd.Process.Kill(); err != nil {
		log.Printf("Failed to kill server: %v", err)
	}

	cmd.Wait()
	fmt.Println("\nTest completed!")
}

// sendRequest sends a JSON-RPC request to the server
func sendRequest(writer io.Writer, method string, params interface{}, id int) {
	request := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
		Params:  params,
	}

	reqBytes, err := json.Marshal(request)
	if err != nil {
		log.Fatalf("Failed to marshal request: %v", err)
	}

	fmt.Printf("\nSending %s request: %s\n", method, string(reqBytes))
	_, err = writer.Write(append(reqBytes, '\n'))
	if err != nil {
		log.Fatalf("Failed to write request: %v", err)
	}
}

// readResponse reads a full line from the reader and returns it
func readResponse(reader *bufio.Reader) string {
	response, err := reader.ReadString('\n')
	if err != nil {
		if err == io.EOF {
			log.Printf("Server closed connection")
			return ""
		}
		log.Fatalf("Failed to read response: %v", err)
	}

	return strings.TrimSpace(response)
}
