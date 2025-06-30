package api

import (
	"context"
	"fmt"
	"log"

	"github.com/fr0g-vibe/fr0g-ai-bridge/internal/client"
	"github.com/fr0g-vibe/fr0g-ai-bridge/internal/models"
	pb "github.com/fr0g-vibe/fr0g-ai-bridge/internal/pb"
)

// GRPCServer implements the Fr0gAiBridge gRPC service
type GRPCServer struct {
	client *client.OpenWebUIClient
}

// NewGRPCServer creates a new gRPC server
func NewGRPCServer(openWebUIClient *client.OpenWebUIClient) *GRPCServer {
	return &GRPCServer{
		client: openWebUIClient,
	}
}

// HealthCheck implements the health check endpoint
func (s *GRPCServer) HealthCheck(ctx context.Context, req *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error) {
	// Check OpenWebUI health
	err := s.client.HealthCheck(ctx)
	
	response := &pb.HealthCheckResponse{
		Version: "1.0.0",
	}

	if err != nil {
		response.Status = "unhealthy"
		log.Printf("gRPC Health check failed: %v", err)
	} else {
		response.Status = "healthy"
	}

	return response, nil
}

// ChatCompletion implements the chat completion endpoint
func (s *GRPCServer) ChatCompletion(ctx context.Context, req *pb.ChatCompletionRequest) (*pb.ChatCompletionResponse, error) {
	// Validate request
	if err := s.validateChatCompletionRequest(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Convert protobuf request to internal model
	modelReq := s.protoToModel(req)

	// Forward to OpenWebUI
	resp, err := s.client.ChatCompletion(ctx, modelReq)
	if err != nil {
		return nil, fmt.Errorf("failed to process chat completion: %w", err)
	}

	// Convert response back to protobuf
	protoResp := s.modelToProto(resp)

	return protoResp, nil
}

// validateChatCompletionRequest validates the gRPC chat completion request
func (s *GRPCServer) validateChatCompletionRequest(req *pb.ChatCompletionRequest) error {
	if req.Model == "" {
		return fmt.Errorf("model is required")
	}
	if len(req.Messages) == 0 {
		return fmt.Errorf("messages are required")
	}
	for i, msg := range req.Messages {
		if msg.Role == "" {
			return fmt.Errorf("message %d: role is required", i)
		}
		if msg.Content == "" {
			return fmt.Errorf("message %d: content is required", i)
		}
	}
	return nil
}

// protoToModel converts protobuf request to internal model
func (s *GRPCServer) protoToModel(req *pb.ChatCompletionRequest) *models.ChatCompletionRequest {
	modelReq := &models.ChatCompletionRequest{
		Model:         req.Model,
		PersonaPrompt: req.PersonaPrompt,
	}

	// Convert messages
	for _, msg := range req.Messages {
		modelReq.Messages = append(modelReq.Messages, models.ChatMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	// Convert optional fields
	if req.Temperature != nil {
		temp := *req.Temperature
		modelReq.Temperature = &temp
	}

	if req.MaxTokens != nil {
		maxTokens := int(*req.MaxTokens)
		modelReq.MaxTokens = &maxTokens
	}

	if req.Stream != nil {
		stream := *req.Stream
		modelReq.Stream = &stream
	}

	return modelReq
}

// modelToProto converts internal model response to protobuf
func (s *GRPCServer) modelToProto(resp *models.ChatCompletionResponse) *pb.ChatCompletionResponse {
	protoResp := &pb.ChatCompletionResponse{
		Id:      resp.ID,
		Object:  resp.Object,
		Created: resp.Created,
		Model:   resp.Model,
		Usage: &pb.Usage{
			PromptTokens:     int32(resp.Usage.PromptTokens),
			CompletionTokens: int32(resp.Usage.CompletionTokens),
			TotalTokens:      int32(resp.Usage.TotalTokens),
		},
	}

	// Convert choices
	for _, choice := range resp.Choices {
		protoResp.Choices = append(protoResp.Choices, &pb.Choice{
			Index: int32(choice.Index),
			Message: &pb.ChatMessage{
				Role:    choice.Message.Role,
				Content: choice.Message.Content,
			},
			FinishReason: choice.FinishReason,
		})
	}

	return protoResp
}
