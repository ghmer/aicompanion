package aicompanion

import (
	"encoding/base64"
	"net/http"
	"time"

	"github.com/ghmer/aicompanion/models"
	"github.com/ghmer/aicompanion/ollama"
	"github.com/ghmer/aicompanion/openai"
	"github.com/ghmer/aicompanion/terminal"
	"github.com/ghmer/aicompanion/utility"
)

// AICompanion defines the interface for interacting with AI models.
type AICompanion interface {
	// PrepareConversation initializes a new conversation with a default system message
	PrepareConversation() []models.Message

	// CreateMessage creates a new message with the given role and input string
	CreateMessage(role models.Role, input string) models.Message

	// CreateMessageWithImages creates a new message with the given role, input string, and images
	CreateMessageWithImages(role models.Role, message string, images []models.Base64Image) models.Message

	// AddMessage adds a new message to the conversation
	AddMessage(message models.Message)

	// GetConfig returns the current configuration for the AI companion
	GetConfig() models.Configuration

	// SetConfig sets a new configuration for the AI companion
	SetConfig(config models.Configuration)

	// GetSystemRole returns the current system role message
	GetSystemRole() models.Message

	// SetSystemRole sets a new system role message
	SetSystemRole(prompt string)

	// GetConversation returns the current conversation
	GetConversation() []models.Message

	// SetConversation sets the current conversation
	SetConversation(conversation []models.Message)

	// GetClient returns the current HTTP client used for requests
	GetClient() *http.Client

	// SetClient sets a new HTTP client for requests
	SetClient(client *http.Client)

	// interactions
	// SendChatRequest sends a chat request to an AI model and returns a response message
	SendChatRequest(message models.Message, streaming bool, callback func(m models.Message) error) (models.Message, error)

	// SendCompletionRequest sends a completion request to an AI model and returns a response message
	SendGenerateRequest(message models.Message, streaming bool, callback func(m models.Message) error) (models.Message, error)

	// SendEmbeddingRequest sends an embedding request to an AI model and returns a response
	SendEmbeddingRequest(embedding models.EmbeddingRequest) (models.EmbeddingResponse, error)

	// SendModerationRequest sends a moderation request to an AI model and returns a response
	SendModerationRequest(moderationRequest models.ModerationRequest) (models.ModerationResponse, error)

	HandleStreamResponse(resp *http.Response, streamType models.StreamType, callback func(m models.Message) error) (models.Message, error)
}

// NewCompanion creates a new Companion instance with the provided configuration.
func NewCompanion(config models.Configuration) AICompanion {
	switch config.ApiProvider {
	case models.Ollama:
		return &ollama.Companion{
			Config: config,
			SystemRole: models.Message{
				Role:    models.System,
				Content: "You are a helpful assistant",
			},
			Conversation: make([]models.Message, 0),
			Client:       &http.Client{Timeout: time.Second * time.Duration(config.HTTPClientTimeout)},
		}
	case models.OpenAI:
		return &openai.Companion{
			Config: config,
			SystemRole: models.Message{
				Role:    models.Developer,
				Content: "You are a helpful assistant",
			},
			Conversation: make([]models.Message, 0),
			Client:       &http.Client{Timeout: time.Second * time.Duration(config.HTTPClientTimeout)},
		}
	}

	return &ollama.Companion{
		Config: config,
		SystemRole: models.Message{
			Role:    models.System,
			Content: "You are a helpful assistant",
		},
		Conversation: make([]models.Message, 0),
		Client:       &http.Client{Timeout: time.Second * time.Duration(config.HTTPClientTimeout)},
	}
}

// NewDefaultConfig creates a new default configuration with the provided API provider, API token, and model.
func NewDefaultConfig(apiProvider models.ApiProvider, apiToken, model string) *models.Configuration {
	var config models.Configuration = models.Configuration{
		ApiProvider:       apiProvider,
		ApiKey:            apiToken,
		AiModel:           model,
		ApiChatURL:        "http://localhost:11434/api/chat",
		ApiGenerateURL:    "http://localhost:11434/api/generate",
		ApiEmbedURL:       "http://localhost:11434/api/embed",
		MaxInputLength:    500,
		HTTPClientTimeout: 300,
		BufferSize:        2048,
		MaxMessages:       20,
		Color:             terminal.Green,
		Output:            true,
	}

	switch apiProvider {
	case models.Ollama:
		config.ApiChatURL = "http://localhost:11434/api/chat"
		config.ApiGenerateURL = "http://localhost:11434/api/generate"
		config.ApiEmbedURL = "http://localhost:11434/api/embed"
		config.ApiModerationURL = "http://localhost:11434/api/generate"

	case models.OpenAI:
		config.ApiChatURL = "https://api.openai.com/v1/chat/completions"
		config.ApiGenerateURL = "https://api.openai.com/v1/completions"
		config.ApiEmbedURL = "https://api.openai.com/v1/embeddings"
		config.ApiModerationURL = "https://api.openai.com/v1/moderations"
	}

	return &config
}

// ReadImageFromFile reads an image from the specified filepath and returns a Base64 encoded image.
func ReadImageFromFile(filepath string) (models.Base64Image, error) {
	content, err := utility.ReadFile(filepath)
	if err != nil {
		return models.Base64Image{}, err
	}

	return models.Base64Image{
		Data: base64.StdEncoding.EncodeToString(content),
	}, nil
}
