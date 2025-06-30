package api

import (
	"context"
	"testing"
	"time"

	"github.com/fr0g-vibe/fr0g-ai-bridge/internal/models"
	pb "github.com/fr0g-vibe/fr0g-ai-bridge/internal/pb"
)

func TestGRPCServer_HealthCheck(t *testing.T) {
	tests := []struct {
		name          string
		healthError   error
		expectedStatus string
	}{
		{
			name:          "healthy",
			healthError:   nil,
			expectedStatus: "healthy",
		},
		{
			name:          "unhealthy",
			healthError:   context.DeadlineExceeded,
			expectedStatus: "unhealthy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockOpenWebUIClient{
				healthCheckError: tt.healthError,
			}

			server := NewGRPCServer(mockClient)
			req := &pb.HealthCheckRequest{}

			resp, err := server.HealthCheck(context.Background(), req)
			if err != nil {
				t.Fatalf("HealthCheck failed: %v", err)
			}

			if resp.Status != tt.expectedStatus {
				t.Errorf("expected status %s, got %s", tt.expectedStatus, resp.Status)
			}

			if resp.Version != "1.0.0" {
				t.Errorf("expected version 1.0.0, got %s", resp.Version)
			}
		})
	}
}

func TestGRPCServer_ChatCompletion(t *testing.T) {
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
		name         string
		request      *pb.ChatCompletionRequest
		mockResponse *models.ChatCompletionResponse
		mockError    error
		wantErr      bool
	}{
		{
			name: "valid request",
			request: &pb.ChatCompletionRequest{
				Model: "test-model",
				Messages: []*pb.ChatMessage{
					{Role: "user", Content: "Hello"},
				},
			},
			mockResponse: mockResponse,
			mockError:    nil,
			wantErr:      false,
		},
		{
			name: "missing model",
			request: &pb.ChatCompletionRequest{
				Messages: []*pb.ChatMessage{
					{Role: "user", Content: "Hello"},
				},
			},
			wantErr: true,
		},
		{
			name: "empty messages",
			request: &pb.ChatCompletionRequest{
				Model:    "test-model",
				Messages: []*pb.ChatMessage{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockOpenWebUIClient{
				chatResponse: tt.mockResponse,
				chatError:    tt.mockError,
			}

			server := NewGRPCServer(mockClient)

			resp, err := server.ChatCompletion(context.Background(), tt.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("ChatCompletion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && resp != nil {
				if resp.Id != mockResponse.ID {
					t.Errorf("expected ID %s, got %s", mockResponse.ID, resp.Id)
				}
				if resp.Model != mockResponse.Model {
					t.Errorf("expected model %s, got %s", mockResponse.Model, resp.Model)
				}
			}
		})
	}
}

func TestGRPCServer_ProtoToModel(t *testing.T) {
	server := &GRPCServer{}
	
	temp := 0.7
	maxTokens := int32(100)
	stream := true

	protoReq := &pb.ChatCompletionRequest{
		Model: "test-model",
		Messages: []*pb.ChatMessage{
			{Role: "user", Content: "Hello"},
			{Role: "assistant", Content: "Hi there!"},
		},
		Temperature:   &temp,
		MaxTokens:     &maxTokens,
		Stream:        &stream,
		PersonaPrompt: "You are a helpful assistant",
	}

	modelReq := server.protoToModel(protoReq)

	if modelReq.Model != "test-model" {
		t.Errorf("expected model test-model, got %s", modelReq.Model)
	}

	if len(modelReq.Messages) != 2 {
		t.Errorf("expected 2 messages, got %d", len(modelReq.Messages))
	}

	if modelReq.Messages[0].Role != "user" || modelReq.Messages[0].Content != "Hello" {
		t.Errorf("first message not converted correctly")
	}

	if modelReq.Temperature == nil || *modelReq.Temperature != 0.7 {
		t.Errorf("temperature not converted correctly")
	}

	if modelReq.MaxTokens == nil || *modelReq.MaxTokens != 100 {
		t.Errorf("max_tokens not converted correctly")
	}

	if modelReq.Stream == nil || *modelReq.Stream != true {
		t.Errorf("stream not converted correctly")
	}

	if modelReq.PersonaPrompt != "You are a helpful assistant" {
		t.Errorf("persona_prompt not converted correctly")
	}
}

func TestGRPCServer_ModelToProto(t *testing.T) {
	server := &GRPCServer{}

	modelResp := &models.ChatCompletionResponse{
		ID:      "test-id",
		Object:  "chat.completion",
		Created: 1234567890,
		Model:   "test-model",
		Choices: []models.Choice{
			{
				Index: 0,
				Message: models.ChatMessage{
					Role:    "assistant",
					Content: "Hello!",
				},
				FinishReason: "stop",
			},
		},
		Usage: models.Usage{
			PromptTokens:     10,
			CompletionTokens: 5,
			TotalTokens:      15,
		},
	}

	protoResp := server.modelToProto(modelResp)

	if protoResp.Id != "test-id" {
		t.Errorf("expected ID test-id, got %s", protoResp.Id)
	}

	if protoResp.Model != "test-model" {
		t.Errorf("expected model test-model, got %s", protoResp.Model)
	}

	if len(protoResp.Choices) != 1 {
		t.Errorf("expected 1 choice, got %d", len(protoResp.Choices))
	}

	if protoResp.Choices[0].Message.Role != "assistant" {
		t.Errorf("choice message role not converted correctly")
	}

	if protoResp.Usage.PromptTokens != 10 {
		t.Errorf("expected prompt tokens 10, got %d", protoResp.Usage.PromptTokens)
	}
}
