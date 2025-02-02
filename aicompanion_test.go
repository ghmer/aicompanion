package aicompanion_test

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/ghmer/aicompanion/models"
	"github.com/ghmer/aicompanion/terminal"
	"github.com/ghmer/aicompanion/vectordb"
)

const (
	EmbeddingModel  string = "embedding-model"
	ChatModel       string = "chat-model"
	GenerateModel   string = "generate-model"
	ModerationModel string = "moderation-model"
)

type MockAICompanion struct {
	Config       models.Configuration
	SystemRole   models.Message
	Conversation []models.Message
	HttpClient   *http.Client
	VectorDb     *vectordb.VectorDb
}

// GetConfig returns the current configuration of the companion.
func (companion *MockAICompanion) GetConfig() models.Configuration {
	return companion.Config
}

// SetConfig sets a new configuration for the companion.
func (companion *MockAICompanion) SetConfig(config models.Configuration) {
	companion.Config = config
	companion.SetSystemRole(config.Prompt.SystemPrompt)
}

// SetEnrichmentPrompt sets a new enrichment prompt for the companion.
func (companion *MockAICompanion) SetEnrichmentPrompt(enrichmentprompt string) {
	companion.Config.Prompt.EnrichmentPrompt = enrichmentprompt
}

// GetEnrichmentPrompt returns the current enrichment prompt of the companion.
func (companion *MockAICompanion) GetEnrichmentPrompt() string {
	return companion.Config.Prompt.EnrichmentPrompt
}

// SetFunctionsPrompt sets a new functions prompt for the companion.
func (companion *MockAICompanion) SetFunctionsPrompt(functionsprompt string) {
	companion.Config.Prompt.FunctionsPrompt = functionsprompt
}

// GetFunctionsPrompt returns the current functions prompt of the companion.
func (companion *MockAICompanion) GetFunctionsPrompt() string {
	return companion.Config.Prompt.FunctionsPrompt
}

// SetSummarizationPrompt sets a new summarization prompt for the companion.
func (companion *MockAICompanion) SetSummarizationPrompt(summarizationprompt string) {
	companion.Config.Prompt.SummarizationPrompt = summarizationprompt
}

// GetSummarizationPrompt returns the current summarization prompt of the companion.
func (companion *MockAICompanion) GetSummarizationPrompt() string {
	return companion.Config.Prompt.SummarizationPrompt
}

// CreateUserMessage creates a new user message with the given input string
func (companion *MockAICompanion) CreateUserMessage(input string, images *[]models.Base64Image) models.Message {
	if images != nil && len(*images) > 0 {
		return companion.CreateMessageWithImages(models.User, input, images)
	}
	return companion.CreateMessage(models.User, input)
}

// CreateAssistantMessage creates a new assistant message with the given input string
func (companion *MockAICompanion) CreateAssistantMessage(input string) models.Message {
	return companion.CreateMessage(models.Assistant, input)
}

// SetVectorDBClient sets a new vector database client for the companion.
func (companion *MockAICompanion) SetVectorDB(vectorDb *vectordb.VectorDb) {
	companion.VectorDb = vectorDb
}

// GetVectorDBClient returns the current vector database client of the companion.
func (companion *MockAICompanion) GetVectorDB() *vectordb.VectorDb {
	return companion.VectorDb
}

// GetCurrentSystemRole returns the current system role of the companion.
func (companion *MockAICompanion) GetSystemRole() models.Message {
	return companion.SystemRole
}

// SetCurrentSystemRole sets a new system role for the companion.
func (companion *MockAICompanion) SetSystemRole(prompt string) {
	companion.Config.Prompt.SystemPrompt = prompt

	var role models.Message = models.Message{
		Role:    models.System,
		Content: prompt,
	}
	companion.SystemRole = role
}

// GetConversation returns the current conversation history of the companion.
func (companion *MockAICompanion) GetConversation() []models.Message {
	return companion.Conversation
}

// SetConversation sets a new conversation history for the companion.
func (companion *MockAICompanion) SetConversation(conversation []models.Message) {
	companion.Conversation = conversation
}

// GetClient returns the current HTTP client of the companion.
func (companion *MockAICompanion) GetHttpClient() *http.Client {
	return companion.HttpClient
}

// SetClient sets a new HTTP client for the companion.
func (companion *MockAICompanion) SetHttpClient(client *http.Client) {
	companion.HttpClient = client
}

// prepareConversation prepares the conversation by appending system role and current conversation messages.
func (companion *MockAICompanion) PrepareConversation(message models.Message) []models.Message {
	messages := append([]models.Message{companion.SystemRole}, companion.Conversation...)
	if len(messages) > companion.Config.MaxMessages {
		messages = messages[len(messages)-companion.Config.MaxMessages:]
	}

	messages = append(messages, message)

	return messages
}

// createMessage creates a new message with the given role and content.
func (companion *MockAICompanion) CreateMessage(role models.Role, input string) models.Message {
	var message models.Message = models.Message{
		Role:    role,
		Content: input,
		Images:  nil,
	}

	return message
}

// CreateMessageWithImages creates a new message with the given role, content and images
func (companion *MockAICompanion) CreateMessageWithImages(role models.Role, input string, images *[]models.Base64Image) models.Message {
	var message models.Message = models.Message{
		Role:    role,
		Content: input,
		Images:  images,
	}

	return message
}

// addMessage adds the given message to the conversation history.
func (companion *MockAICompanion) AddMessage(message models.Message) {
	companion.Conversation = append(companion.Conversation, message)
}

// ClearLine clears the current line if output is enabled in the configuration.
func (companion *MockAICompanion) ClearLine() {
	if companion.Config.Terminal.Output {
		fmt.Print(terminal.ClearLine)
	}
}

// Print prints the given content with the specified color and reset code if output is enabled in the configuration.
func (companion *MockAICompanion) Print(content string) {
	if companion.Config.Terminal.Output {
		fmt.Printf("%s%s%s", companion.Config.Terminal.Color, content, terminal.Reset)
	}
}

// Println prints the given content with the specified color and reset code followed by a newline character if output is enabled in the configuration.
func (companion *MockAICompanion) Println(content string) {
	if companion.Config.Terminal.Output {
		fmt.Printf("%s%s%s\n", companion.Config.Terminal.Color, content, terminal.Reset)
	}
}

// PrintError prints an error message with the specified color and reset code followed by a newline character.
func (companion *MockAICompanion) PrintError(err error) {
	fmt.Printf("%s%v%s\n", terminal.Red, err, terminal.Reset)
}

// SendModerationRequest sends a request to the OpenAI API to moderate a given text input.
func (companion *MockAICompanion) SendModerationRequest(moderationRequest models.ModerationRequest) (models.ModerationResponse, error) {
	return models.ModerationResponse{}, errors.New("unsupported")
}

// SendEmbeddingRequest sends an embedding request to the server using the provided embedding request object.
func (companion *MockAICompanion) SendEmbeddingRequest(embedding models.EmbeddingRequest) (models.EmbeddingResponse, error) {
	embeddingResponse := models.EmbeddingResponse{
		Model:            EmbeddingModel,
		Embeddings:       [][]float32{},
		OriginalResponse: nil,
	}

	return embeddingResponse, nil
}

func (mac *MockAICompanion) SendChatRequest(message models.MessageRequest, streaming bool, callback func(m models.Message) error) (models.Message, error) {
	var response models.Message = models.Message{
		Role: models.Assistant, Content: "Hello! I am pleased to meet you",
	}
	if streaming {
		if callback != nil {
			callback(response)
		}
		return models.Message{}, nil
	}
	return response, nil
}

// ProcessUserInput processes the user input by sending it to the API and handling the response.
func (companion *MockAICompanion) SendGenerateRequest(message models.MessageRequest, streaming bool, callback func(m models.Message) error) (models.Message, error) {
	var response models.Message = models.Message{
		Role: models.Assistant, Content: "Hello! This is a generated message",
	}
	if streaming {
		if callback != nil {
			callback(response)
		}
		return models.Message{}, nil
	}
	return response, nil
}

// HandleStreamResponse handles the streaming response from the Ollama API.
func (companion *MockAICompanion) HandleStreamResponse(resp *http.Response, streamType models.StreamType, callback func(m models.Message) error) (models.Message, error) {
	var result models.Message = models.Message{
		Role:    models.Assistant,
		Content: "Hello! I'm an AI assistant. How can I help you today?",
	}

	return result, nil
}

// GetModels returns a list of available models from the API.
func (companion *MockAICompanion) GetModels() ([]models.Model, error) {
	var result []models.Model = []models.Model{
		{Model: ChatModel, Name: ChatModel},
		{Model: GenerateModel, Name: GenerateModel},
		{Model: EmbeddingModel, Name: EmbeddingModel},
		{Model: ModerationModel, Name: ModerationModel},
	}

	return result, nil
}

func (companion *MockAICompanion) RunFunction(models.Function) (models.FunctionResponse, error) {
	return models.FunctionResponse{}, errors.New("not implemented")
}

func TestAICompanion(t *testing.T) {
	companion := &MockAICompanion{}

	t.Run("Test PrepareConversation", func(t *testing.T) {
		msg := models.Message{Role: models.User, Content: "Hello"}
		messages := companion.PrepareConversation(msg)
		if len(messages) != 1 || messages[0].Content != "Hello" {
			t.Errorf("PrepareConversation failed, expected %v, got %v", msg, messages)
		}
	})

	t.Run("Test CreateMessage", func(t *testing.T) {
		role := models.User
		content := "Test message"
		msg := companion.CreateMessage(role, content)
		if msg.Role != role || msg.Content != content {
			t.Errorf("CreateMessage failed, expected role %v and content %v, got role %v and content %v", role, content, msg.Role, msg.Content)
		}
	})

	t.Run("Test CreateMessageWithImages", func(t *testing.T) {
		role := models.User
		content := "Image message"
		images := []models.Base64Image{{Data: "iVBORw0KGgo="}}
		msg := companion.CreateMessageWithImages(role, content, &images)
		if msg.Role != role || msg.Content != content || msg.Images == nil || len(*msg.Images) != 1 {
			t.Errorf("CreateMessageWithImages failed, expected role %v, content %v, and one image", role, content)
		}
	})

	t.Run("Test CreateUserMessage", func(t *testing.T) {
		content := "User message"
		images := []models.Base64Image{}
		msg := companion.CreateUserMessage(content, &images)
		if msg.Role != models.User || msg.Content != content {
			t.Errorf("CreateUserMessage failed, expected role %v and content %v", models.User, content)
		}
	})

	t.Run("Test CreateAssistantMessage", func(t *testing.T) {
		content := "Assistant message"
		msg := companion.CreateAssistantMessage(content)
		if msg.Role != models.Assistant || msg.Content != content {
			t.Errorf("CreateAssistantMessage failed, expected role %v and content %v", models.Assistant, content)
		}
	})

	t.Run("Test AddMessage", func(t *testing.T) {
		msg := models.Message{Role: models.User, Content: "New message"}
		companion.AddMessage(msg)
		if len(companion.Conversation) != 1 || companion.Conversation[0].Content != "New message" {
			t.Errorf("AddMessage failed, got %v", companion.Conversation)
		}
	})

	t.Run("Test GetConfig and SetConfig", func(t *testing.T) {
		config := models.Configuration{ApiProvider: models.Ollama}
		companion.SetConfig(config)
		if companion.GetConfig().ApiProvider != models.Ollama {
			t.Errorf("GetConfig or SetConfig failed, expected ApiProvider %v, got %v", models.Ollama, companion.GetConfig().ApiProvider)
		}
	})

	t.Run("Test GetSystemRole and SetSystemRole", func(t *testing.T) {
		prompt := "you are a helpful assistant"
		companion.SetSystemRole(prompt)
		if companion.GetSystemRole().Content != prompt {
			t.Errorf("GetSystemRole or SetSystemRole failed, expected SystemRole %v, got %v", prompt, companion.GetSystemRole().Content)
		}
	})

	t.Run("Test GetEnrichmentPrompt and SetEnrichmentPrompt", func(t *testing.T) {
		prompt := "Enrichment prompt"
		companion.SetEnrichmentPrompt(prompt)
		if companion.GetEnrichmentPrompt() != prompt {
			t.Errorf("GetEnrichmentPrompt or SetEnrichmentPrompt failed, expected EnrichmentPrompt %v, got %v", prompt, companion.GetEnrichmentPrompt())
		}
	})

	t.Run("Test GetFunctionsPrompt and SetFunctionsPrompt", func(t *testing.T) {
		prompt := "Functions prompt"
		companion.SetFunctionsPrompt(prompt)
		if companion.GetFunctionsPrompt() != prompt {
			t.Errorf("GetFunctionsPrompt or SetFunctionsPrompt failed, expected FunctionsPrompt %v, got %v", prompt, companion.GetFunctionsPrompt())
		}
	})

	t.Run("Test RunFunctions", func(t *testing.T) {
		_, err := companion.RunFunction(models.Function{})
		if err.Error() != "not implemented" {
			t.Errorf("RunFunction failed, expected error %v, got %v", "not implemented", err)
		}
	})

	t.Run("Test SendChatRequest Standard", func(t *testing.T) {
		request := models.MessageRequest{}
		response, err := companion.SendChatRequest(request, false, nil)
		if err != nil || response.Content != "Hello! I am pleased to meet you" {
			t.Errorf("SendChatRequest failed, expected content 'Hello! I am pleased to meet you', got content %v, error: %v", response.Content, err)
		}
	})

	t.Run("Test SendGenerateRequest Standard", func(t *testing.T) {
		request := models.MessageRequest{}
		response, err := companion.SendGenerateRequest(request, false, nil)
		if err != nil || response.Content != "Hello! This is a generated message" {
			t.Errorf("SendChatRequest failed, expected content 'Hello! This is a generated message', got content %v, error: %v", response.Content, err)
		}
	})
}
