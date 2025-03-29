package domain

import (
	"encoding/json"
	"testing"
)

func TestNotificationToJSONRPC(t *testing.T) {
	tests := []struct {
		name           string
		notification   *Notification
		jsonrpcVersion string
		want           JSONRPCNotification
	}{
		{
			name: "Basic notification",
			notification: &Notification{
				Method: "test.notification",
				Params: map[string]interface{}{
					"key": "value",
				},
			},
			jsonrpcVersion: "2.0",
			want: JSONRPCNotification{
				JSONRPC: "2.0",
				Method:  "test.notification",
				Params: map[string]interface{}{
					"key": "value",
				},
			},
		},
		{
			name: "Empty params",
			notification: &Notification{
				Method: "test.empty",
				Params: map[string]interface{}{},
			},
			jsonrpcVersion: "2.0",
			want: JSONRPCNotification{
				JSONRPC: "2.0",
				Method:  "test.empty",
				Params:  map[string]interface{}{},
			},
		},
		{
			name: "Complex params",
			notification: &Notification{
				Method: "test.complex",
				Params: map[string]interface{}{
					"string":  "text",
					"number":  123,
					"boolean": true,
					"array":   []interface{}{1, 2, 3},
					"object": map[string]interface{}{
						"nested": "value",
					},
				},
			},
			jsonrpcVersion: "2.0",
			want: JSONRPCNotification{
				JSONRPC: "2.0",
				Method:  "test.complex",
				Params: map[string]interface{}{
					"string":  "text",
					"number":  123,
					"boolean": true,
					"array":   []interface{}{1, 2, 3},
					"object": map[string]interface{}{
						"nested": "value",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.notification.ToJSONRPC(tt.jsonrpcVersion)

			// Check JSONRPC version
			if got.JSONRPC != tt.want.JSONRPC {
				t.Errorf("ToJSONRPC().JSONRPC = %v, want %v", got.JSONRPC, tt.want.JSONRPC)
			}

			// Check Method
			if got.Method != tt.want.Method {
				t.Errorf("ToJSONRPC().Method = %v, want %v", got.Method, tt.want.Method)
			}

			// Compare Params using JSON marshaling for deep equality
			gotJSON, err := json.Marshal(got.Params)
			if err != nil {
				t.Errorf("Failed to marshal got.Params: %v", err)
			}

			wantJSON, err := json.Marshal(tt.want.Params)
			if err != nil {
				t.Errorf("Failed to marshal want.Params: %v", err)
			}

			if string(gotJSON) != string(wantJSON) {
				t.Errorf("ToJSONRPC().Params = %v, want %v", string(gotJSON), string(wantJSON))
			}
		})
	}
}

func TestJSONRPCNotificationMarshal(t *testing.T) {
	notification := JSONRPCNotification{
		JSONRPC: "2.0",
		Method:  "test.notification",
		Params: map[string]interface{}{
			"key":    "value",
			"number": 42,
		},
	}

	// Marshal to JSON
	jsonBytes, err := json.Marshal(notification)
	if err != nil {
		t.Fatalf("Failed to marshal JSONRPCNotification: %v", err)
	}

	// Unmarshal back to verify JSON structure
	var unmarshaled map[string]interface{}
	err = json.Unmarshal(jsonBytes, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// Check fields
	if unmarshaled["jsonrpc"] != "2.0" {
		t.Errorf("jsonrpc = %v, want %v", unmarshaled["jsonrpc"], "2.0")
	}

	if unmarshaled["method"] != "test.notification" {
		t.Errorf("method = %v, want %v", unmarshaled["method"], "test.notification")
	}

	params, ok := unmarshaled["params"].(map[string]interface{})
	if !ok {
		t.Fatalf("params is not a map[string]interface{}, got %T", unmarshaled["params"])
	}

	if params["key"] != "value" {
		t.Errorf("params.key = %v, want %v", params["key"], "value")
	}

	// JSON numbers are float64 when unmarshaled
	if params["number"] != float64(42) {
		t.Errorf("params.number = %v, want %v", params["number"], float64(42))
	}
}
