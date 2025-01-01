package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

// Config represents the JSON structure
type Configuration struct {
	DefaultAPIURL        string `json:"default_api_url"`
	MaxInputLength       int    `json:"max_input_length"`
	HTTPClientTimeout    int    `json:"http_client_timeout"`
	BufferSize           int    `json:"buffer_size"`
	SelectedResponseType string `json:"selected_response_type"`
	BearerToken          string `json:"bearer_token"`
}

func NewConfigFromFile(filePath string) (*Configuration, error) {
	// Read the file content
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Unmarshal JSON into Configuration struct
	var config Configuration
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config JSON: %w", err)
	}

	// Perform sanitization
	if config.DefaultAPIURL == "" {
		return nil, errors.New("invalid configuration: DefaultAPIURL is required")
	}

	if !strings.HasPrefix(config.DefaultAPIURL, "http://") && !strings.HasPrefix(config.DefaultAPIURL, "https://") {
		return nil, errors.New("invalid configuration: DefaultAPIURL must start with http:// or https://")
	}

	if config.MaxInputLength <= 0 {
		config.MaxInputLength = 512 // Default value
	}

	if config.HTTPClientTimeout <= 0 {
		config.HTTPClientTimeout = 10 // Default to 10 seconds
	}

	if config.BufferSize <= 0 {
		config.BufferSize = 1024 // Default to 1KB
	}

	if config.SelectedResponseType == "" {
		config.SelectedResponseType = "OpenAI" // Default response type
	}

	if config.BearerToken == "" {
		return nil, errors.New("invalid configuration: BearerToken is required")
	}

	return &config, nil
}

// RequestPayload represents the structure of the request payload.
type RequestPayload struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

func (r *RequestPayload) AddMessage(role Role, content string, images []string) {
	r.Messages = append(r.Messages, NewMessage(role, content, images))
}

// Message represents the structure of the request Message.
type Message struct {
	Role    string   `json:"role"`
	Content string   `json:"content"`
	Images  []string `json:"images,omitempty"`
}

func NewMessage(role Role, content string, images []string) Message {
	return Message{
		Role:    string(role),
		Content: content,
		Images:  images,
	}
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

type Scene struct {
	Assistant1     string `json:"assistant1"`
	Assistant2     string `json:"assistant2"`
	OpeningMessage string `json:"opening-message"`
}

func NewSceneFromFile(filePath string) (*Scene, error) {
	// Read the file content
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Unmarshal JSON into Configuration struct
	var scene Scene
	if err := json.Unmarshal(data, &scene); err != nil {
		return nil, fmt.Errorf("failed to parse config JSON: %w", err)
	}

	return &scene, nil
}
