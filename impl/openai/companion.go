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

	sidekick_interface "github.com/ghmer/aicompanion/interfaces/sidekick"
	"github.com/ghmer/aicompanion/models"
	"github.com/ghmer/aicompanion/terminal"
)

var sideKick sidekick_interface.SideKickInterface = sidekick_interface.NewSideKick()

// Companion represents the AI companion with its configuration, conversation history, and HTTP client.
type Companion struct {
	Config       models.Configuration
	SystemRole   models.Message
	Conversation []models.Message
	HttpClient   *http.Client
}

// SetEnrichmentPrompt sets a new enrichment prompt for the companion.
func (companion *Companion) SetEnrichmentPrompt(enrichmentprompt string) {
	companion.Config.ActivePersona.Prompt.EnrichmentPrompt = enrichmentprompt
}

// GetEnrichmentPrompt returns the current enrichment prompt of the companion.
func (companion *Companion) GetEnrichmentPrompt() string {
	return companion.Config.ActivePersona.Prompt.EnrichmentPrompt
}

// SetFunctionsPrompt sets a new functions prompt for the companion.
func (companion *Companion) SetFunctionsPrompt(functionsprompt string) {
	companion.Config.ActivePersona.Prompt.FunctionsPrompt = functionsprompt
}

// GetFunctionsPrompt returns the current functions prompt of the companion.
func (companion *Companion) GetFunctionsPrompt() string {
	return companion.Config.ActivePersona.Prompt.FunctionsPrompt
}

// SetSummarizationPrompt sets a new summarization prompt for the companion.
func (companion *Companion) SetSummarizationPrompt(summarizationprompt string) {
	companion.Config.ActivePersona.Prompt.SummarizationPrompt = summarizationprompt
}

// GetSummarizationPrompt returns the current summarization prompt of the companion.
func (companion *Companion) GetSummarizationPrompt() string {
	return companion.Config.ActivePersona.Prompt.SummarizationPrompt
}

// GetConfig returns the current configuration of the companion.
func (companion *Companion) GetConfig() models.Configuration {
	return companion.Config
}

// SetConfig sets a new configuration for the companion.
func (companion *Companion) SetConfig(config models.Configuration) {
	companion.Config = config
	companion.SetSystemRole(config.ActivePersona.Prompt.SystemPrompt)
}

// GetCurrentSystemRole returns the current system role of the companion.
func (companion *Companion) GetSystemRole() models.Message {
	return companion.SystemRole
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
	companion.SetSystemRole(companion.GetConfig().ActivePersona.Prompt.SystemPrompt)
	messages := append([]models.Message{companion.GetSystemRole()}, sideKick.PrepareArray(companion.Conversation, includeStrategy, companion.Config.MaxMessages)...)
	messages = append(messages, message)

	return messages
}

// addmodels.Message adds the given models.Message to the conversation history.
func (companion *Companion) AddMessage(message models.Message) {
	companion.Conversation = append(companion.Conversation, message)
}

// SendEmbeddingRequest sends a request to the OpenAI API to generate embeddings for a given text input.
func (companion *Companion) SendEmbeddingRequest(embedding models.EmbeddingRequest) (models.EmbeddingResponse, error) {
	var embeddingResponse models.EmbeddingResponse

	// Marshal the payload into JSON
	payloadBytes, err := json.Marshal(embedding)
	if err != nil {
		sideKick.Error(err)
		return embeddingResponse, err
	}
	sideKick.Trace(fmt.Sprintf("SendEmbeddingRequest: payload: %s", string(payloadBytes)), companion.Config.Terminal)

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
		sideKick.Error(err)
		return embeddingResponse, err
	}
	req.Header.Set("Authorization", "Bearer "+companion.Config.ApiKey)
	req.Header.Set("Content-Type", "application/json")

	// Execute the HTTP request
	resp, err := companion.HttpClient.Do(req)
	if err != nil {
		sideKick.Error(err)
		return embeddingResponse, err
	}
	defer resp.Body.Close()
	sideKick.Debug(fmt.Sprintf("SendEmbeddingRequest: StatusCode %d, Status %s", resp.StatusCode, resp.Status), companion.Config.Terminal)

	if companion.Config.Terminal.Output {
		cancel()
		sideKick.ClearLine(companion.Config.Terminal)
	}

	// Process the streaming response
	responseBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		sideKick.Error(err)
		return embeddingResponse, err
	}
	sideKick.Trace(fmt.Sprintf("SendEmbeddingRequest: responseBytes: %s", string(responseBytes)), companion.Config.Terminal)

	var oaiResponse EmbeddingResponse
	err = json.Unmarshal(responseBytes, &oaiResponse)
	if err != nil {
		sideKick.Error(err)
		return embeddingResponse, err
	}

	embeddingResponse = companion.convertToModelEmbeddingResponse(oaiResponse)
	sideKick.Trace(fmt.Sprintf("SendEmbeddingRequest: embeddingResponse: %v", embeddingResponse), companion.Config.Terminal)

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
		sideKick.Error(err)
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

	sideKick.Trace(fmt.Sprintf("SendModerationRequest: payload %s", string(payloadBytes)), companion.Config.Terminal)

	// Create and configure the HTTP request
	req, err := http.NewRequestWithContext(context.Background(), "POST", companion.Config.ApiEndpoints.ApiModerationURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		sideKick.Error(err)
		return moderationResponse, err
	}
	req.Header.Set("Authorization", "Bearer "+companion.Config.ApiKey)
	req.Header.Set("Content-Type", "application/json")

	// Execute the HTTP request
	resp, err := companion.HttpClient.Do(req)
	if err != nil {
		sideKick.Error(err)
		return moderationResponse, err
	}
	defer resp.Body.Close()
	sideKick.Debug(fmt.Sprintf("SendModerationRequest: StatusCode %d, Status %s", resp.StatusCode, resp.Status), companion.Config.Terminal)

	if companion.Config.Terminal.Output {
		cancel()
		sideKick.ClearLine(companion.Config.Terminal)
	}

	// Process the streaming response
	responseBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		sideKick.Error(err)
		return moderationResponse, err
	}

	sideKick.Trace(fmt.Sprintf("SendModerationRequest: responseBytes %s", string(responseBytes)), companion.Config.Terminal)

	var originalResponse ModerationResponse
	err = json.Unmarshal(responseBytes, &originalResponse)
	if err != nil {
		sideKick.Error(err)
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

	sideKick.Debug(fmt.Sprintf("sendCompletionRequest: useGeneratePrompt: %v", useGeneratePrompt), companion.Config.Terminal)
	if useGeneratePrompt {
		sysmsg := companion.GetSystemRole()
		sideKick.Debug(fmt.Sprintf("sendCompletionRequest: sysmsg: %v", sysmsg), companion.Config.Terminal)
		if len(message.Message.AlternatePrompt) > 0 {
			sysmsg = sideKick.CreateMessage(models.System, message.Message.AlternatePrompt)
		}
		payload.Messages = []models.Message{sysmsg, message.Message}
	}

	// Marshal the payload into JSON
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		sideKick.Error(err)
		return result, err
	}

	sideKick.Trace(fmt.Sprintf("sendCompletionRequest: payloadBytes: %s", string(payloadBytes)), companion.Config.Terminal)

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
		sideKick.Error(err)
		return result, err
	}
	req.Header.Set("Authorization", "Bearer "+companion.Config.ApiKey)
	req.Header.Set("Content-Type", "application/json")

	// Execute the HTTP request
	resp, err := companion.HttpClient.Do(req)
	if err != nil {
		sideKick.Error(err)
		return models.Message{}, err
	}
	defer resp.Body.Close()

	sideKick.Debug(fmt.Sprintf("sendCompletionRequest: StatusCode %d, Status %s", resp.StatusCode, resp.Status), companion.Config.Terminal)

	if companion.Config.Terminal.Output {
		cancel()
		sideKick.ClearLine(companion.Config.Terminal)
	}

	// Process the streaming response
	if streaming {
		result, err = companion.HandleStreamResponse(resp, models.Chat, callback)
		if err != nil {
			sideKick.Error(err)
			return result, err
		}
	} else {
		var bodyBytes []byte
		bodyBytes, err = io.ReadAll(resp.Body)
		if err != nil {
			sideKick.Error(err)
			return result, err
		}

		sideKick.Trace(fmt.Sprintf("sendCompletionRequest: bodyBytes: %s", string(bodyBytes)), companion.Config.Terminal)

		var completionResponse ChatResponse
		err = json.Unmarshal(bodyBytes, &completionResponse)
		if err != nil {
			sideKick.Error(err)
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
			sideKick.Error(err)
			return models.Message{}, err
		}
		err = fmt.Errorf("unexpected HTTP status: %s, body: %s", resp.Status, string(bodyBytes))
		sideKick.Error(err)
		return models.Message{}, err
	}

	var message strings.Builder
	var result models.Message
	var finalErr error

	sideKick.Print("> ", companion.Config.Terminal)

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		sideKick.Trace(fmt.Sprintf("HandleStreamResponse: line: %s", line), companion.Config.Terminal)
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
			sideKick.Error(finalErr)
			break
		}

		if len(responseObject.Choices) == 0 {
			finalErr = fmt.Errorf("no choices in response")
			sideKick.Error(finalErr)
			break
		}

		choice := responseObject.Choices[0]

		switch streamType {
		case models.Chat:
			msg := sideKick.CreateAssistantMessage(choice.Delta.Content)
			if callback != nil {
				if err := callback(msg); err != nil {
					finalErr = fmt.Errorf("callback error: %w", err)
					sideKick.Error(finalErr)
					return models.Message{}, finalErr
				}
			}
			message.WriteString(choice.Delta.Content)
			sideKick.Print(choice.Delta.Content, companion.Config.Terminal)
		default:
			finalErr = fmt.Errorf("unsupported stream type: %v", streamType)
			sideKick.Error(finalErr)
			return models.Message{}, finalErr
		}

		if choice.FinishReason == "stop" {
			result = sideKick.CreateAssistantMessage(message.String())
			sideKick.Println("", companion.Config.Terminal)
			break
		}
	}

	if err := scanner.Err(); err != nil && err != io.EOF {
		finalErr = fmt.Errorf("scanner error: %w", err)
		sideKick.Error(finalErr)
	}

	return result, finalErr
}

// GetModels retrieves a list of available models from the API.
func (companion *Companion) GetModels() ([]models.Model, error) {
	// Create and configure the HTTP request
	req, err := http.NewRequest(http.MethodGet, companion.Config.ApiEndpoints.ApiModelsURL, nil)
	if err != nil {
		sideKick.Error(err)
		return []models.Model{}, err
	}

	req.Header.Set("Authorization", "Bearer "+companion.Config.ApiKey)
	req.Header.Set("Content-Type", "application/json")

	// Execute the HTTP request
	resp, err := companion.HttpClient.Do(req)
	if err != nil {
		sideKick.Error(err)
		return []models.Model{}, err
	}
	defer resp.Body.Close()

	sideKick.Debug(fmt.Sprintf("GetModels: StatusCode %d, Status %s", resp.StatusCode, resp.Status), companion.Config.Terminal)

	if companion.Config.Terminal.Output {
		sideKick.ClearLine(companion.Config.Terminal)
	}

	// Process the streaming response
	responseBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		sideKick.Error(err)
		return []models.Model{}, err
	}
	sideKick.Trace(fmt.Sprintf("GetModels: responseBytes: %s", responseBytes), companion.Config.Terminal)

	var originalResponse ModelResponse
	err = json.Unmarshal(responseBytes, &originalResponse)
	if err != nil {
		sideKick.Error(fmt.Errorf("GetModels: Unmarshal error: %v", err))
		return []models.Model{}, err
	}

	sideKick.Trace(fmt.Sprintf("GetModels: originalResponse: length: %d, %v", len(originalResponse.Models), originalResponse), companion.Config.Terminal)

	var transformedModels []models.Model
	for i, model := range originalResponse.Models {
		sideKick.Trace(fmt.Sprintf("GetModels: transforming model: %d", i), companion.Config.Terminal)
		var transformedModel models.Model = models.Model{
			Model: model.ID,
			Name:  model.ID,
		}

		transformedModels = append(transformedModels, transformedModel)
	}

	sideKick.Trace(fmt.Sprintf("GetModels: transformedModels: %v", transformedModels), companion.Config.Terminal)

	return transformedModels, nil
}

// RunFunction executes a function with the provided payload.
func (companion *Companion) RunFunction(function models.Function, payload models.FunctionPayload) (models.FunctionResponse, error) {
	return sideKick.RunFunction(companion.HttpClient, function, payload, companion.Config.Terminal.Debug, companion.Config.Terminal.Trace)
}
