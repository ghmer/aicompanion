package models

import (
	"ai-companion/terminal"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
)

// Configuration represents the configuration for the application.
type Configuration struct {
	AIType            AIType      `json:"ai_type"`             // AI interaction type (chat, embed, moderation)
	AiModel           string      `json:"ai_model"`            // Specific AI model to use
	ApiChatURL        string      `json:"api_chat_url"`        // URL for chat API
	ApiGenerateURL    string      `json:"api_generate_url"`    // URL for generate API
	ApiEmbedURL       string      `json:"api_embed_url"`       // URL for embedding API
	ApiModerationURL  string      `json:"api_moderation_url"`  // URL for moderation API
	MaxInputLength    int         `json:"max_input_length"`    // Maximum length of input allowed
	HTTPClientTimeout int         `json:"http_client_timeout"` // HTTP client timeout duration
	BufferSize        int         `json:"buffer_size"`         // Buffer size for processing data
	ApiProvider       ApiProvider `json:"api_provider"`        // API provider used
	ApiKey            string      `json:"api_key"`             // API key for authentication
	MaxMessages       int         `json:"max_messages"`        // Maximum number of messages in a conversation
	UserColor         string      `json:"term_color"`          // Color for user output in terminal
	Output            bool        `json:"term_output"`         // Flag to enable/disabled terminal output
	Color             terminal.TermColor
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

	if config.AIType == "" {
		return nil, errors.New("invalid configuration: AIType is required")
	}

	if config.AiModel == "" {
		return nil, errors.New("invalid configuration: AiModel is required")
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

	// If AIType is Chat, validate and set default URLs if not provided
	if config.AIType == Chat {
		if config.ApiChatURL == "" {
			fmt.Println("using default url for chat api")
			if config.ApiProvider == Ollama {
				config.ApiChatURL = "http://localhost:11434/api/chat"
			}
			if config.ApiProvider == OpenAI {
				config.ApiChatURL = "https://api.openai.com/v1/chat/completions"
			}
		}

		// Ensure URL starts with http:// or https://
		if !strings.HasPrefix(config.ApiChatURL, "http://") && !strings.HasPrefix(config.ApiChatURL, "https://") {
			return nil, errors.New("invalid configuration: ApiChatURL must start with http:// or https://")
		}

		// If AIType is Chat, validate and set default URLs if not provided
		if config.ApiGenerateURL == "" {
			fmt.Println("using default url for chat api")
			if config.ApiProvider == Ollama {
				config.ApiGenerateURL = "http://localhost:11434/api/generate"
			}
			if config.ApiProvider == OpenAI {
				config.ApiGenerateURL = "https://api.openai.com/v1/completions"
			}
		}

		// Ensure URL starts with http:// or https://
		if !strings.HasPrefix(config.ApiGenerateURL, "http://") && !strings.HasPrefix(config.ApiGenerateURL, "https://") {
			return nil, errors.New("invalid configuration: ApiGenerateURL must start with http:// or https://")
		}
	}

	// If AIType is Embed, validate and set default URLs if not provided
	if config.AIType == Embed {
		if config.ApiEmbedURL == "" {
			fmt.Println("using default url for chat api")
			if config.ApiProvider == Ollama {
				config.ApiEmbedURL = "http://localhost:11434/api/embed"
			}
			if config.ApiProvider == OpenAI {
				config.ApiEmbedURL = "https://api.openai.com/v1/embeddings"
			}
		}

		// Ensure URL starts with http:// or https://
		if !strings.HasPrefix(config.ApiEmbedURL, "http://") && !strings.HasPrefix(config.ApiEmbedURL, "https://") {
			return nil, errors.New("invalid configuration: ApiEmbedURL must start with http:// or https://")
		}
	}

	// If AIType is Moderation, validate and set default URLs if not provided
	if config.AIType == Moderation {
		if config.ApiModerationURL == "" {
			fmt.Println("using default url for chat api")
			if config.ApiProvider == Ollama {
				config.ApiModerationURL = "http://localhost:11434/api/moderate"
			}
			if config.ApiProvider == OpenAI {
				config.ApiModerationURL = "https://api.openai.com/v1/moderations"
			}
		}

		// Ensure URL starts with http:// or https://
		if !strings.HasPrefix(config.ApiModerationURL, "http://") && !strings.HasPrefix(config.ApiModerationURL, "https://") {
			return nil, errors.New("invalid configuration: ApiModerationURL must start with http:// or https://")
		}
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

	if config.ApiKey == "" {
		return nil, errors.New("invalid configuration: api_key is required")
	}

	return &config, nil
}

// Message represents an individual message in the chat.
type Message struct {
	Role    Role     `json:"role"`             // Role of the message (user, assistant, system)
	Content string   `json:"content"`          // Content of the message
	Images  []string `json:"images,omitempty"` // Images associated with the message
}

// ApiProvider indicates the type of response expected.
type ApiProvider string

const (
	OpenAI = "OpenAI" // OpenAI model type
	Ollama = "Ollama" // Ollama model type
)

// AIType indicates the type of AI interaction, either chat or generate.
type AIType string

const (
	Chat       AIType = "chat"       // Chat completion type
	Embed      AIType = "embed"      // Embedding type
	Moderation AIType = "moderation" // Moderation type
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
	Model           string      `json:"model"`                       // Model used for embedding
	Embeddings      [][]float64 `json:"embeddings"`                  // Generated embeddings
	TotalDuration   int64       `json:"total_duration,omitempty"`    // Total duration of the request
	LoadDuration    int64       `json:"load_duration,omitempty"`     // Duration spent loading data
	PromptEvalCount int         `json:"prompt_eval_count,omitempty"` // Number of prompt evaluations made
}
