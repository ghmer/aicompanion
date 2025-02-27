package aicompanion

import (
	"encoding/base64"
	"net/http"
	"time"

	"github.com/ghmer/aicompanion/impl/ollama"
	"github.com/ghmer/aicompanion/impl/openai"
	sidekick_interface "github.com/ghmer/aicompanion/interfaces/sidekick"
	"github.com/ghmer/aicompanion/models"
	"github.com/ghmer/aicompanion/terminal"
)

const (
	SystemPrompt        = "You are a helpful assistant"
	EnrichmentPrompt    = "Answer the following query with the provided context"
	SummarizationPrompt = "Summarize the given conversation in 3 to 6 words without using punctuation marks, emojis or formatting. Only return the summary and nothing else."

	DefaultHTTPTimeout = 300
	DefaultMaxMessages = 20
)

var OllamaEndpoints = models.ApiEndpointUrls{
	ApiChatURL:       "http://localhost:11434/api/chat",
	ApiGenerateURL:   "http://localhost:11434/api/generate",
	ApiEmbedURL:      "http://localhost:11434/api/embed",
	ApiModerationURL: "http://localhost:11434/api/generate",
	ApiModelsURL:     "http://localhost:11434/api/tags",
}

var OpenAIEndpoints = models.ApiEndpointUrls{
	ApiChatURL:       "https://api.openai.com/v1/chat/completions",
	ApiGenerateURL:   "https://api.openai.com/v1/completions",
	ApiEmbedURL:      "https://api.openai.com/v1/embeddings",
	ApiModerationURL: "https://api.openai.com/v1/moderations",
	ApiModelsURL:     "https://api.openai.com/v1/models",
}

// AICompanion defines the interface for interacting with AI models.
type AICompanion interface {
	// PrepareConversation prepares the conversation by appending system role and current conversation messages.
	PrepareConversation(message models.Message, includeStrategy models.IncludeStrategy) []models.Message

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

	// GetEnrichmentPrompt returns the current enrichment prompt
	GetEnrichmentPrompt() string

	// SetEnrichmentPrompt sets a new enrichment prompt
	SetEnrichmentPrompt(prompt string)

	// GetSummarizationPrompt returns the current summarization prompt
	GetSummarizationPrompt() string

	// SetSummarizationPrompt sets a summarization prompt
	SetSummarizationPrompt(prompt string)

	// GetConversation returns the current conversation
	GetConversation() []models.Message

	// SetConversation sets the current conversation
	SetConversation(conversation []models.Message)

	// GetClient returns the current HTTP client used for requests
	GetHttpClient() *http.Client

	// SetClient sets a new HTTP client for requests
	SetHttpClient(client *http.Client)

	/*
		// SetVectorDB sets the vector database instance.
		SetVectorDB(vectorDb *vectordb.VectorDb)

		// GetVectorDb returns the current vector database instance.
		GetVectorDB() *vectordb.VectorDb
	*/

	// interactions
	// GetModels returns all models that the endpoint supports
	GetModels() ([]models.Model, error)

	// SendChatRequest sends a chat request to an AI model and returns a response message
	SendChatRequest(message models.MessageRequest, streaming bool, callback func(m models.Message) error) (models.Message, error)

	// SendCompletionRequest sends a completion request to an AI model and returns a response message
	SendGenerateRequest(message models.MessageRequest, streaming bool, callback func(m models.Message) error) (models.Message, error)

	// SendEmbeddingRequest sends an embedding request to an AI model and returns a response
	SendEmbeddingRequest(embedding models.EmbeddingRequest) (models.EmbeddingResponse, error)

	// SendModerationRequest sends a moderation request to an AI model and returns a response
	SendModerationRequest(moderationRequest models.ModerationRequest) (models.ModerationResponse, error)

	// HandleStreamResponse handles streaming responses from an HTTP request.
	HandleStreamResponse(resp *http.Response, streamType models.StreamType, callback func(m models.Message) error) (models.Message, error)

	SendToolRequest(message models.MessageRequest) (models.Message, error)

	// RunFunction runs a function and returns the response
	RunFunction(tool models.Tool, payload models.FunctionPayload) (models.FunctionResponse, error)
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
				Content: config.ActivePersona.Prompt.SystemPrompt,
			},
			Conversation: make([]models.Message, 0),
			HttpClient:   &http.Client{Timeout: time.Second * time.Duration(config.HttpConfig.HTTPClientTimeout)},
		}
	case models.OpenAI:
		client = &openai.Companion{
			Config: config,
			SystemRole: models.Message{
				Role:    models.System,
				Content: config.ActivePersona.Prompt.SystemPrompt,
			},
			Conversation: make([]models.Message, 0),
			HttpClient:   &http.Client{Timeout: time.Second * time.Duration(config.HttpConfig.HTTPClientTimeout)},
		}
	}

	return client
}

// NewDefaultConfig creates a new default configuration with the provided API provider, API token, and model.
func NewDefaultConfig(apiProvider models.ApiProvider, apiToken, chatModel, generateModel, embeddingModel string) *models.Configuration {
	var config models.Configuration = models.Configuration{
		ApiProvider: apiProvider,
		ApiKey:      apiToken,
		AiModels: models.AiModels{
			ChatModel:      models.Model{Model: chatModel, Name: chatModel},
			GenerateModel:  models.Model{Model: generateModel, Name: generateModel},
			EmbeddingModel: models.Model{Model: embeddingModel, Name: embeddingModel},
		},
		HttpConfig: models.HttpConfiguration{
			HTTPClientTimeout: DefaultHTTPTimeout,
		},
		MaxMessages:     DefaultMaxMessages,
		IncludeStrategy: models.IncludeBoth,
		Terminal: models.Terminal{
			Color:  terminal.Green,
			Output: false,
			Debug:  false,
			Trace:  false,
		},
	}

	persona := models.Persona{
		Name: "default",
		Prompt: models.Prompt{
			SystemPrompt:        SystemPrompt,
			EnrichmentPrompt:    EnrichmentPrompt,
			SummarizationPrompt: SummarizationPrompt,
		},
		Knowledge: []string{},
	}

	config.ActivePersona = persona
	config.Personas = []models.Persona{persona}

	var apiEndpoints models.ApiEndpointUrls
	switch apiProvider {
	case models.Ollama:
		apiEndpoints = OllamaEndpoints

	case models.OpenAI:
		apiEndpoints = OpenAIEndpoints
	}

	config.ApiEndpoints = apiEndpoints

	config.RAGQueryOptions = models.VectorDBQueryOptions{
		Limit:               0,
		SimilarityThreshold: 0.0,
	}

	return &config
}

// ReadImageFromFile reads an image from the specified filepath and returns a Base64 encoded image.
func ReadImageFromFile(filepath string) (models.Base64Image, error) {
	sidekick := sidekick_interface.NewSideKick()
	content, err := sidekick.ReadFile(filepath)
	if err != nil {
		return models.Base64Image{}, err
	}

	return models.Base64Image{
		Data: base64.StdEncoding.EncodeToString(content),
	}, nil
}
