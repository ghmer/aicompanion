package openai

import (
	"encoding/json"

	"github.com/ghmer/aicompanion/models"
)

// ModelsRequest represents the request payload for the Models endpoint.
type ModelsRequest struct {
	// No parameters required for listing models.
}

// Model represents a single model in the response.
type Model struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	OwnedBy string `json:"owned_by"`
}

// ModelResponse represents the response structure for the models endpoint.
type ModelResponse struct {
	Object string  `json:"object"`
	Models []Model `json:"data"`
}

// CompletionsRequest represents the input payload for generating text completions.
type CompletionsRequest struct {
	Model       string   `json:"model"`
	Prompt      string   `json:"prompt"`
	MaxTokens   int      `json:"max_tokens,omitempty"`
	Temperature float32  `json:"temperature,omitempty"`
	TopP        float32  `json:"top_p,omitempty"`
	N           int      `json:"n,omitempty"`
	Stream      bool     `json:"stream,omitempty"`
	Stop        []string `json:"stop,omitempty"`
}

// Choice represents a single completion choice.
type Choice struct {
	Delta        Delta   `json:"delta"`
	Index        int     `json:"index"`
	LogProbs     float32 `json:"logprobs,omitempty"`
	FinishReason string  `json:"finish_reason,omitempty"`
	Message      Message `json:"message,omitempty"`
}

type Delta struct {
	Content string `json:"content"`
}

// CompletionsResponse represents the output of a text completion request.
type CompletionsResponse struct {
	ID                string   `json:"id"`
	Object            string   `json:"object"`
	Created           int64    `json:"created"`
	Model             string   `json:"model"`
	Choices           []Choice `json:"choices,omitempty"`
	Usage             Usage    `json:"usage,omitempty"`
	SystemFingerprint string   `json:"system_fingerprint"`
}

type Usage struct {
	PromptTokens            int                     `json:"prompt_tokens,omitempty"`
	CompletionTokens        int                     `json:"completion_tokens,omitempty"`
	TotalTokens             int                     `json:"total_tokens,omitempty"`
	PromptTokensDetails     PromptTokensDetails     `json:"prompt_tokens_details,omitempty"`
	CompletionTokensDetails CompletionTokensDetails `json:"completion_tokens_details,omitempty"`
}

type PromptTokensDetails struct {
	CachedTokens int `json:"cached_tokens"`
}

type CompletionTokensDetails struct {
	ReasoningTokens          int `json:"reasoning_tokens,omitempty"`           // Assuming reasoning_tokens is a nullable field
	AcceptedPredictionTokens int `json:"accepted_prediction_tokens,omitempty"` // Assuming accepted_prediction_tokens is a nullable field
	RejectedPredictionTokens int `json:"rejected_prediction_tokens,omitempty"` // Assuming rejected_prediction_tokens is a nullable field
}

// ChatRequest represents the input payload for chat completions.
type ChatRequest struct {
	Model       string            `json:"model"`
	Messages    []models.Message  `json:"messages"`
	MaxTokens   int               `json:"max_tokens,omitempty"`
	Temperature float32           `json:"temperature,omitempty"`
	Stream      bool              `json:"stream,omitempty"`
	Tools       []models.Function `json:"tools,omitempty"`
}

// Message represents an individual message in the chat.
type Message struct {
	Role            models.Role           `json:"role"`             // Role of the message (user, assistant, system)
	Content         string                `json:"content"`          // Content of the message
	Images          *[]models.Base64Image `json:"images,omitempty"` // Images associated with the message
	AlternatePrompt string                `json:"alternate_prompt,omitempty"`
	ToolCalls       []ToolCall            `json:"tool_calls,omitempty"`
}

// ChatResponse represents the response for a chat completion.
type ChatResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
}

// EmbeddingsRequest represents the input payload for generating embeddings.
type EmbeddingsRequest struct {
	Model          string         `json:"model"`
	Input          []string       `json:"input"`
	EncodingFormat EncodingFormat `json:"encoding_format"`
}

type EncodingFormat string

const (
	Float  EncodingFormat = "float"
	Base64 EncodingFormat = "base64"
)

// EmbeddingResponse represents the output of the embeddings request.
type EmbeddingResponse struct {
	Object string      `json:"object"`
	Data   []Embedding `json:"data"`
	Model  string      `json:"model"`
	Usage  Usage       `json:"usage"`
}

// Embedding represents a single embedding vector.
type Embedding struct {
	Object    string    `json:"object"`
	Embedding []float32 `json:"embedding"`
	Index     int       `json:"index"`
}

// ImagesRequest represents the input payload for generating images.
type ImagesRequest struct {
	Prompt string `json:"prompt"`
	N      int    `json:"n"`
	Size   string `json:"size"`
}

// ImageData represents a single image in the response.
type ImageData struct {
	URL string `json:"url"`
}

// ImagesResponse represents the response for an image generation request.
type ImagesResponse struct {
	Created int64       `json:"created"`
	Data    []ImageData `json:"data"`
}

// ModerationRequest represents a request to check if a given text contains any content that is considered inappropriate or harmful by OpenAI's standards.
type ModerationRequest struct {
	Input string `json:"input"`
}

// ModerationResponse represents the root structure of the moderation response.
type ModerationResponse struct {
	ID      string             `json:"id"`
	Model   string             `json:"model"`
	Results []ModerationResult `json:"results"`
}

// ModerationResult represents a single result in the moderation response.
type ModerationResult struct {
	Flagged        bool                     `json:"flagged"`
	Categories     ModerationCategories     `json:"categories"`
	CategoryScores ModerationCategoryScores `json:"category_scores"`
}

// ModerationCategories represents the categories for moderation.
type ModerationCategories struct {
	Sexual                bool `json:"sexual"`
	Hate                  bool `json:"hate"`
	Harassment            bool `json:"harassment"`
	SelfHarm              bool `json:"self-harm"`
	SexualMinors          bool `json:"sexual/minors"`
	HateThreatening       bool `json:"hate/threatening"`
	ViolenceGraphic       bool `json:"violence/graphic"`
	SelfHarmIntent        bool `json:"self-harm/intent"`
	SelfHarmInstructions  bool `json:"self-harm/instructions"`
	HarassmentThreatening bool `json:"harassment/threatening"`
	Violence              bool `json:"violence"`
}

// ModerationCategoryScores represents the scores for each moderation category.
type ModerationCategoryScores struct {
	Sexual                float32 `json:"sexual"`
	Hate                  float32 `json:"hate"`
	Harassment            float32 `json:"harassment"`
	SelfHarm              float32 `json:"self-harm"`
	SexualMinors          float32 `json:"sexual/minors"`
	HateThreatening       float32 `json:"hate/threatening"`
	ViolenceGraphic       float32 `json:"violence/graphic"`
	SelfHarmIntent        float32 `json:"self-harm/intent"`
	SelfHarmInstructions  float32 `json:"self-harm/instructions"`
	HarassmentThreatening float32 `json:"harassment/threatening"`
	Violence              float32 `json:"violence"`
}

type ToolCall struct {
	Payload FunctionPayload `json:"function"`
}

func (toolCall *ToolCall) TransformToModel() (models.ToolCall, error) {
	var model models.ToolCall
	var arguments map[string]interface{}
	err := json.Unmarshal([]byte(toolCall.Payload.Arguments), &arguments)
	if err != nil {
		return model, err
	}
	model.Payload = models.FunctionPayload{
		FunctionName: toolCall.Payload.FunctionName,
		Arguments:    arguments,
	}
	return model, nil
}

type FunctionPayload struct {
	FunctionName string `json:"name"`      // The name of the function.
	Arguments    string `json:"arguments"` // List of parameters the function takes.
}
