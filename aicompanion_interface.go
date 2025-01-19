package aicompanion

import (
	"encoding/base64"
	"net/http"
	"time"

	"github.com/ghmer/aicompanion/models"
	"github.com/ghmer/aicompanion/ollama"
	"github.com/ghmer/aicompanion/openai"
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
	SendChatRequest(message models.Message) (models.Message, error)

	// SendCompletionRequest sends a completion request to an AI model and returns a response message
	SendGenerateRequest(message models.Message) (models.Message, error)

	// SendEmbeddingRequest sends an embedding request to an AI model and returns a response
	SendEmbeddingRequest(embedding models.EmbeddingRequest) (models.EmbeddingResponse, error)

	// SendModerationRequest sends a moderation request to an AI model and returns a response
	SendModerationRequest(moderationRequest models.ModerationRequest) (models.ModerationResponse, error)
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

func ReadImageFromFile(filepath string) (models.Base64Image, error) {
	content, err := utility.ReadFile(filepath)
	if err != nil {
		return models.Base64Image{}, err
	}

	return models.Base64Image{
		Data: base64.StdEncoding.EncodeToString(content),
	}, nil
}
