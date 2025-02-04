package models

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

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
	ApiProvider     ApiProvider       `json:"api_provider"` // API provider used
	ApiKey          string            `json:"api_key"`      // API key for authentication
	ApiEndpoints    ApiEndpointUrls   `json:"api_endpoints"`
	AiModels        AiModels          `json:"ai_models"` // Specific AI model to use
	HttpConfig      HttpConfiguration `json:"http_config"`
	MaxMessages     int               `json:"max_messages"` // Maximum number of messages in a conversation
	IncludeStrategy IncludeStrategy   `json:"include_strategy"`
	Terminal        Terminal          `json:"terminal"`
	Prompt          Prompt            `json:"prompts"`
}

type Conversation struct {
	Id           string    `json:"id"`
	Summary      string    `json:"summary"`
	Created      time.Time `json:"created,omitempty"`
	Updated      time.Time `json:"updated,omitempty"`
	Conversation []Message `json:"conversation"`
}

func (c *Conversation) AddMessage(msg Message) {
	c.Conversation = append(c.Conversation, msg)
}

type IncludeStrategy string

const (
	IncludeBoth      IncludeStrategy = "both"
	IncludeAssistant IncludeStrategy = "assistant"
	IncludeUser      IncludeStrategy = "user"
)

type Terminal struct {
	UserColor string `json:"term_color"`  // Color for user output in terminal
	Output    bool   `json:"term_output"` // Flag to enable/disabled terminal output
	Debug     bool   `json:"debug"`
	Color     terminal.TermColor
}

type Prompt struct {
	SystemPrompt        string `json:"system_prompt"`
	EnrichmentPrompt    string `json:"enrichment_prompt"`
	FunctionsPrompt     string `json:"functions_prompt"`
	SummarizationPrompt string `json:"summarization_prompt"`
}

// AiModels represents the AI models used by the application.
type AiModels struct {
	ChatModel      Model `json:"chat_model"`
	GenerateModel  Model `json:"generate_model"`
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
	HTTPClientTimeout int `json:"http_client_timeout"` // HTTP client timeout duration
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

	if config.Terminal.UserColor == "" {
		config.Terminal.Color = terminal.Green
	} else {
		color, exists := terminal.TranslateColor(config.Terminal.UserColor)
		if !exists {
			config.Terminal.Color = terminal.BrightMagenta
		} else {
			config.Terminal.Color = color
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

	if config.HttpConfig.HTTPClientTimeout <= 0 {
		config.HttpConfig.HTTPClientTimeout = 10 // Default to 10 seconds
	}

	if config.ApiKey == "" {
		return nil, errors.New("invalid configuration: api_key is required")
	}

	return &config, nil
}

type MessageRequest struct {
	OriginalMessage       Message `json:"original_message,omitempty"`
	Message               Message `json:"message"`
	RetainOriginalMessage bool    `json:"retain_original"`
}

// Message represents an individual message in the chat.
type Message struct {
	Role            Role           `json:"role"`             // Role of the message (user, assistant, system)
	Content         string         `json:"content"`          // Content of the message
	Images          *[]Base64Image `json:"images,omitempty"` // Images associated with the message
	AlternatePrompt string         `json:"alternate_prompt,omitempty"`
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

// the type of streaming endpoint (chat/generate)
type StreamType int

const (
	Generate StreamType = 1
	Chat     StreamType = 2
)

type Function struct {
	Endpoint           string             `json:"function_endpoint"`
	ApiKey             string             `json:"function_apikey"`
	FunctionDefinition FunctionDefinition `json:"function_definition"` // The function definition.
}

type FunctionDefinition struct {
	FunctionName string      `json:"function_name"` // The name of the function.
	Description  string      `json:"description"`
	Parameters   []Parameter `json:"parameters"` // List of parameters the function takes.
}

// Parameter represents the details of a single parameter.
type Parameter struct {
	Name        string `json:"name"` // Name of the parameter.
	Type        string `json:"type"` // Type of the parameter.
	Description string `json:"description"`
}

type FunctionResponse struct {
}
