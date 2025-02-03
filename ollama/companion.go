package ollama

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/ghmer/aicompanion/models"
	"github.com/ghmer/aicompanion/terminal"
	"github.com/ghmer/aicompanion/vectordb"
)

// Companion represents the AI companion with its configuration, conversation history, and HTTP client.
type Companion struct {
	Config       models.Configuration
	SystemRole   models.Message
	Conversation []models.Message
	HttpClient   *http.Client
	VectorDb     *vectordb.VectorDb
}

func (companion *Companion) Debug(payload string) {
	if companion.Config.Terminal.Debug {
		fmt.Println(payload)
	}
}

// GetConfig returns the current configuration of the companion.
func (companion *Companion) GetConfig() models.Configuration {
	return companion.Config
}

// SetConfig sets a new configuration for the companion.
func (companion *Companion) SetConfig(config models.Configuration) {
	companion.Config = config
	companion.SetSystemRole(config.Prompt.SystemPrompt)
}

// SetEnrichmentPrompt sets a new enrichment prompt for the companion.
func (companion *Companion) SetEnrichmentPrompt(enrichmentprompt string) {
	companion.Config.Prompt.EnrichmentPrompt = enrichmentprompt
}

// GetEnrichmentPrompt returns the current enrichment prompt of the companion.
func (companion *Companion) GetEnrichmentPrompt() string {
	return companion.Config.Prompt.EnrichmentPrompt
}

// SetFunctionsPrompt sets a new functions prompt for the companion.
func (companion *Companion) SetFunctionsPrompt(functionsprompt string) {
	companion.Config.Prompt.FunctionsPrompt = functionsprompt
}

// GetFunctionsPrompt returns the current functions prompt of the companion.
func (companion *Companion) GetFunctionsPrompt() string {
	return companion.Config.Prompt.FunctionsPrompt
}

// SetSummarizationPrompt sets a new summarization prompt for the companion.
func (companion *Companion) SetSummarizationPrompt(summarizationprompt string) {
	companion.Config.Prompt.SummarizationPrompt = summarizationprompt
}

// GetSummarizationPrompt returns the current summarization prompt of the companion.
func (companion *Companion) GetSummarizationPrompt() string {
	return companion.Config.Prompt.SummarizationPrompt
}

// CreateUserMessage creates a new user message with the given input string
func (companion *Companion) CreateUserMessage(input string, images *[]models.Base64Image) models.Message {
	if images != nil && len(*images) > 0 {
		return companion.CreateMessageWithImages(models.User, input, images)
	}
	return companion.CreateMessage(models.User, input)
}

// CreateAssistantMessage creates a new assistant message with the given input string
func (companion *Companion) CreateAssistantMessage(input string) models.Message {
	return companion.CreateMessage(models.Assistant, input)
}

// SetVectorDBClient sets a new vector database client for the companion.
func (companion *Companion) SetVectorDB(vectorDb *vectordb.VectorDb) {
	companion.VectorDb = vectorDb
}

// GetVectorDBClient returns the current vector database client of the companion.
func (companion *Companion) GetVectorDB() *vectordb.VectorDb {
	return companion.VectorDb
}

// GetCurrentSystemRole returns the current system role of the companion.
func (companion *Companion) GetSystemRole() models.Message {
	return companion.SystemRole
}

// SetCurrentSystemRole sets a new system role for the companion.
func (companion *Companion) SetSystemRole(prompt string) {
	companion.Config.Prompt.SystemPrompt = prompt

	var role models.Message = models.Message{
		Role:    models.System,
		Content: prompt,
	}
	companion.SystemRole = role
}

// GetConversation returns the current conversation history of the companion.
func (companion *Companion) GetConversation() []models.Message {
	return companion.Conversation
}

// SetConversation sets a new conversation history for the companion.
func (companion *Companion) SetConversation(conversation []models.Message) {
	companion.Conversation = conversation
}

// GetClient returns the current HTTP client of the companion.
func (companion *Companion) GetHttpClient() *http.Client {
	return companion.HttpClient
}

// SetClient sets a new HTTP client for the companion.
func (companion *Companion) SetHttpClient(client *http.Client) {
	companion.HttpClient = client
}

// prepareConversation prepares the conversation by appending system role and current conversation messages.
func (companion *Companion) PrepareConversation(message models.Message) []models.Message {
	messages := append([]models.Message{companion.SystemRole}, companion.Conversation...)
	if len(messages) > companion.Config.MaxMessages {
		messages = messages[len(messages)-companion.Config.MaxMessages:]
	}

	messages = append(messages, message)

	return messages
}

// createMessage creates a new message with the given role and content.
func (companion *Companion) CreateMessage(role models.Role, input string) models.Message {
	var message models.Message = models.Message{
		Role:    role,
		Content: input,
		Images:  nil,
	}

	return message
}

// CreateMessageWithImages creates a new message with the given role, content and images
func (companion *Companion) CreateMessageWithImages(role models.Role, input string, images *[]models.Base64Image) models.Message {
	var message models.Message = models.Message{
		Role:    role,
		Content: input,
		Images:  images,
	}

	return message
}

// addMessage adds the given message to the conversation history.
func (companion *Companion) AddMessage(message models.Message) {
	companion.Conversation = append(companion.Conversation, message)
}

// ClearLine clears the current line if output is enabled in the configuration.
func (companion *Companion) ClearLine() {
	if companion.Config.Terminal.Output {
		fmt.Print(terminal.ClearLine)
	}
}

// Print prints the given content with the specified color and reset code if output is enabled in the configuration.
func (companion *Companion) Print(content string) {
	if companion.Config.Terminal.Output {
		fmt.Printf("%s%s%s", companion.Config.Terminal.Color, content, terminal.Reset)
	}
}

// Println prints the given content with the specified color and reset code followed by a newline character if output is enabled in the configuration.
func (companion *Companion) Println(content string) {
	if companion.Config.Terminal.Output {
		fmt.Printf("%s%s%s\n", companion.Config.Terminal.Color, content, terminal.Reset)
	}
}

// PrintError prints an error message with the specified color and reset code followed by a newline character.
func (companion *Companion) PrintError(err error) {
	fmt.Printf("%s%v%s\n", terminal.Red, err, terminal.Reset)
}

// SendModerationRequest sends a request to the OpenAI API to moderate a given text input.
func (companion *Companion) SendModerationRequest(moderationRequest models.ModerationRequest) (models.ModerationResponse, error) {
	return models.ModerationResponse{}, errors.New("unsupported")
}

func (companion *Companion) CreateEmbeddingRequest(input []string) *models.EmbeddingRequest {
	return &models.EmbeddingRequest{
		Model: companion.Config.AiModels.EmbeddingModel.Model,
		Input: input,
	}
}

func (companion *Companion) CreateModerationRequest(input string) *models.ModerationRequest {
	return &models.ModerationRequest{
		Input: input,
	}
}

// SendEmbeddingRequest sends an embedding request to the server using the provided embedding request object.
func (companion *Companion) SendEmbeddingRequest(embedding models.EmbeddingRequest) (models.EmbeddingResponse, error) {
	var embeddingResponse models.EmbeddingResponse

	// Marshal the payload into JSON
	payloadBytes, err := json.Marshal(embedding)
	if err != nil {
		companion.PrintError(err)
		return embeddingResponse, err
	}

	var ctx context.Context
	var cancel context.CancelFunc
	if companion.Config.Terminal.Output {
		ctx, cancel = context.WithCancel(context.Background())
		cs := terminal.NewSpinningCharacter('?', 100, 10)
		cs.StartSpinning(ctx)
		defer cancel()
	}

	// Create and configure the HTTP request
	req, err := http.NewRequestWithContext(context.Background(), "POST", companion.Config.ApiEndpoints.ApiEmbedURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		companion.PrintError(err)
		return embeddingResponse, err
	}
	req.Header.Set("Authorization", "Bearer "+companion.Config.ApiKey)
	req.Header.Set("Content-Type", "application/json")

	// Execute the HTTP request
	resp, err := companion.HttpClient.Do(req)
	if err != nil {
		companion.PrintError(err)
		return embeddingResponse, err
	}
	defer resp.Body.Close()

	if companion.Config.Terminal.Output {
		cancel()
		companion.ClearLine()
	}

	// Process the streaming response
	responseBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		companion.PrintError(err)
		return embeddingResponse, err
	}

	var originalResponse EmbeddingResponse
	err = json.Unmarshal(responseBytes, &originalResponse)
	if err != nil {
		companion.PrintError(err)
		return embeddingResponse, err
	}

	embeddingResponse = models.EmbeddingResponse{
		Model:            originalResponse.Model,
		Embeddings:       originalResponse.Embeddings,
		OriginalResponse: originalResponse,
	}

	return embeddingResponse, nil
}

// ProcessUserInput processes the user input by sending it to the API and handling the response.
func (companion *Companion) SendChatRequest(message models.MessageRequest, streaming bool, callback func(m models.Message) error) (models.Message, error) {
	var result models.Message
	var payload CompletionRequest = CompletionRequest{
		Model:    string(companion.Config.AiModels.ChatModel.Model),
		Messages: companion.PrepareConversation(message.Message),
		Stream:   streaming,
	}

	// Marshal the payload into JSON
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		companion.PrintError(err)
		return result, err
	}
	companion.Debug(fmt.Sprintf("SendChatRequest: payloadBytes: %s", string(payloadBytes)))

	var ctx context.Context
	var cancel context.CancelFunc
	if companion.Config.Terminal.Output {
		ctx, cancel = context.WithCancel(context.Background())
		cs := terminal.NewSpinningCharacter('?', 100, 10)
		cs.StartSpinning(ctx)
		defer cancel()
	}

	// Create and configure the HTTP request
	req, err := http.NewRequestWithContext(context.Background(), "POST", companion.Config.ApiEndpoints.ApiChatURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		companion.PrintError(err)
		return result, err
	}
	req.Header.Set("Authorization", "Bearer "+companion.Config.ApiKey)
	req.Header.Set("Content-Type", "application/json")

	// Execute the HTTP request
	resp, err := companion.HttpClient.Do(req)
	if err != nil {
		companion.PrintError(err)
		return models.Message{}, err
	}
	defer resp.Body.Close()

	if companion.Config.Terminal.Output {
		cancel()
		companion.ClearLine()
	}

	// Process the streaming response
	if streaming {
		result, err = companion.HandleStreamResponse(resp, models.Chat, callback)
		if err != nil {
			companion.PrintError(err)
		}
	} else {
		var bodyBytes []byte
		bodyBytes, err = io.ReadAll(resp.Body)
		if err != nil {
			companion.PrintError(err)
			return result, nil
		}

		companion.Debug(fmt.Sprintf("SendChatRequest: bodyBytes: %s", string(bodyBytes)))

		var completionResponse CompletionResponse
		err = json.Unmarshal(bodyBytes, &completionResponse)
		if err != nil {
			companion.PrintError(err)
			return result, nil
		}

		result = completionResponse.Message
	}
	switch message.RetainOriginalMessage {
	case true:
		companion.AddMessage(message.OriginalMessage)
	case false:
		companion.AddMessage(message.Message)
	}

	companion.AddMessage(result)

	return result, nil
}

// ProcessUserInput processes the user input by sending it to the API and handling the response.
func (companion *Companion) SendGenerateRequest(message models.MessageRequest, streaming bool, callback func(m models.Message) error) (models.Message, error) {
	var result models.Message
	var payload CompletionRequest = CompletionRequest{
		Model:  string(companion.Config.AiModels.GenerateModel.Model),
		Images: message.Message.Images,
		Prompt: message.Message.Content,
		Stream: streaming,
	}

	// Marshal the payload into JSON
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		companion.PrintError(err)
		return result, err
	}

	companion.Debug(fmt.Sprintf("SendChatRequest: payloadBytes: %s", string(payloadBytes)))

	var ctx context.Context
	var cancel context.CancelFunc
	if companion.Config.Terminal.Output {
		ctx, cancel = context.WithCancel(context.Background())
		cs := terminal.NewSpinningCharacter('?', 100, 10)
		cs.StartSpinning(ctx)
		defer cancel()
	}

	// Create and configure the HTTP request
	req, err := http.NewRequestWithContext(context.Background(), "POST", companion.Config.ApiEndpoints.ApiGenerateURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		companion.PrintError(err)
		return result, err
	}
	req.Header.Set("Authorization", "Bearer "+companion.Config.ApiKey)
	req.Header.Set("Content-Type", "application/json")

	// Execute the HTTP request
	resp, err := companion.HttpClient.Do(req)
	if err != nil {
		companion.PrintError(err)
		return models.Message{}, err
	}
	defer resp.Body.Close()

	if companion.Config.Terminal.Output {
		cancel()
		companion.ClearLine()
	}

	// Process the streaming response
	if streaming {
		result, err = companion.HandleStreamResponse(resp, models.Generate, callback)
		if err != nil {
			companion.PrintError(err)
			return result, err
		}
	} else {
		var bodyBytes []byte
		bodyBytes, err = io.ReadAll(resp.Body)
		if err != nil {
			companion.PrintError(err)
			return result, err
		}

		companion.Debug(fmt.Sprintf("SendChatRequest: payloadBytes: %s", string(payloadBytes)))

		var completionResponse CompletionResponse
		err = json.Unmarshal(bodyBytes, &completionResponse)
		if err != nil {
			companion.PrintError(err)
			return result, err
		}

		result = companion.CreateAssistantMessage(completionResponse.Response)
	}

	return result, nil
}

// HandleStreamResponse handles the streaming response from the Ollama API.
func (companion *Companion) HandleStreamResponse(resp *http.Response, streamType models.StreamType, callback func(m models.Message) error) (models.Message, error) {
	var message strings.Builder
	var result models.Message

	companion.Debug(fmt.Sprintf("HandleStreamResponse: resp.StatusCode: %d, status: %s", resp.StatusCode, resp.Status))
	if resp.StatusCode != http.StatusOK {
		bodyBytes, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			err := fmt.Errorf("unexpected HTTP status: %s, and failed to read body: %v", resp.Status, readErr)
			companion.PrintError(err)
			resp.Body.Close()
			return models.Message{}, err
		}
		err := fmt.Errorf("unexpected HTTP status: %s, body: %s", resp.Status, string(bodyBytes))
		companion.PrintError(err)
		resp.Body.Close()
		return models.Message{}, err
	}
	defer resp.Body.Close()

	companion.Print("> ")

	scanner := bufio.NewScanner(resp.Body)

OuterLoop:
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		companion.Debug(fmt.Sprintf("HandleStreamResponse: line: %s", line))
		if len(line) == 0 {
			continue
		}

		var responseObject CompletionResponse
		if err := json.Unmarshal([]byte(line), &responseObject); err != nil {
			companion.PrintError(err)
			return models.Message{}, err // Fail fast on unmarshaling error
		}

		switch streamType {
		case models.Chat:
			// Print the content from each choice in the chunk
			message.WriteString(responseObject.Message.Content)
			if callback != nil {
				if err := callback(responseObject.Message); err != nil {
					companion.PrintError(err)
					return models.Message{}, err
				}
			}
			companion.Print(responseObject.Message.Content)
		case models.Generate:
			// Print the content from each choice in the chunk
			message.WriteString(responseObject.Response)
			if callback != nil {
				msg := companion.CreateAssistantMessage(responseObject.Response)
				if err := callback(msg); err != nil {
					companion.PrintError(err)
					return models.Message{}, err
				}
			}
			companion.Print(responseObject.Response)
		default:
			err := fmt.Errorf("unsupported stream type: %v", streamType)
			companion.PrintError(err)
			return models.Message{}, err
		}

		if responseObject.Done {
			result = companion.CreateAssistantMessage(message.String())
			companion.Println("")
			break OuterLoop
		}
	}

	if err := scanner.Err(); err != nil && err != io.EOF {
		companion.PrintError(err)
		return models.Message{}, err
	}

	return result, nil
}

// GetModels returns a list of available models from the API.
func (companion *Companion) GetModels() ([]models.Model, error) {
	// Create and configure the HTTP request
	req, err := http.NewRequest(http.MethodGet, companion.Config.ApiEndpoints.ApiModelsURL, nil)
	if err != nil {
		companion.PrintError(err)
		return []models.Model{}, err
	}

	req.Header.Set("Authorization", "Bearer "+companion.Config.ApiKey)
	req.Header.Set("Content-Type", "application/json")

	// Execute the HTTP request
	resp, err := companion.HttpClient.Do(req)
	if err != nil {
		companion.PrintError(err)
		return []models.Model{}, err
	}
	defer resp.Body.Close()

	if companion.Config.Terminal.Output {
		companion.ClearLine()
	}

	// Process the streaming response
	responseBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		companion.PrintError(err)
		return []models.Model{}, err
	}

	companion.Debug(fmt.Sprintf("GetModels: responseBytes: %s", responseBytes))

	var originalResponse ModelResponse
	err = json.Unmarshal(responseBytes, &originalResponse)
	if err != nil {
		companion.PrintError(err)
		return []models.Model{}, err
	}

	return originalResponse.Models, nil
}

func (companion *Companion) RunFunction(models.Function) (models.FunctionResponse, error) {
	return models.FunctionResponse{}, errors.New("not implemented")
}
