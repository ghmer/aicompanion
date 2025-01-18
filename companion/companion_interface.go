package companion

import (
	"net/http"
	"time"

	"github.com/ghmer/aicompanion/models"
	"github.com/ghmer/aicompanion/ollama"
	"github.com/ghmer/aicompanion/openai"
)

// AICompanion defines the interface for interacting with AI models.
type AICompanion interface {
	PrepareConversation() []models.Message
	CreateMessage(role models.Role, input string) models.Message
	CreateMessageWithImages(role models.Role, message string, images []models.Base64Image) models.Message
	ReadFile(filepath string) string
	AddMessage(message models.Message)
	HandleStreamResponse(resp *http.Response) (models.Message, error)
	GetConfig() models.Configuration
	SetConfig(config models.Configuration)
	GetSystemRole() models.Message
	SetSystemRole(prompt string)
	GetConversation() []models.Message
	SetConversation(conversation []models.Message)
	GetClient() *http.Client
	SetClient(client *http.Client)
	// interactions
	SendChatRequest(message models.Message) (models.Message, error)
	SendCompletionRequest(message models.Message) (models.Message, error)
	SendEmbeddingRequest(embedding models.EmbeddingRequest) (models.EmbeddingResponse, error)
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
