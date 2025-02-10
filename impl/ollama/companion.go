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
func (companion *Companion) PrepareConversation(message models.Message, includeStrategy models.IncludeStrategy) []models.Message {
	messages := append([]models.Message{companion.SystemRole}, sideKick.PrepareArray(companion.Conversation, includeStrategy, companion.Config.MaxMessages)...)
	messages = append(messages, message)

	return messages
}

// addMessage adds the given message to the conversation history.
func (companion *Companion) AddMessage(message models.Message) {
	companion.Conversation = append(companion.Conversation, message)
}

// SendModerationRequest sends a request to the OpenAI API to moderate a given text input.
func (companion *Companion) SendModerationRequest(moderationRequest models.ModerationRequest) (models.ModerationResponse, error) {
	return models.ModerationResponse{}, errors.New("unsupported")
}

// SendEmbeddingRequest sends an embedding request to the server using the provided embedding request object.
func (companion *Companion) SendEmbeddingRequest(embedding models.EmbeddingRequest) (models.EmbeddingResponse, error) {
	var embeddingResponse models.EmbeddingResponse

	// Marshal the payload into JSON
	payloadBytes, err := json.Marshal(embedding)
	if err != nil {
		sideKick.Error(err)
		return embeddingResponse, err
	}

	sideKick.Trace(fmt.Sprintf("SendEmbeddingRequest: payload %s", string(payloadBytes)), companion.Config.Terminal)

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

	sideKick.Trace(fmt.Sprintf("SendEmbeddingRequest: responseBytes %s", string(responseBytes)), companion.Config.Terminal)

	var originalResponse EmbeddingResponse
	err = json.Unmarshal(responseBytes, &originalResponse)
	if err != nil {
		sideKick.Error(err)
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
	sideKick.Trace(fmt.Sprintf("parameters:\nmessage: %v\nstreaming: %v\n", message, streaming), companion.Config.Terminal)
	sideKick.Trace(fmt.Sprintf("message.message.content: %s\n", message.Message.Content), companion.Config.Terminal)
	var result models.Message
	var payload CompletionRequest = CompletionRequest{
		Model:    string(companion.Config.AiModels.ChatModel.Model),
		Messages: companion.PrepareConversation(message.Message, companion.Config.IncludeStrategy),
		Stream:   streaming,
	}

	// Marshal the payload into JSON
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		sideKick.Error(err)
		return result, err
	}
	sideKick.Trace(fmt.Sprintf("SendChatRequest: payloadBytes: %s", string(payloadBytes)), companion.Config.Terminal)

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

	sideKick.Debug(fmt.Sprintf("SendChatRequest: StatusCode %d, Status %s", resp.StatusCode, resp.Status), companion.Config.Terminal)

	if companion.Config.Terminal.Output {
		cancel()
		sideKick.ClearLine(companion.Config.Terminal)
	}

	// Process the streaming response
	if streaming {
		result, err = companion.HandleStreamResponse(resp, models.Chat, callback)
		if err != nil {
			sideKick.Error(err)
		}
	} else {
		var bodyBytes []byte
		bodyBytes, err = io.ReadAll(resp.Body)
		if err != nil {
			sideKick.Error(err)
			return result, nil
		}

		sideKick.Trace(fmt.Sprintf("SendChatRequest: bodyBytes: %s", string(bodyBytes)), companion.Config.Terminal)

		var completionResponse CompletionResponse
		err = json.Unmarshal(bodyBytes, &completionResponse)
		if err != nil {
			sideKick.Error(err)
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
		sideKick.Error(err)
		return result, err
	}

	sideKick.Trace(fmt.Sprintf("SendGenerateRequest: payloadBytes: %s", string(payloadBytes)), companion.Config.Terminal)

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

	sideKick.Debug(fmt.Sprintf("SendGenerateRequest: StatusCode %d, Status %s", resp.StatusCode, resp.Status), companion.Config.Terminal)

	if companion.Config.Terminal.Output {
		cancel()
		sideKick.ClearLine(companion.Config.Terminal)
	}

	// Process the streaming response
	if streaming {
		result, err = companion.HandleStreamResponse(resp, models.Generate, callback)
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

		sideKick.Trace(fmt.Sprintf("SendGenerateRequest: payloadBytes: %s", string(payloadBytes)), companion.Config.Terminal)

		var completionResponse CompletionResponse
		err = json.Unmarshal(bodyBytes, &completionResponse)
		if err != nil {
			sideKick.Error(err)
			return result, err
		}

		result = sideKick.CreateAssistantMessage(completionResponse.Response)
	}

	return result, nil
}

// HandleStreamResponse handles the streaming response from the Ollama API.
func (companion *Companion) HandleStreamResponse(resp *http.Response, streamType models.StreamType, callback func(m models.Message) error) (models.Message, error) {
	var message strings.Builder
	var result models.Message

	sideKick.Debug(fmt.Sprintf("HandleStreamResponse: resp.StatusCode: %d, status: %s", resp.StatusCode, resp.Status), companion.Config.Terminal)
	if resp.StatusCode != http.StatusOK {
		bodyBytes, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			err := fmt.Errorf("unexpected HTTP status: %s, and failed to read body: %v", resp.Status, readErr)
			sideKick.Error(err)
			resp.Body.Close()
			return models.Message{}, err
		}
		err := fmt.Errorf("unexpected HTTP status: %s, body: %s", resp.Status, string(bodyBytes))
		sideKick.Error(err)
		resp.Body.Close()
		return models.Message{}, err
	}
	defer resp.Body.Close()

	sideKick.Print("> ", companion.Config.Terminal)

	scanner := bufio.NewScanner(resp.Body)

OuterLoop:
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		sideKick.Trace(fmt.Sprintf("HandleStreamResponse: line: %s", line), companion.Config.Terminal)
		if len(line) == 0 {
			continue
		}

		var responseObject CompletionResponse
		if err := json.Unmarshal([]byte(line), &responseObject); err != nil {
			sideKick.Error(err)
			return models.Message{}, err // Fail fast on unmarshaling error
		}

		switch streamType {
		case models.Chat:
			// Print the content from each choice in the chunk
			message.WriteString(responseObject.Message.Content)
			if callback != nil {
				if err := callback(responseObject.Message); err != nil {
					sideKick.Error(err)
					return models.Message{}, err
				}
			}
			sideKick.Print(responseObject.Message.Content, companion.Config.Terminal)
		case models.Generate:
			// Print the content from each choice in the chunk
			message.WriteString(responseObject.Response)
			if callback != nil {
				msg := sideKick.CreateAssistantMessage(responseObject.Response)
				if err := callback(msg); err != nil {
					sideKick.Error(err)
					return models.Message{}, err
				}
			}
			sideKick.Print(responseObject.Response, companion.Config.Terminal)
		default:
			err := fmt.Errorf("unsupported stream type: %v", streamType)
			sideKick.Error(err)
			return models.Message{}, err
		}

		if responseObject.Done {
			result = sideKick.CreateAssistantMessage(message.String())
			sideKick.Println("", companion.Config.Terminal)
			break OuterLoop
		}
	}

	if err := scanner.Err(); err != nil && err != io.EOF {
		sideKick.Error(err)
		return models.Message{}, err
	}

	return result, nil
}

// GetModels returns a list of available models from the API.
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
		sideKick.Error(err)
		return []models.Model{}, err
	}

	return originalResponse.Models, nil
}

// RunFunction executes a function with the provided payload.
func (companion *Companion) RunFunction(function models.Function, payload models.FunctionPayload) (models.FunctionResponse, error) {
	return sideKick.RunFunction(companion.HttpClient, function, payload, companion.Config.Terminal.Debug, companion.Config.Terminal.Trace)
}
