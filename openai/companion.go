package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/ghmer/aicompanion/models"
	"github.com/ghmer/aicompanion/terminal"
)

// Companion represents the AI companion with its configuration, conversation history, and HTTP client.
type Companion struct {
	Config       models.Configuration
	SystemRole   models.Message
	Conversation []models.Message
	Client       *http.Client
}

// GetConfig returns the current configuration of the companion.
func (companion *Companion) GetConfig() models.Configuration {
	return companion.Config
}

// SetConfig sets a new configuration for the companion.
func (companion *Companion) SetConfig(config models.Configuration) {
	companion.Config = config
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
func (companion *Companion) GetClient() *http.Client {
	return companion.Client
}

// SetClient sets a new HTTP client for the companion.
func (companion *Companion) SetClient(client *http.Client) {
	companion.Client = client
}

// prepareConversation prepares the conversation by appending system role and current conversation models.Messages.
func (companion *Companion) PrepareConversation() []models.Message {
	messages := append([]models.Message{companion.SystemRole}, companion.Conversation...)
	if len(messages) > companion.Config.MaxMessages {
		messages = messages[len(messages)-companion.Config.MaxMessages:]
	}

	return messages
}

// createmodels.Message creates a new models.Message with the given role and content.
func (companion *Companion) CreateMessage(role models.Role, input string) models.Message {
	var message models.Message = models.Message{
		Role:    role,
		Content: input,
	}

	return message
}

// CreateMessageWithImages creates a new message with the given role, content and images
func (companion *Companion) CreateMessageWithImages(role models.Role, input string, images []models.Base64Image) models.Message {
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
	if companion.Config.Output {
		fmt.Print(terminal.ClearLine)
	}
}

// Print prints the given content to the console with color and reset.
func (companion *Companion) Print(content string) {
	if companion.Config.Output {
		fmt.Printf("%s%s%s", companion.Config.Color, content, terminal.Reset)
	}
}

// Println prints the given content to the console with color and a newline character, then resets the color.
func (companion *Companion) Println(content string) {
	if companion.Config.Output {
		fmt.Printf("%s%s%s\n", companion.Config.Color, content, terminal.Reset)
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

	var ctx context.Context
	var cancel context.CancelFunc
	if companion.Config.Output {
		ctx, cancel = context.WithCancel(context.Background())
		cs := terminal.NewSpinningCharacter('?', 100, 10)
		cs.StartSpinning(ctx)
		defer cancel()
	}

	// Create and configure the HTTP request
	req, err := http.NewRequestWithContext(context.Background(), "POST", companion.Config.ApiEmbedURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		companion.PrintError(err)
		return embeddingResponse, err
	}
	req.Header.Set("Authorization", "Bearer "+companion.Config.ApiKey)
	req.Header.Set("Content-Type", "application/json")

	// Execute the HTTP request
	resp, err := companion.Client.Do(req)
	if err != nil {
		companion.PrintError(err)
		return embeddingResponse, err
	}
	defer resp.Body.Close()

	if companion.Config.Output {
		cancel()
		companion.ClearLine()
	}

	// Process the streaming response
	responseBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		companion.PrintError(err)
		return embeddingResponse, err
	}

	var oaiResponse EmbeddingResponse
	err = json.Unmarshal(responseBytes, &oaiResponse)
	if err != nil {
		companion.PrintError(err)
		return embeddingResponse, err
	}

	embeddingResponse = companion.convertToModelEmbeddingResponse(oaiResponse)

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
	if companion.Config.Output {
		ctx, cancel = context.WithCancel(context.Background())
		cs := terminal.NewSpinningCharacter('?', 100, 10)
		cs.StartSpinning(ctx)
		defer cancel()
	}

	// Create and configure the HTTP request
	req, err := http.NewRequestWithContext(context.Background(), "POST", companion.Config.ApiModerationURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		companion.PrintError(err)
		return moderationResponse, err
	}
	req.Header.Set("Authorization", "Bearer "+companion.Config.ApiKey)
	req.Header.Set("Content-Type", "application/json")

	// Execute the HTTP request
	resp, err := companion.Client.Do(req)
	if err != nil {
		companion.PrintError(err)
		return moderationResponse, err
	}
	defer resp.Body.Close()

	if companion.Config.Output {
		cancel()
		companion.ClearLine()
	}

	// Process the streaming response
	responseBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		companion.PrintError(err)
		return moderationResponse, err
	}

	var originalResponse ModerationResponse
	err = json.Unmarshal(responseBytes, &originalResponse)
	if err != nil {
		companion.PrintError(err)
		return moderationResponse, err
	}

	moderationResponse = models.ModerationResponse{
		ID:               originalResponse.ID,
		Model:            originalResponse.Model,
		OriginalResponse: originalResponse,
	}

	return moderationResponse, nil
}

// SendGenerateRequest sends a request to the OpenAI API to generate a completion for a given prompt.
func (companion *Companion) SendGenerateRequest(message models.Message, streaming bool, callback func(m models.Message) error) (models.Message, error) {
	companion.AddMessage(message)
	var result models.Message
	var payload CompletionsRequest = CompletionsRequest{
		Model:  string(companion.Config.AiModel),
		Prompt: message.Content,
		Stream: true,
	}

	// Marshal the payload into JSON
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		companion.PrintError(err)
		return result, err
	}

	var ctx context.Context
	var cancel context.CancelFunc
	if companion.Config.Output {
		ctx, cancel = context.WithCancel(context.Background())
		cs := terminal.NewSpinningCharacter('?', 100, 10)
		cs.StartSpinning(ctx)
		defer cancel()
	}

	// Create and configure the HTTP request
	req, err := http.NewRequestWithContext(context.Background(), "POST", companion.Config.ApiGenerateURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		companion.PrintError(err)
		return result, err
	}
	req.Header.Set("Authorization", "Bearer "+companion.Config.ApiKey)
	req.Header.Set("Content-Type", "application/json")

	// Execute the HTTP request
	resp, err := companion.Client.Do(req)
	if err != nil {
		companion.PrintError(err)
		return models.Message{}, err
	}
	defer resp.Body.Close()

	if companion.Config.Output {
		cancel()
		companion.ClearLine()
	}

	// Process the streaming response
	result, err = companion.HandleStreamResponse(resp, models.Chat, callback)
	if err != nil {
		companion.PrintError(err)
	}
	companion.Conversation = append(companion.Conversation, result)

	return result, nil
}

// ProcessUserInput processes the user input by sending it to the API and handling the response.
func (companion *Companion) SendChatRequest(message models.Message, streaming bool, callback func(m models.Message) error) (models.Message, error) {
	companion.AddMessage(message)
	var result models.Message
	var payload ChatRequest = ChatRequest{
		Model:    string(companion.Config.AiModel),
		Messages: companion.PrepareConversation(),
		Stream:   true,
	}

	// Marshal the payload into JSON
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		companion.PrintError(err)
		return result, err
	}

	var ctx context.Context
	var cancel context.CancelFunc
	if companion.Config.Output {
		ctx, cancel = context.WithCancel(context.Background())
		cs := terminal.NewSpinningCharacter('?', 100, 10)
		cs.StartSpinning(ctx)
		defer cancel()
	}

	// Create and configure the HTTP request
	req, err := http.NewRequestWithContext(context.Background(), "POST", companion.Config.ApiChatURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		companion.PrintError(err)
		return result, err
	}
	req.Header.Set("Authorization", "Bearer "+companion.Config.ApiKey)
	req.Header.Set("Content-Type", "application/json")

	// Execute the HTTP request
	resp, err := companion.Client.Do(req)
	if err != nil {
		companion.PrintError(err)
		return models.Message{}, err
	}
	defer resp.Body.Close()

	if companion.Config.Output {
		cancel()
		companion.ClearLine()
	}

	// Process the streaming response
	result, err = companion.HandleStreamResponse(resp, models.Chat, callback)
	if err != nil {
		companion.PrintError(err)
	}
	companion.Conversation = append(companion.Conversation, result)

	return result, nil
}

// handleStreamResponse handles the streaming response from the API.
func (companion *Companion) HandleStreamResponse(resp *http.Response, streamType models.StreamType, callback func(m models.Message) error) (models.Message, error) {
	var message strings.Builder
	var result models.Message
	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("unexpected http status: %s", resp.Status)
		companion.PrintError(err)
		return models.Message{}, err
	}

	buffer := make([]byte, companion.Config.BufferSize)
	if companion.Config.Output {
		companion.Print("> ")
	}
	// handle response
Outerloop:
	for {
		n, err := resp.Body.Read(buffer) // Read data from the response body into a buffer
		if n > 0 {
			lines := strings.Split(string(buffer[:n]), "\n") // Split the buffer content by newline characters to get individual lines
			for _, line := range lines {
				line = strings.TrimSpace(line) // Remove leading and trailing whitespace from each line
				if len(line) == 0 {
					continue
				}

				if strings.TrimSpace(line) == "[DONE]" { // Check if the line is "[DONE]"
					break Outerloop
				}

				line = strings.TrimPrefix(line, "data:")

				var responseObject ChatResponse
				if err := json.Unmarshal([]byte(line), &responseObject); err != nil {
					companion.PrintError(err)
					companion.Println(line)
					break
				}

				// Print the content from each choice in the chunk
				msg := companion.CreateMessage(models.Assistant, responseObject.Choices[0].Delta.Content)
				if callback != nil {
					if err := callback(msg); err != nil {
						companion.PrintError(err)
					}
				}
				message.WriteString(responseObject.Choices[0].Delta.Content)
				companion.Print(responseObject.Choices[0].Delta.Content)

				if responseObject.Choices[0].FinishReason == "stop" {
					result = companion.CreateMessage(models.Assistant, message.String())
					companion.Println("")
					break Outerloop
				}
			}
		}
		// Handle EOF and other errors during streaming
		if err == io.EOF {
			break
		} else if err != nil {
			companion.PrintError(err) // Print any error that occurred during streaming
			break
		}
	}

	return result, nil
}
