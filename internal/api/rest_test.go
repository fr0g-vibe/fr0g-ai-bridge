package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/fr0g-vibe/fr0g-ai-bridge/internal/client"
	"github.com/fr0g-vibe/fr0g-ai-bridge/internal/models"
)

// mockOpenWebUIClient is a mock implementation of OpenWebUIClient for testing
type mockOpenWebUIClient struct {
	healthCheckError error
	chatResponse     *models.ChatCompletionResponse
	chatError        error
}

func (m *mockOpenWebUIClient) HealthCheck(ctx context.Context) error {
	return m.healthCheckError
}

func (m *mockOpenWebUIClient) ChatCompletion(ctx context.Context, req *models.ChatCompletionRequest) (*models.ChatCompletionResponse, error) {
	if m.chatError != nil {
		return nil, m.chatError
	}
	return m.chatResponse, nil
}

func TestRESTServer_HealthCheck(t *testing.T) {
	tests := []struct {
		name           string
		healthError    error
		expectedStatus int
		expectedHealth string
	}{
		{
			name:           "healthy",
			healthError:    nil,
			expectedStatus: http.StatusOK,
			expectedHealth: "healthy",
		},
		{
			name:           "unhealthy",
			healthError:    context.DeadlineExceeded,
			expectedStatus: http.StatusServiceUnavailable,
			expectedHealth: "unhealthy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockOpenWebUIClient{
				healthCheckError: tt.healthError,
			}

			server := NewRESTServer(mockClient)
			req := httptest.NewRequest("GET", "/health", nil)
			w := httptest.NewRecorder()

			server.handleHealth(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			var response models.HealthResponse
			if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			if response.Status != tt.expectedHealth {
				t.Errorf("expected status %s, got %s", tt.expectedHealth, response.Status)
			}

			if response.Version != "1.0.0" {
				t.Errorf("expected version 1.0.0, got %s", response.Version)
			}
		})
	}
}

func TestRESTServer_ChatCompletion(t *testing.T) {
	mockResponse := &models.ChatCompletionResponse{
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

	tests := []struct {
		name           string
		request        models.ChatCompletionRequest
		mockResponse   *models.ChatCompletionResponse
		mockError      error
		expectedStatus int
	}{
		{
			name: "valid request",
			request: models.ChatCompletionRequest{
				Model: "test-model",
				Messages: []models.ChatMessage{
					{Role: "user", Content: "Hello"},
				},
			},
			mockResponse:   mockResponse,
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name: "missing model",
			request: models.ChatCompletionRequest{
				Messages: []models.ChatMessage{
					{Role: "user", Content: "Hello"},
				},
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "empty messages",
			request: models.ChatCompletionRequest{
				Model:    "test-model",
				Messages: []models.ChatMessage{},
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockOpenWebUIClient{
				chatResponse: tt.mockResponse,
				chatError:    tt.mockError,
			}

			server := NewRESTServer(mockClient)

			reqBody, _ := json.Marshal(tt.request)
			req := httptest.NewRequest("POST", "/api/chat/completions", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			server.handleChatCompletion(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var response models.ChatCompletionResponse
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}

				if response.ID != mockResponse.ID {
					t.Errorf("expected ID %s, got %s", mockResponse.ID, response.ID)
				}
			}
		})
	}
}

func TestRESTServer_ValidateChatCompletionRequest(t *testing.T) {
	server := &RESTServer{}

	tests := []struct {
		name    string
		request models.ChatCompletionRequest
		wantErr bool
	}{
		{
			name: "valid request",
			request: models.ChatCompletionRequest{
				Model: "test-model",
				Messages: []models.ChatMessage{
					{Role: "user", Content: "Hello"},
				},
			},
			wantErr: false,
		},
		{
			name: "missing model",
			request: models.ChatCompletionRequest{
				Messages: []models.ChatMessage{
					{Role: "user", Content: "Hello"},
				},
			},
			wantErr: true,
		},
		{
			name: "empty messages",
			request: models.ChatCompletionRequest{
				Model:    "test-model",
				Messages: []models.ChatMessage{},
			},
			wantErr: true,
		},
		{
			name: "message missing role",
			request: models.ChatCompletionRequest{
				Model: "test-model",
				Messages: []models.ChatMessage{
					{Content: "Hello"},
				},
			},
			wantErr: true,
		},
		{
			name: "message missing content",
			request: models.ChatCompletionRequest{
				Model: "test-model",
				Messages: []models.ChatMessage{
					{Role: "user"},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := server.validateChatCompletionRequest(&tt.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateChatCompletionRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
