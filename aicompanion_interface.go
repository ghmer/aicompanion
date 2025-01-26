package aicompanion

import (
	"encoding/base64"
	"net/http"
	"time"

	"github.com/ghmer/aicompanion/models"
	"github.com/ghmer/aicompanion/ollama"
	"github.com/ghmer/aicompanion/openai"
	"github.com/ghmer/aicompanion/rag"
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

	//CreateUserMessage creates a new user message with the given input string
	CreateUserMessage(input string, images []models.Base64Image) models.Message

	// CreateAssistantMessage creates a new assistant message with the given input string
	CreateAssistantMessage(input string) models.Message

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

	// GetSystemRole returns the current system role message
	GetEnrichmentPrompt() string

	// SetSystemRole sets a new system role message
	SetEnrichmentPrompt(prompt string)

	// GetConversation returns the current conversation
	GetConversation() []models.Message

	// SetConversation sets the current conversation
	SetConversation(conversation []models.Message)

	// GetClient returns the current HTTP client used for requests
	GetClient() *http.Client

	// SetClient sets a new HTTP client for requests
	SetClient(client *http.Client)

	// interactions
	// GetModels returns all models that the endpoint supports
	GetModels() ([]models.Model, error)

	// SendChatRequest sends a chat request to an AI model and returns a response message
	SendChatRequest(message models.Message, streaming bool, callback func(m models.Message) error) (models.Message, error)

	// SendCompletionRequest sends a completion request to an AI model and returns a response message
	SendGenerateRequest(message models.Message, streaming bool, callback func(m models.Message) error) (models.Message, error)

	// SendEmbeddingRequest sends an embedding request to an AI model and returns a response
	SendEmbeddingRequest(embedding models.EmbeddingRequest) (models.EmbeddingResponse, error)

	// SendModerationRequest sends a moderation request to an AI model and returns a response
	SendModerationRequest(moderationRequest models.ModerationRequest) (models.ModerationResponse, error)

	HandleStreamResponse(resp *http.Response, streamType models.StreamType, callback func(m models.Message) error) (models.Message, error)

	SetVectorDBClient(vectorDbClient *rag.VectorDbClient)

	GetVectorDBClient() *rag.VectorDbClient
}

// NewCompanion creates a new Companion instance with the provided configuration.
func NewCompanion(config models.Configuration) AICompanion {
	var client AICompanion
	switch config.ApiProvider {
	case models.Ollama:
		client = &ollama.Companion{
			Config: config,
			SystemRole: models.Message{
				Role:    models.System,
				Content: "You are a helpful assistant",
			},
			Conversation: make([]models.Message, 0),
			Client:       &http.Client{Timeout: time.Second * time.Duration(config.HttpConfig.HTTPClientTimeout)},
		}
	case models.OpenAI:
		client = &openai.Companion{
			Config: config,
			SystemRole: models.Message{
				Role:    models.Developer,
				Content: "You are a helpful assistant",
			},
			Conversation: make([]models.Message, 0),
			Client:       &http.Client{Timeout: time.Second * time.Duration(config.HttpConfig.HTTPClientTimeout)},
		}
	}

	switch config.VectorDBConfig.Type {
	case models.SqlVectorDb:
		vectorClient, _ := rag.NewSQLiteVectorDb(config.VectorDBConfig.Endpoint, true)
		client.SetVectorDBClient(&vectorClient)
	case models.WeaviateDb:
		vectorClient, _ := rag.NewWeaviateClient(config.VectorDBConfig.Endpoint, config.VectorDBConfig.ApiKey)
		client.SetVectorDBClient(&vectorClient)
	default:
	}

	return client
}

// NewDefaultConfig creates a new default configuration with the provided API provider, API token, and model.
func NewDefaultConfig(apiProvider models.ApiProvider, apiToken, chatModel, embeddingModel string, vectorDbType models.VectorDbType, vectorDbUrl, vectorDbToken string) *models.Configuration {
	var config models.Configuration = models.Configuration{
		ApiProvider: apiProvider,
		ApiKey:      apiToken,
		AiModels: models.AiModels{
			ChatModel:      models.Model{Model: chatModel, Name: chatModel},
			EmbeddingModel: models.Model{Model: embeddingModel, Name: embeddingModel},
		},
		HttpConfig: models.HttpConfiguration{
			MaxInputLength:    500,
			HTTPClientTimeout: 300,
			BufferSize:        2048,
		},
		MaxMessages: 20,
		Color:       terminal.Green,
		Output:      true,
	}

	config.SystemPrompt = "You are a helpful assistant"
	config.EnrichmentPrompt = "answer following query with the provided context:\nuser query: %s\ncontext: %s"

	var apiEndpoints models.ApiEndpointUrls
	switch apiProvider {
	case models.Ollama:
		apiEndpoints = models.ApiEndpointUrls{
			ApiChatURL:       "http://localhost:11434/api/chat",
			ApiGenerateURL:   "http://localhost:11434/api/generate",
			ApiEmbedURL:      "http://localhost:11434/api/embed",
			ApiModerationURL: "http://localhost:11434/api/generate",
			ApiModelsURL:     "http://localhost:11434/api/tags",
		}

	case models.OpenAI:
		apiEndpoints = models.ApiEndpointUrls{
			ApiChatURL:       "https://api.openai.com/v1/chat/completions",
			ApiGenerateURL:   "https://api.openai.com/v1/completions",
			ApiEmbedURL:      "https://api.openai.com/v1/embeddings",
			ApiModerationURL: "https://api.openai.com/v1/moderations",
			ApiModelsURL:     "https://api.openai.com/v1/models",
		}
	}

	config.VectorDBConfig = models.VectorDbConfiguration{
		Type:     vectorDbType,
		Endpoint: vectorDbUrl,
		ApiKey:   vectorDbToken,
	}

	config.ApiEndpoints = apiEndpoints

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
