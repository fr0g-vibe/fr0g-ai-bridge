package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/fr0g-vibe/fr0g-ai-bridge/internal/models"
)

func TestOpenWebUIClient_ChatCompletion(t *testing.T) {
	// Mock response
	mockResponse := models.ChatCompletionResponse{
		ID:      "test-id",
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   "test-model",
		Choices: []models.Choice{
			{
				Index: 0,
				Message: models.ChatMessage{
					Role:    "assistant",
					Content: "Hello! How can I help you?",
				},
				FinishReason: "stop",
			},
		},
		Usage: models.Usage{
			PromptTokens:     10,
			CompletionTokens: 8,
			TotalTokens:      18,
		},
	}

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/chat/completions" {
			t.Errorf("expected path /api/chat/completions, got %s", r.URL.Path)
		}

		if r.Method != "POST" {
			t.Errorf("expected POST method, got %s", r.Method)
		}

		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	// Create client
	client := NewOpenWebUIClient(server.URL, "test-api-key", 30*time.Second)

	// Test request
	req := &models.ChatCompletionRequest{
		Model: "test-model",
		Messages: []models.ChatMessage{
			{Role: "user", Content: "Hello"},
		},
	}

	resp, err := client.ChatCompletion(context.Background(), req)
	if err != nil {
		t.Fatalf("ChatCompletion failed: %v", err)
	}

	if resp.ID != mockResponse.ID {
		t.Errorf("expected ID %s, got %s", mockResponse.ID, resp.ID)
	}

	if resp.Model != mockResponse.Model {
		t.Errorf("expected model %s, got %s", mockResponse.Model, resp.Model)
	}

	if len(resp.Choices) != 1 {
		t.Errorf("expected 1 choice, got %d", len(resp.Choices))
	}
}

func TestOpenWebUIClient_HealthCheck(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		expectedError  bool
	}{
		{
			name:          "healthy",
			statusCode:    http.StatusOK,
			expectedError: false,
		},
		{
			name:          "unhealthy",
			statusCode:    http.StatusInternalServerError,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/models" {
					t.Errorf("expected path /api/models, got %s", r.URL.Path)
				}

				w.WriteHeader(tt.statusCode)
			}))
			defer server.Close()

			client := NewOpenWebUIClient(server.URL, "", 30*time.Second)

			err := client.HealthCheck(context.Background())
			if (err != nil) != tt.expectedError {
				t.Errorf("HealthCheck() error = %v, expectedError %v", err, tt.expectedError)
			}
		})
	}
}

func TestOpenWebUIClient_PrepareOpenWebUIRequest(t *testing.T) {
	client := &OpenWebUIClient{}

	tests := []struct {
		name     string
		request  *models.ChatCompletionRequest
		expected int // expected number of messages
	}{
		{
			name: "no persona prompt",
			request: &models.ChatCompletionRequest{
				Model: "test-model",
				Messages: []models.ChatMessage{
					{Role: "user", Content: "Hello"},
				},
			},
			expected: 1,
		},
		{
			name: "with persona prompt, no existing system message",
			request: &models.ChatCompletionRequest{
				Model: "test-model",
				Messages: []models.ChatMessage{
					{Role: "user", Content: "Hello"},
				},
				PersonaPrompt: "You are a helpful assistant",
			},
			expected: 2, // system message + user message
		},
		{
			name: "with persona prompt, existing system message",
			request: &models.ChatCompletionRequest{
				Model: "test-model",
				Messages: []models.ChatMessage{
					{Role: "system", Content: "You are an AI"},
					{Role: "user", Content: "Hello"},
				},
				PersonaPrompt: "You are a helpful assistant",
			},
			expected: 2, // modified system message + user message
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := client.prepareOpenWebUIRequest(tt.request)

			if len(result.Messages) != tt.expected {
				t.Errorf("expected %d messages, got %d", tt.expected, len(result.Messages))
			}

			if result.PersonaPrompt != "" {
				t.Errorf("expected PersonaPrompt to be cleared, got %s", result.PersonaPrompt)
			}

			// Check if persona prompt was added to system message
			if tt.request.PersonaPrompt != "" {
				found := false
				for _, msg := range result.Messages {
					if msg.Role == "system" && msg.Content != "" {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected system message with persona prompt")
				}
			}
		})
	}
}
