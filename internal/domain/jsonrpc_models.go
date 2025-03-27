package domain

// JSONRPCRequest represents a JSON-RPC request in the domain layer.
type JSONRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// JSONRPCResponse represents a JSON-RPC response in the domain layer.
type JSONRPCResponse struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      interface{}   `json:"id"`
	Result  interface{}   `json:"result,omitempty"`
	Error   *JSONRPCError `json:"error,omitempty"`
}

// JSONRPCError represents a JSON-RPC error in the domain layer.
type JSONRPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// CreateResponse creates a new JSONRPCResponse with the given ID and result.
func CreateResponse(jsonrpcVersion string, id interface{}, result interface{}) JSONRPCResponse {
	return JSONRPCResponse{
		JSONRPC: jsonrpcVersion,
		ID:      id,
		Result:  result,
	}
}

// CreateErrorResponse creates a new JSONRPCResponse with the given ID and error.
func CreateErrorResponse(jsonrpcVersion string, id interface{}, code int, message string) JSONRPCResponse {
	return JSONRPCResponse{
		JSONRPC: jsonrpcVersion,
		ID:      id,
		Error: &JSONRPCError{
			Code:    code,
			Message: message,
		},
	}
}
