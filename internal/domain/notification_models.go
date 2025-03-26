package domain

// JSONRPCNotification represents a notification sent to clients via JSON-RPC.
type JSONRPCNotification struct {
	JSONRPC string                 `json:"jsonrpc"`
	Method  string                 `json:"method"`
	Params  map[string]interface{} `json:"params,omitempty"`
}

// ToJSONRPC converts a domain Notification to a JSONRPCNotification.
func (n *Notification) ToJSONRPC(jsonrpcVersion string) JSONRPCNotification {
	return JSONRPCNotification{
		JSONRPC: jsonrpcVersion,
		Method:  n.Method,
		Params:  n.Params,
	}
}
