package domain

import (
	"testing"

	"github.com/google/uuid"
)

func TestNewClientSession(t *testing.T) {
	tests := []struct {
		name      string
		userAgent string
		want      *ClientSession
	}{
		{
			name:      "Valid user agent",
			userAgent: "test-agent",
			want: &ClientSession{
				UserAgent: "test-agent",
				Connected: true,
			},
		},
		{
			name:      "Empty user agent",
			userAgent: "",
			want: &ClientSession{
				UserAgent: "",
				Connected: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewClientSession(tt.userAgent)

			// Verify user agent and connected status
			if got.UserAgent != tt.want.UserAgent {
				t.Errorf("NewClientSession().UserAgent = %v, want %v", got.UserAgent, tt.want.UserAgent)
			}
			if got.Connected != tt.want.Connected {
				t.Errorf("NewClientSession().Connected = %v, want %v", got.Connected, tt.want.Connected)
			}

			// Check if ID is a valid UUID
			_, err := uuid.Parse(got.ID)
			if err != nil {
				t.Errorf("NewClientSession().ID is not a valid UUID: %v", err)
			}
		})
	}
}

func TestResourceContents(t *testing.T) {
	tests := []struct {
		name     string
		contents ResourceContents
	}{
		{
			name: "Text content",
			contents: ResourceContents{
				URI:      "test/uri.txt",
				MIMEType: "text/plain",
				Content:  []byte("Hello, World!"),
				Text:     "Hello, World!",
			},
		},
		{
			name: "Binary content",
			contents: ResourceContents{
				URI:      "test/image.png",
				MIMEType: "image/png",
				Content:  []byte{0x89, 0x50, 0x4E, 0x47},
				Text:     "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify fields are correctly set
			if tt.contents.URI == "" {
				t.Error("ResourceContents.URI should not be empty")
			}
			if tt.contents.MIMEType == "" {
				t.Error("ResourceContents.MIMEType should not be empty")
			}
			if len(tt.contents.Content) == 0 {
				t.Error("ResourceContents.Content should not be empty")
			}
		})
	}
}

func TestNotification(t *testing.T) {
	tests := []struct {
		name         string
		notification Notification
	}{
		{
			name: "Empty params",
			notification: Notification{
				Method: "test/method",
				Params: map[string]interface{}{},
			},
		},
		{
			name: "With params",
			notification: Notification{
				Method: "test/method",
				Params: map[string]interface{}{
					"key":     "value",
					"numeric": 42,
					"boolean": true,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify fields are correctly set
			if tt.notification.Method == "" {
				t.Error("Notification.Method should not be empty")
			}
			if tt.notification.Params == nil {
				t.Error("Notification.Params should not be nil")
			}
		})
	}
}
