package models

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/ghmer/aicompanion/terminal"
)

// Model represents an AI model with its name and identifier.
type Model struct {
	Model string `json:"model"`
	Name  string `json:"name"`
}

// Document represents a stored document with metadata and embeddings.
type Document struct {
	ID         string                 `json:"id"`
	ClassName  string                 `json:"classname"`
	Embeddings []float32              `json:"embeddings"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// Configuration represents the configuration for the application.
type Configuration struct {
	ApiProvider      ApiProvider           `json:"api_provider"` // API provider used
	ApiKey           string                `json:"api_key"`      // API key for authentication
	ApiEndpoints     ApiEndpointUrls       `json:"api_endpoints"`
	AiModels         AiModels              `json:"ai_models"` // Specific AI model to use
	HttpConfig       HttpConfiguration     `json:"http_config"`
	VectorDBConfig   VectorDbConfiguration `json:"vectordb_config"`
	MaxMessages      int                   `json:"max_messages"` // Maximum number of messages in a conversation
	UserColor        string                `json:"term_color"`   // Color for user output in terminal
	Output           bool                  `json:"term_output"`  // Flag to enable/disabled terminal output
	SystemPrompt     string                `json:"system_prompt"`
	EnrichmentPrompt string                `json:"enrichment_prompt"`
	Color            terminal.TermColor
}

type AiModels struct {
	ChatModel      Model `json:"chat_model"`
	EmbeddingModel Model `json:"embedding_model"`
}

type ApiEndpointUrls struct {
	ApiChatURL       string `json:"api_chat_url"`       // URL for chat API
	ApiGenerateURL   string `json:"api_generate_url"`   // URL for generate API
	ApiEmbedURL      string `json:"api_embed_url"`      // URL for embedding API
	ApiModerationURL string `json:"api_moderation_url"` // URL for moderation API
	ApiModelsURL     string `json:"api_models_url"`     // URL for model API
}

type HttpConfiguration struct {
	MaxInputLength    int `json:"max_input_length"`    // Maximum length of input allowed
	HTTPClientTimeout int `json:"http_client_timeout"` // HTTP client timeout duration
	BufferSize        int `json:"buffer_size"`         // Buffer size for processing data
}

type VectorDbType string

const (
	SqlVectorDb VectorDbType = "sqlvdb"
	WeaviateDb  VectorDbType = "weaviate"
)

type VectorDbConfiguration struct {
	Type     VectorDbType `json:"type"`
	Endpoint string       `json:"endpoint_url"`
	ApiKey   string       `json:"api_key"`
}

// NewConfigFromFile creates a new Configuration instance from a JSON file.
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
	if config.AiModels.ChatModel.Model == "" {
		return nil, errors.New("invalid configuration: ChatModel is required")
	}

	// Perform sanitization
	if config.AiModels.EmbeddingModel.Model == "" {
		return nil, errors.New("invalid configuration: EmbeddingModel is required")
	}

	if config.UserColor == "" {
		config.Color = terminal.Green
	} else {
		color, exists := terminal.TranslateColor(config.UserColor)
		if !exists {
			config.Color = terminal.BrightMagenta
		} else {
			config.Color = color
		}
	}

	if config.ApiProvider == "" {
		config.ApiProvider = "Ollama" // Default api provider
	}

	// set default urls if no custom ones were provided
	if config.ApiEndpoints.ApiChatURL == "" {
		fmt.Print("using default url for chat api: ")
		if config.ApiProvider == Ollama {
			config.ApiEndpoints.ApiChatURL = "http://localhost:11434/api/chat"
		}
		if config.ApiProvider == OpenAI {
			config.ApiEndpoints.ApiChatURL = "https://api.openai.com/v1/chat/completions"
		}
	}

	// Ensure URL starts with http:// or https://
	if !strings.HasPrefix(config.ApiEndpoints.ApiChatURL, "http://") && !strings.HasPrefix(config.ApiEndpoints.ApiChatURL, "https://") {
		return nil, errors.New("invalid configuration: ApiChatURL must start with http:// or https://")
	}

	// If AIType is Generate, validate and set default URLs if not provided
	if config.ApiEndpoints.ApiGenerateURL == "" {
		fmt.Print("using default url for generate api: ")
		if config.ApiProvider == Ollama {
			config.ApiEndpoints.ApiGenerateURL = "http://localhost:11434/api/generate"
		}
		if config.ApiProvider == OpenAI {
			config.ApiEndpoints.ApiGenerateURL = "https://api.openai.com/v1/completions"
		}
	}

	// Ensure URL starts with http:// or https://
	if !strings.HasPrefix(config.ApiEndpoints.ApiGenerateURL, "http://") && !strings.HasPrefix(config.ApiEndpoints.ApiGenerateURL, "https://") {
		return nil, errors.New("invalid configuration: ApiGenerateURL must start with http:// or https://")
	}

	if config.ApiEndpoints.ApiEmbedURL == "" {
		fmt.Print("using default url for embed api: ")
		if config.ApiProvider == Ollama {
			config.ApiEndpoints.ApiEmbedURL = "http://localhost:11434/api/embed"
		}
		if config.ApiProvider == OpenAI {
			config.ApiEndpoints.ApiEmbedURL = "https://api.openai.com/v1/embeddings"
		}
	}

	// Ensure URL starts with http:// or https://
	if !strings.HasPrefix(config.ApiEndpoints.ApiEmbedURL, "http://") && !strings.HasPrefix(config.ApiEndpoints.ApiEmbedURL, "https://") {
		return nil, errors.New("invalid configuration: ApiEmbedURL must start with http:// or https://")
	}

	if config.ApiEndpoints.ApiModerationURL == "" {
		fmt.Print("using default url for moderation api: ")
		if config.ApiProvider == Ollama {
			config.ApiEndpoints.ApiModerationURL = "http://localhost:11434/api/moderate"
		}
		if config.ApiProvider == OpenAI {
			config.ApiEndpoints.ApiModerationURL = "https://api.openai.com/v1/moderations"
		}
	}

	// Ensure URL starts with http:// or https://
	if !strings.HasPrefix(config.ApiEndpoints.ApiModerationURL, "http://") && !strings.HasPrefix(config.ApiEndpoints.ApiModerationURL, "https://") {
		return nil, errors.New("invalid configuration: ApiModerationURL must start with http:// or https://")
	}

	if config.HttpConfig.MaxInputLength <= 0 {
		config.HttpConfig.MaxInputLength = 2038 // Default value
	}

	if config.HttpConfig.HTTPClientTimeout <= 0 {
		config.HttpConfig.HTTPClientTimeout = 10 // Default to 10 seconds
	}

	if config.HttpConfig.BufferSize <= 0 {
		config.HttpConfig.BufferSize = 1024 // Default to 1KB
	}

	if config.ApiKey == "" {
		return nil, errors.New("invalid configuration: api_key is required")
	}

	return &config, nil
}

// Message represents an individual message in the chat.
type Message struct {
	Role    Role          `json:"role"`             // Role of the message (user, assistant, system)
	Content string        `json:"content"`          // Content of the message
	Images  []Base64Image `json:"images,omitempty"` // Images associated with the message
}

// Base64Image represents an image encoded in base64.
type Base64Image struct {
	Data string // The base64-encoded data of the image
}

// SetData encodes the provided byte slice into a base64 string and assigns it to Data.
func (image *Base64Image) SetData(data []byte) {
	image.Data = base64.StdEncoding.EncodeToString(data)
}

// GetData decodes the base64-encoded data in Data back into a byte slice.
func (image *Base64Image) GetData() ([]byte, error) {
	return base64.StdEncoding.DecodeString(image.Data)
}

// MarshalJSON custom marshals Base64Image as a single string.
func (b Base64Image) MarshalJSON() ([]byte, error) {
	return json.Marshal(b.Data)
}

// ApiProvider indicates the type of response expected.
type ApiProvider string

const (
	OpenAI = "openai" // OpenAI model type
	Ollama = "ollama" // Ollama model type
)

// Role represents a role in a conversation, such as user, assistant, or system.
type Role string

const (
	System    Role = "system"    // System role
	Developer Role = "developer" // Developer role
	Assistant Role = "assistant" // Assistant role
	User      Role = "user"      // User role
)

// EmbeddingsRequest represents the input payload for generating embeddings.
type EmbeddingRequest struct {
	Model          string         `json:"model"`                     // Model to use for embedding
	Input          []string       `json:"input"`                     // Input text or data
	EncodingFormat EncodingFormat `json:"encoding_format,omitempty"` // Encoding format for output
}

// EncodingFormat specifies the encoding format for embeddings.
type EncodingFormat string

const (
	Float  EncodingFormat = "float"  // Float format
	Base64 EncodingFormat = "base64" // Base64 format
)

// EmbeddingResponse represents the response payload from generating embeddings.
type EmbeddingResponse struct {
	Model            string      `json:"model"`             // Model used for embedding
	Embeddings       [][]float32 `json:"embeddings"`        // Generated embeddings
	OriginalResponse any         `json:"original-response"` // the original response of the API call.
}

// ModerationRequest represents a request to check if a given text contains any content that is considered inappropriate or harmful by OpenAI's standards.
type ModerationRequest struct {
	Input string `json:"input"`
}

// ModerationResponse represents the root structure of the moderation response.
type ModerationResponse struct {
	ID               string `json:"id"`
	Model            Model  `json:"model"`
	OriginalResponse any    `json:"results"`
}

type StreamType int

const (
	Generate StreamType = 1
	Chat     StreamType = 2
)
