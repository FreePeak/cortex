package domain

import (
	"testing"
)

func TestDomainError(t *testing.T) {
	tests := []struct {
		name    string
		message string
		code    int
	}{
		{
			name:    "Not found error",
			message: "not found",
			code:    404,
		},
		{
			name:    "Invalid input error",
			message: "invalid input",
			code:    400,
		},
		{
			name:    "Internal server error",
			message: "internal server error",
			code:    500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewError(tt.message, tt.code)

			if err.Message != tt.message {
				t.Errorf("NewError().Message = %v, want %v", err.Message, tt.message)
			}
			if err.Code != tt.code {
				t.Errorf("NewError().Code = %v, want %v", err.Code, tt.code)
			}
			if err.Error() != tt.message {
				t.Errorf("NewError().Error() = %v, want %v", err.Error(), tt.message)
			}
		})
	}
}

func TestResourceNotFoundError(t *testing.T) {
	uri := "test/resource"
	err := NewResourceNotFoundError(uri)

	if err.URI != uri {
		t.Errorf("NewResourceNotFoundError().URI = %v, want %v", err.URI, uri)
	}
	if err.Err == nil {
		t.Error("NewResourceNotFoundError().Err should not be nil")
	}
	if err.Err.Code != 404 {
		t.Errorf("NewResourceNotFoundError().Err.Code = %v, want 404", err.Err.Code)
	}
	if err.Error() == "" {
		t.Error("NewResourceNotFoundError().Error() should not return empty string")
	}
}

func TestToolNotFoundError(t *testing.T) {
	name := "test-tool"
	err := NewToolNotFoundError(name)

	if err.Name != name {
		t.Errorf("NewToolNotFoundError().Name = %v, want %v", err.Name, name)
	}
	if err.Err == nil {
		t.Error("NewToolNotFoundError().Err should not be nil")
	}
	if err.Err.Code != 404 {
		t.Errorf("NewToolNotFoundError().Err.Code = %v, want 404", err.Err.Code)
	}
	if err.Error() == "" {
		t.Error("NewToolNotFoundError().Error() should not return empty string")
	}
}

func TestPromptNotFoundError(t *testing.T) {
	name := "test-prompt"
	err := NewPromptNotFoundError(name)

	if err.Name != name {
		t.Errorf("NewPromptNotFoundError().Name = %v, want %v", err.Name, name)
	}
	if err.Err == nil {
		t.Error("NewPromptNotFoundError().Err should not be nil")
	}
	if err.Err.Code != 404 {
		t.Errorf("NewPromptNotFoundError().Err.Code = %v, want 404", err.Err.Code)
	}
	if err.Error() == "" {
		t.Error("NewPromptNotFoundError().Error() should not return empty string")
	}
}

func TestSessionNotFoundError(t *testing.T) {
	id := "test-session-id"
	err := NewSessionNotFoundError(id)

	if err.ID != id {
		t.Errorf("NewSessionNotFoundError().ID = %v, want %v", err.ID, id)
	}
	if err.Err == nil {
		t.Error("NewSessionNotFoundError().Err should not be nil")
	}
	if err.Err.Code != 404 {
		t.Errorf("NewSessionNotFoundError().Err.Code = %v, want 404", err.Err.Code)
	}
	if err.Error() == "" {
		t.Error("NewSessionNotFoundError().Error() should not return empty string")
	}
}

func TestValidationError(t *testing.T) {
	field := "name"
	message := "must not be empty"
	err := NewValidationError(field, message)

	if err.Field != field {
		t.Errorf("NewValidationError().Field = %v, want %v", err.Field, field)
	}
	if err.Message != message {
		t.Errorf("NewValidationError().Message = %v, want %v", err.Message, message)
	}
	if err.Err == nil {
		t.Error("NewValidationError().Err should not be nil")
	}
	if err.Err.Code != 400 {
		t.Errorf("NewValidationError().Err.Code = %v, want 400", err.Err.Code)
	}
	if err.Error() == "" {
		t.Error("NewValidationError().Error() should not return empty string")
	}
}
