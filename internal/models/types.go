package models

import "time"

// ChatMessage represents a single message in a conversation
type ChatMessage struct {
	Role    string `json:"role"`    // "user", "assistant", "system"
	Content string `json:"content"` // The message content
}

// ChatCompletionRequest represents a request to the chat completion endpoint
type ChatCompletionRequest struct {
	Model         string        `json:"model"`                   // Model name to use
	Messages      []ChatMessage `json:"messages"`                // Conversation messages
	Temperature   *float64      `json:"temperature,omitempty"`   // Sampling temperature
	MaxTokens     *int          `json:"max_tokens,omitempty"`    // Maximum tokens to generate
	Stream        *bool         `json:"stream,omitempty"`        // Whether to stream the response
	PersonaPrompt string        `json:"persona_prompt,omitempty"` // Additional persona context
}

// ChatCompletionResponse represents the response from chat completion
type ChatCompletionResponse struct {
	ID      string    `json:"id"`      // Unique response ID
	Object  string    `json:"object"`  // Object type
	Created int64     `json:"created"` // Creation timestamp
	Model   string    `json:"model"`   // Model used
	Choices []Choice  `json:"choices"` // Response choices
	Usage   Usage     `json:"usage"`   // Token usage information
}

// Choice represents a single response choice
type Choice struct {
	Index        int         `json:"index"`         // Choice index
	Message      ChatMessage `json:"message"`       // The response message
	FinishReason string      `json:"finish_reason"` // Reason for completion
}

// Usage represents token usage statistics
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`     // Tokens in the prompt
	CompletionTokens int `json:"completion_tokens"` // Tokens in the completion
	TotalTokens      int `json:"total_tokens"`      // Total tokens used
}

// HealthResponse represents a health check response
type HealthResponse struct {
	Status  string    `json:"status"`
	Version string    `json:"version"`
	Time    time.Time `json:"time"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	Code    int    `json:"code,omitempty"`
}
