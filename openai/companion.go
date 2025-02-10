package openai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
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

func (companion *Companion) Trace(payload string) {
	if companion.Config.Terminal.Trace {
		fmt.Println(payload)
	}
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

// GetConfig returns the current configuration of the companion.
func (companion *Companion) GetConfig() models.Configuration {
	return companion.Config
}

// SetConfig sets a new configuration for the companion.
func (companion *Companion) SetConfig(config models.Configuration) {
	companion.Config = config
	companion.SetSystemRole(config.Prompt.SystemPrompt)
}

// GetCurrentSystemRole returns the current system role of the companion.
func (companion *Companion) GetSystemRole() models.Message {
	return companion.SystemRole
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

func (companion *Companion) SetVectorDB(vectorDbClient *vectordb.VectorDb) {
	companion.VectorDb = vectorDbClient
}

func (companion *Companion) GetVectorDB() *vectordb.VectorDb {
	return companion.VectorDb
}

// SetCurrentSystemRole sets a new system role for the companion.
func (companion *Companion) SetSystemRole(prompt string) {
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
func (companion *Companion) PrepareConversation(message models.Message, includeStrategy models.IncludeStrategy) []models.Message {
	messages := append([]models.Message{companion.SystemRole}, companion.PrepareArray(companion.Conversation, includeStrategy)...)
	messages = append(messages, message)

	return messages
}

// PrepareArray prepares an array of messages based on the includeStrategy.
func (companion *Companion) PrepareArray(messages []models.Message, includeStrategy models.IncludeStrategy) []models.Message {
	var newarray []models.Message
	for _, msg := range messages {
		switch includeStrategy {
		case models.IncludeAssistant:
			{
				if msg.Role == models.Assistant {
					newarray = append(newarray, msg)
				}
			}
		case models.IncludeUser:
			{
				if msg.Role == models.User {
					newarray = append(newarray, msg)
				}
			}
		case models.IncludeBoth:
			{
				newarray = append(newarray, msg)
			}
		default:
			{
				newarray = append(newarray, msg)
			}
		}
	}

	if len(newarray) > companion.Config.MaxMessages {
		newarray = newarray[len(newarray)-companion.Config.MaxMessages:]
	}

	return newarray
}

// createMessage creates a new models.Message with the given role and content.
func (companion *Companion) CreateMessage(role models.Role, input string) models.Message {
	var message models.Message = models.Message{
		Role:    role,
		Content: input,
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

// addmodels.Message adds the given models.Message to the conversation history.
func (companion *Companion) AddMessage(message models.Message) {
	companion.Conversation = append(companion.Conversation, message)
}

// ClearLine clears the current line if output is enabled in the configuration
func (companion *Companion) ClearLine() {
	if companion.Config.Terminal.Output {
		fmt.Print(terminal.ClearLine)
	}
}

// Print prints the given content to the console with color and reset.
func (companion *Companion) Print(content string) {
	if companion.Config.Terminal.Output {
		fmt.Printf("%s%s%s", companion.Config.Terminal.Color, content, terminal.Reset)
	}
}

// Println prints the given content to the console with color and a newline character, then resets the color.
func (companion *Companion) Println(content string) {
	if companion.Config.Terminal.Output {
		fmt.Printf("%s%s%s\n", companion.Config.Terminal.Color, content, terminal.Reset)
	}
}

// PrintError prints an error message to the console in red.
func (companion *Companion) PrintError(err error) {
	fmt.Printf("%s%v%s\n", terminal.Red, err, terminal.Reset)
}

// SendEmbeddingRequest sends a request to the OpenAI API to generate embeddings for a given text input.
func (companion *Companion) SendEmbeddingRequest(embedding models.EmbeddingRequest) (models.EmbeddingResponse, error) {
	var embeddingResponse models.EmbeddingResponse

	// Marshal the payload into JSON
	payloadBytes, err := json.Marshal(embedding)
	if err != nil {
		companion.PrintError(err)
		return embeddingResponse, err
	}
	companion.Trace(fmt.Sprintf("SendEmbeddingRequest: payload: %s", string(payloadBytes)))

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
	companion.Debug(fmt.Sprintf("SendEmbeddingRequest: StatusCode %d, Status %s", resp.StatusCode, resp.Status))

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
	companion.Trace(fmt.Sprintf("SendEmbeddingRequest: responseBytes: %s", string(responseBytes)))

	var oaiResponse EmbeddingResponse
	err = json.Unmarshal(responseBytes, &oaiResponse)
	if err != nil {
		companion.PrintError(err)
		return embeddingResponse, err
	}

	embeddingResponse = companion.convertToModelEmbeddingResponse(oaiResponse)
	companion.Trace(fmt.Sprintf("SendEmbeddingRequest: embeddingResponse: %v", embeddingResponse))

	return embeddingResponse, nil
}

// convertToModelEmbeddingResponse converts the OpenAI API response to a models.EmbeddingResponse.
func (companion *Companion) convertToModelEmbeddingResponse(response EmbeddingResponse) models.EmbeddingResponse {
	var embeddings [][]float32
	for _, embedding := range response.Data {
		embeddings = append(embeddings, embedding.Embedding)
	}

	return models.EmbeddingResponse{
		Model:            response.Model,
		Embeddings:       embeddings,
		OriginalResponse: response,
	}
}

// SendModerationRequest sends a request to the OpenAI API to moderate a given text input.
func (companion *Companion) SendModerationRequest(moderationRequest models.ModerationRequest) (models.ModerationResponse, error) {
	var moderationResponse models.ModerationResponse

	// Marshal the payload into JSON
	payloadBytes, err := json.Marshal(moderationRequest)
	if err != nil {
		companion.PrintError(err)
		return moderationResponse, err
	}

	var ctx context.Context
	var cancel context.CancelFunc
	if companion.Config.Terminal.Output {
		ctx, cancel = context.WithCancel(context.Background())
		cs := terminal.NewSpinningCharacter('?', 100, 10)
		cs.StartSpinning(ctx)
		defer cancel()
	}

	companion.Trace(fmt.Sprintf("SendModerationRequest: payload %s", string(payloadBytes)))

	// Create and configure the HTTP request
	req, err := http.NewRequestWithContext(context.Background(), "POST", companion.Config.ApiEndpoints.ApiModerationURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		companion.PrintError(err)
		return moderationResponse, err
	}
	req.Header.Set("Authorization", "Bearer "+companion.Config.ApiKey)
	req.Header.Set("Content-Type", "application/json")

	// Execute the HTTP request
	resp, err := companion.HttpClient.Do(req)
	if err != nil {
		companion.PrintError(err)
		return moderationResponse, err
	}
	defer resp.Body.Close()
	companion.Debug(fmt.Sprintf("SendModerationRequest: StatusCode %d, Status %s", resp.StatusCode, resp.Status))

	if companion.Config.Terminal.Output {
		cancel()
		companion.ClearLine()
	}

	// Process the streaming response
	responseBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		companion.PrintError(err)
		return moderationResponse, err
	}

	companion.Trace(fmt.Sprintf("SendModerationRequest: responseBytes %s", string(responseBytes)))

	var originalResponse ModerationResponse
	err = json.Unmarshal(responseBytes, &originalResponse)
	if err != nil {
		companion.PrintError(err)
		return moderationResponse, err
	}

	moderationResponse = models.ModerationResponse{
		ID:               originalResponse.ID,
		Model:            models.Model{Model: originalResponse.Model},
		OriginalResponse: originalResponse,
	}

	return moderationResponse, nil
}

// SendGenerateRequest sends a request to the OpenAI API to generate a completion for a given prompt.
func (companion *Companion) SendGenerateRequest(message models.MessageRequest, streaming bool, callback func(m models.Message) error) (models.Message, error) {
	return companion.sendCompletionRequest(message, streaming, true, callback)
}

// ProcessUserInput processes the user input by sending it to the API and handling the response.
func (companion *Companion) SendChatRequest(message models.MessageRequest, streaming bool, callback func(m models.Message) error) (models.Message, error) {
	return companion.sendCompletionRequest(message, streaming, false, callback)
}

func (companion *Companion) sendCompletionRequest(message models.MessageRequest, streaming bool, useGeneratePrompt bool, callback func(m models.Message) error) (models.Message, error) {
	var result models.Message
	var payload ChatRequest = ChatRequest{
		Model:    companion.Config.AiModels.ChatModel.Model,
		Messages: companion.PrepareConversation(message.Message, companion.Config.IncludeStrategy),
		Stream:   streaming,
	}

	companion.Debug(fmt.Sprintf("sendCompletionRequest: useGeneratePrompt: %v", useGeneratePrompt))
	if useGeneratePrompt {
		sysmsg := companion.GetSystemRole()
		companion.Debug(fmt.Sprintf("sendCompletionRequest: sysmsg: %v", sysmsg))
		if len(message.Message.AlternatePrompt) > 0 {
			sysmsg = companion.CreateMessage(models.System, message.Message.AlternatePrompt)
		}
		payload.Messages = []models.Message{sysmsg, message.Message}
	}

	// Marshal the payload into JSON
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		companion.PrintError(err)
		return result, err
	}

	companion.Trace(fmt.Sprintf("sendCompletionRequest: payloadBytes: %s", string(payloadBytes)))

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

	companion.Debug(fmt.Sprintf("sendCompletionRequest: StatusCode %d, Status %s", resp.StatusCode, resp.Status))

	if companion.Config.Terminal.Output {
		cancel()
		companion.ClearLine()
	}

	// Process the streaming response
	if streaming {
		result, err = companion.HandleStreamResponse(resp, models.Chat, callback)
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

		companion.Trace(fmt.Sprintf("sendCompletionRequest: bodyBytes: %s", string(bodyBytes)))

		var completionResponse ChatResponse
		err = json.Unmarshal(bodyBytes, &completionResponse)
		if err != nil {
			companion.PrintError(err)
			return result, err
		}

		result = completionResponse.Choices[0].Message
	}

	if !useGeneratePrompt {
		switch message.RetainOriginalMessage {
		case true:
			companion.AddMessage(message.OriginalMessage)
		case false:
			companion.AddMessage(message.Message)
		}

		companion.AddMessage(result)
	}

	return result, nil
}

func (companion *Companion) HandleStreamResponse(resp *http.Response, streamType models.StreamType, callback func(m models.Message) error) (models.Message, error) {
	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			err = fmt.Errorf("unexpected HTTP status: %s, and failed to read body: %v", resp.Status, err)
			companion.PrintError(err)
			return models.Message{}, err
		}
		err = fmt.Errorf("unexpected HTTP status: %s, body: %s", resp.Status, string(bodyBytes))
		companion.PrintError(err)
		return models.Message{}, err
	}

	var message strings.Builder
	var result models.Message
	var finalErr error

	companion.Print("> ")

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		companion.Trace(fmt.Sprintf("HandleStreamResponse: line: %s", line))
		if len(line) == 0 {
			continue
		}

		if line == "[DONE]" {
			break
		}

		line = strings.TrimPrefix(line, "data:")
		var responseObject ChatResponse
		if err := json.Unmarshal([]byte(line), &responseObject); err != nil {
			finalErr = fmt.Errorf("failed to unmarshal line: %v, error: %w", line, err)
			companion.PrintError(finalErr)
			break
		}

		if len(responseObject.Choices) == 0 {
			finalErr = fmt.Errorf("no choices in response")
			companion.PrintError(finalErr)
			break
		}

		choice := responseObject.Choices[0]

		switch streamType {
		case models.Chat:
			msg := companion.CreateAssistantMessage(choice.Delta.Content)
			if callback != nil {
				if err := callback(msg); err != nil {
					finalErr = fmt.Errorf("callback error: %w", err)
					companion.PrintError(finalErr)
					return models.Message{}, finalErr
				}
			}
			message.WriteString(choice.Delta.Content)
			companion.Print(choice.Delta.Content)
		default:
			finalErr = fmt.Errorf("unsupported stream type: %v", streamType)
			companion.PrintError(finalErr)
			return models.Message{}, finalErr
		}

		if choice.FinishReason == "stop" {
			result = companion.CreateAssistantMessage(message.String())
			companion.Println("")
			break
		}
	}

	if err := scanner.Err(); err != nil && err != io.EOF {
		finalErr = fmt.Errorf("scanner error: %w", err)
		companion.PrintError(finalErr)
	}

	return result, finalErr
}

// GetModels retrieves a list of available models from the API.
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

	companion.Debug(fmt.Sprintf("GetModels: StatusCode %d, Status %s", resp.StatusCode, resp.Status))

	if companion.Config.Terminal.Output {
		companion.ClearLine()
	}

	// Process the streaming response
	responseBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		companion.PrintError(err)
		return []models.Model{}, err
	}
	companion.Trace(fmt.Sprintf("GetModels: responseBytes: %s", responseBytes))

	var originalResponse ModelResponse
	err = json.Unmarshal(responseBytes, &originalResponse)
	if err != nil {
		companion.PrintError(fmt.Errorf("GetModels: Unmarshal error: %v", err))
		return []models.Model{}, err
	}

	companion.Trace(fmt.Sprintf("GetModels: originalResponse: length: %d, %v", len(originalResponse.Models), originalResponse))

	var transformedModels []models.Model
	for i, model := range originalResponse.Models {
		companion.Trace(fmt.Sprintf("GetModels: transforming model: %d", i))
		var transformedModel models.Model = models.Model{
			Model: model.ID,
			Name:  model.ID,
		}

		transformedModels = append(transformedModels, transformedModel)
	}

	companion.Trace(fmt.Sprintf("GetModels: transformedModels: %v", transformedModels))

	return transformedModels, nil
}

// RunFunction executes a function with the provided payload.
func (companion *Companion) RunFunction(function models.Function, payload models.FunctionPayload) (models.FunctionResponse, error) {
	result := models.FunctionResponse{}

	payloadBytes, err := json.Marshal(payload.Parameters)
	if err != nil {
		companion.PrintError(err)
		return result, err
	}

	// Create and configure the HTTP request
	req, err := http.NewRequestWithContext(context.Background(), "POST", function.Endpoint, bytes.NewBuffer(payloadBytes))
	if err != nil {
		companion.PrintError(err)
		return result, err
	}
	req.Header.Set("Authorization", "Bearer "+function.ApiKey)
	req.Header.Set("Content-Type", "application/json")

	// Execute the HTTP request
	resp, err := companion.HttpClient.Do(req)
	if err != nil {
		companion.PrintError(err)
		return result, err
	}
	defer resp.Body.Close()
	companion.Trace(fmt.Sprintf("RunFunction: payload %s", string(payloadBytes)))
	companion.Debug(fmt.Sprintf("RunFunction: StatusCode %d, Status %s", resp.StatusCode, resp.Status))

	responseBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		companion.PrintError(err)
		return result, err
	}

	companion.Trace(fmt.Sprintf("RunFunction: responseBytes %s", string(responseBytes)))

	err = json.Unmarshal(responseBytes, &result)
	if err != nil {
		companion.PrintError(err)
		return result, err
	}
	return result, nil
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
