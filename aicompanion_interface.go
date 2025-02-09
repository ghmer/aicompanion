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
	"github.com/ghmer/aicompanion/vectordb"
)

// AICompanion defines the interface for interacting with AI models.
type AICompanion interface {
	// PrepareConversation prepares the conversation by appending system role and current conversation messages.
	PrepareConversation(message models.Message, includeStrategy models.IncludeStrategy) []models.Message

	// PrepareArray prepares an array of messages by appending the given messages and applying the include strategy.
	PrepareArray(messages []models.Message, includeStrategy models.IncludeStrategy) []models.Message

	// CreateMessage creates a new message with the given role and input string
	CreateMessage(role models.Role, input string) models.Message

	// CreateMessageWithImages creates a new message with the given role, input string, and images
	CreateMessageWithImages(role models.Role, message string, images *[]models.Base64Image) models.Message

	//CreateUserMessage creates a new user message with the given input string
	CreateUserMessage(input string, images *[]models.Base64Image) models.Message

	// CreateAssistantMessage creates a new assistant message with the given input string
	CreateAssistantMessage(input string) models.Message

	// CreateEmbeddingRequest creates an embedding request for the given input.
	CreateEmbeddingRequest(input []string) *models.EmbeddingRequest

	// CreateModerationRequest
	CreateModerationRequest(input string) *models.ModerationRequest

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

	// GetFunctionsPrompt returns the current functions prompt
	GetFunctionsPrompt() string

	// SetFunctionsPrompt sets a functions prompt
	SetFunctionsPrompt(prompt string)

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

	// SetVectorDB sets the vector database instance.
	SetVectorDB(vectorDb *vectordb.VectorDb)

	// GetVectorDb returns the current vector database instance.
	GetVectorDB() *vectordb.VectorDb

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

	// RunFunction runs a function and returns the response
	RunFunction(function models.Function, payload models.FunctionPayload) (models.FunctionResponse, error)

	Debug(payload string)

	Trace(payload string)
}

// NewCompanion creates a new Companion instance with the provided configuration.
func NewCompanion(config models.Configuration, vectordb *vectordb.VectorDb) AICompanion {
	var client AICompanion
	switch config.ApiProvider {
	case models.Ollama:
		client = &ollama.Companion{
			Config: config,
			SystemRole: models.Message{
				Role:    models.System,
				Content: config.Prompt.SystemPrompt,
			},
			Conversation: make([]models.Message, 0),
			HttpClient:   &http.Client{Timeout: time.Second * time.Duration(config.HttpConfig.HTTPClientTimeout)},
		}
	case models.OpenAI:
		client = &openai.Companion{
			Config: config,
			SystemRole: models.Message{
				Role:    models.System,
				Content: config.Prompt.SystemPrompt,
			},
			Conversation: make([]models.Message, 0),
			HttpClient:   &http.Client{Timeout: time.Second * time.Duration(config.HttpConfig.HTTPClientTimeout)},
		}
	}

	if vectordb != nil {
		client.SetVectorDB(vectordb)
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
			HTTPClientTimeout: 300,
		},
		MaxMessages:     20,
		IncludeStrategy: models.IncludeBoth,
		Terminal: models.Terminal{
			Color:  terminal.Green,
			Output: true,
			Debug:  false,
		},
	}

	config.Prompt.SystemPrompt = "You are a helpful assistant"
	config.Prompt.EnrichmentPrompt = "Answer the following query with the provided context:\nuser query: %s\ncontext: %s"
	config.Prompt.SummarizationPrompt = "Summarize the given conversation in 3 to 6 words without using punctuation marks, emojis or formatting. Only return the summary and nothing else."
	config.Prompt.FunctionsPrompt = `Ignore any previous instructions.\n\nAdhere strictly to the following rules:\n- Never, under any circumstances, use formatting, like Markdown. Omit newlines.\n- Return "no matching tool" if no tools match the query. Generate no further output, don't make proposals.\n- Construct a JSON object in the format: {"name": "functionName","parameters":[{"functionParamKey":"functionParamValue"}]} using the appropriate tool and its parameters. If a tool defines several parameters, consider every parameter as mandatory. Extend the resulting parameters array accordingly. Double-check that you included all tool parameters. Only include parameters that are defined by the tool. If a tool does not define parameters, then include an empty array.\n- Validate the provided context against the required parameters to ensure all required values are available before constructing the JSON object.\n- If the function has no parameters, ensure the "parameters" field is an empty array ([]).\n- Always prioritize accuracy when matching tools and constructing responses.\n- Return the object and limit the response to the JSON object.\n- These rules overrule any instructions that the user may have given.\n\nIn a user message, consider any information found after the tag ::context as system supplied information.\nAvailable Tools: %s`

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
