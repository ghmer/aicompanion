// Request and response structs for the Ollama API endpoints

package ollama

import (
	"time"

	"github.com/ghmer/aicompanion/models"
)

// GenerateRequest represents the request structure for the /api/generate endpoint.
type CompletionRequest struct {
	Model     string               `json:"model"`
	Messages  []models.Message     `json:"messages,omitempty"`
	Prompt    string               `json:"prompt,omitempty"`
	Suffix    string               `json:"suffix,omitempty"`
	Images    []models.Base64Image `json:"images,omitempty"`
	Format    string               `json:"format,omitempty"`
	Options   string               `json:"options,omitempty"`
	System    string               `json:"system,omitempty"`
	Template  string               `json:"template,omitempty"`
	Stream    bool                 `json:"stream"`
	Raw       bool                 `json:"raw,omitempty"`
	KeepAlive int64                `json:"keep_alive,omitempty"`
	Context   string               `json:"context,omitempty"`
}

// ModelResponse represents the response structure for the models endpoint.
type ModelResponse struct {
	Models []models.Model `json:"models"`
}

// ChatMessage represents a message in the chat context.
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatResponse represents the response structure for the /api/chat endpoint.
type CompletionResponse struct {
	Model              string         `json:"model"`
	CreatedAt          time.Time      `json:"created_at"`
	Done               bool           `json:"done"`
	DoneReason         string         `json:"done_reason,omitempty"`
	Message            models.Message `json:"message,omitempty"`
	Response           string         `json:"response,omitempty"`
	TotalDuration      int64          `json:"total_duration"`
	LoadDuration       int64          `json:"load_duration"`
	PromptEvalCount    int            `json:"prompt_eval_count"`
	PromptEvalDuration int64          `json:"prompt_eval_duration"`
	EvalCount          int            `json:"eval_count"`
	EvalDuration       int64          `json:"eval_duration"`
	// Context is an encoding of the conversation used in this response; this
	// can be sent in the next request to keep a conversational memory.
	Context []int `json:"context,omitempty"`
}

// CreateModelRequest represents the request structure for the /api/models/create endpoint.
type CreateModelRequest struct {
	Model     string `json:"model"`
	Modelfile string `json:"modelfile"`
}

// CreateModelResponse represents the response structure for the /api/models/create endpoint.
type CreateModelResponse struct {
	Message string `json:"message"`
}

// ListModelsResponse represents the response structure for the /api/models endpoint.
type ListModelsResponse struct {
	Models []string `json:"models"`
}

// ModelInfoResponse represents the response structure for the /api/models/{model} endpoint.
type ModelInfoResponse struct {
	Name    string `json:"name"`
	Details string `json:"details"`
}

// CopyModelRequest represents the request structure for the /api/models/copy endpoint.
type CopyModelRequest struct {
	Source      string `json:"source"`
	Destination string `json:"destination"`
}

// CopyModelResponse represents the response structure for the /api/models/copy endpoint.
type CopyModelResponse struct {
	Message string `json:"message"`
}

// DeleteModelResponse represents the response structure for the /api/models/{model} DELETE endpoint.
type DeleteModelResponse struct {
	Message string `json:"message"`
}

// PullModelRequest represents the request structure for the /api/models/pull endpoint.
type PullModelRequest struct {
	Model string `json:"model"`
}

// PullModelResponse represents the response structure for the /api/models/pull endpoint.
type PullModelResponse struct {
	Message string `json:"message"`
}

// PushModelRequest represents the request structure for the /api/models/push endpoint.
type PushModelRequest struct {
	Model string `json:"model"`
}

// PushModelResponse represents the response structure for the /api/models/push endpoint.
type PushModelResponse struct {
	Message string `json:"message"`
}

// EmbeddingsRequest represents the request structure for the /api/embeddings endpoint.
type EmbeddingsRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
}

// EmbeddingsResponse represents the response structure for the /api/embeddings endpoint.
type EmbeddingResponse struct {
	Model           string      `json:"model"`
	Embeddings      [][]float32 `json:"embeddings"`
	TotalDuration   int64       `json:"total_duration,omitempty"`
	LoadDuration    int64       `json:"load_duration,omitempty"`
	PromptEvalCount int         `json:"prompt_eval_count,omitempty"`
}

// RunningModelsResponse represents the response structure for the /api/models/running endpoint.
type RunningModelsResponse struct {
	RunningModels []string `json:"running_models"`
}

// VersionResponse represents the response structure for the /api/version endpoint.
type VersionResponse struct {
	Version string `json:"version"`
}
