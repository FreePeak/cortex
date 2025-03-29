package domain

import (
	"testing"
)

func TestCreateResponse(t *testing.T) {
	// Test parameters
	jsonrpcVersion := "2.0"
	id := 123
	result := map[string]interface{}{
		"success": true,
		"data":    "test data",
	}

	// Create response
	response := CreateResponse(jsonrpcVersion, id, result)

	// Assertions
	if response.JSONRPC != jsonrpcVersion {
		t.Errorf("CreateResponse().JSONRPC = %v, want %v", response.JSONRPC, jsonrpcVersion)
	}
	if response.ID != id {
		t.Errorf("CreateResponse().ID = %v, want %v", response.ID, id)
	}
	if response.Result == nil {
		t.Error("CreateResponse().Result should not be nil")
	}
	if response.Error != nil {
		t.Error("CreateResponse().Error should be nil")
	}

	// Type assertion and validation of result
	resultMap, ok := response.Result.(map[string]interface{})
	if !ok {
		t.Errorf("CreateResponse().Result should be a map[string]interface{}, got %T", response.Result)
	} else {
		if resultMap["success"] != result["success"] {
			t.Errorf("CreateResponse().Result[\"success\"] = %v, want %v", resultMap["success"], result["success"])
		}
		if resultMap["data"] != result["data"] {
			t.Errorf("CreateResponse().Result[\"data\"] = %v, want %v", resultMap["data"], result["data"])
		}
	}
}

func TestCreateErrorResponse(t *testing.T) {
	// Test parameters
	jsonrpcVersion := "2.0"
	id := "abc123"
	code := -32600
	message := "Invalid Request"

	// Create error response
	response := CreateErrorResponse(jsonrpcVersion, id, code, message)

	// Assertions
	if response.JSONRPC != jsonrpcVersion {
		t.Errorf("CreateErrorResponse().JSONRPC = %v, want %v", response.JSONRPC, jsonrpcVersion)
	}
	if response.ID != id {
		t.Errorf("CreateErrorResponse().ID = %v, want %v", response.ID, id)
	}
	if response.Result != nil {
		t.Error("CreateErrorResponse().Result should be nil")
	}
	if response.Error == nil {
		t.Error("CreateErrorResponse().Error should not be nil")
	} else {
		if response.Error.Code != code {
			t.Errorf("CreateErrorResponse().Error.Code = %v, want %v", response.Error.Code, code)
		}
		if response.Error.Message != message {
			t.Errorf("CreateErrorResponse().Error.Message = %v, want %v", response.Error.Message, message)
		}
	}
}

func TestJSONRPCRequest(t *testing.T) {
	// Create request
	request := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      456,
		Method:  "test.method",
		Params: map[string]interface{}{
			"param1": "value1",
			"param2": 42,
		},
	}

	// Assertions
	if request.JSONRPC != "2.0" {
		t.Errorf("JSONRPCRequest.JSONRPC = %v, want %v", request.JSONRPC, "2.0")
	}
	if request.ID != 456 {
		t.Errorf("JSONRPCRequest.ID = %v, want %v", request.ID, 456)
	}
	if request.Method != "test.method" {
		t.Errorf("JSONRPCRequest.Method = %v, want %v", request.Method, "test.method")
	}

	// Type assertion and validation of params
	paramsMap, ok := request.Params.(map[string]interface{})
	if !ok {
		t.Errorf("JSONRPCRequest.Params should be a map[string]interface{}, got %T", request.Params)
	} else {
		if paramsMap["param1"] != "value1" {
			t.Errorf("JSONRPCRequest.Params[\"param1\"] = %v, want %v", paramsMap["param1"], "value1")
		}
		if paramsMap["param2"] != 42 {
			t.Errorf("JSONRPCRequest.Params[\"param2\"] = %v, want %v", paramsMap["param2"], 42)
		}
	}
}
