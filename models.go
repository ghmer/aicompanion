package main

import "time"

// Config represents the JSON structure
type Configuration struct {
	DefaultAPIURL        string `json:"default_api_url"`
	MaxInputLength       int    `json:"max_input_length"`
	HTTPClientTimeout    int    `json:"http_client_timeout"`
	BufferSize           int    `json:"buffer_size"`
	SelectedResponseType string `json:"selected_response_type"`
	BearerToken          string `json:"bearer_token"`
}

// RequestPayload represents the structure of the request payload.
type RequestPayload struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

// Message represents the structure of a message
type Message struct {
	Role    string   `json:"role"`
	Content string   `json:"content"`
	Images  []string `json:"images,omitempty"`
}

// ResponseChunk represents a single chunk of data from the OpenAI API stream.
type ResponseChunk struct {
	ID      string   `json:"id"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Object  string   `json:"object"`
}

// Choice represents a choice the AI made
type Choice struct {
	Index        int    `json:"index"`
	Logprobs     *int   `json:"logprobs"`
	FinishReason string `json:"finish_reason"`
	Delta        Delta  `json:"delta"`
}

// Delta represents the next token in the stream
type Delta struct {
	Content string `json:"content"`
}

// ResponsePayload represents a single chunk of data from the Ollama API stream
type ResponsePayload struct {
	Model     string    `json:"model"`
	CreatedAt time.Time `json:"created_at"`
	Message   Message   `json:"message"`
	Done      bool      `json:"done"`
}

type AIModel string

const (
	DefaultModel AIModel = "llama3.2:latest"
	VisionModel  AIModel = "llama3.2-vision:latest"
)

type ResponseType string

const (
	OpenAI = "OpenAI"
	Ollama = "Ollama"
)

type Role string

const (
	System    Role = "system"
	Assistant Role = "assistant"
	User      Role = "user"
)
