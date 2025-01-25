package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/ghmer/aicompanion/models"
	"github.com/ghmer/aicompanion/rag"
	"github.com/ghmer/aicompanion/terminal"
)

// Companion represents the AI companion with its configuration, conversation history, and HTTP client.
type Companion struct {
	Config         models.Configuration
	SystemRole     models.Message
	Conversation   []models.Message
	Client         *http.Client
	VectorDbClient *rag.VectorDbClient
}

// GetConfig returns the current configuration of the companion.
func (companion *Companion) GetConfig() models.Configuration {
	return companion.Config
}

// SetConfig sets a new configuration for the companion.
func (companion *Companion) SetConfig(config models.Configuration) {
	companion.Config = config
}

// CreateUserMessage creates a new user message with the given input string
func (companion *Companion) CreateUserMessage(input string, images []models.Base64Image) models.Message {
	if images != nil || len(images) > 0 {
		return companion.CreateMessageWithImages(models.User, input, images)
	}
	return companion.CreateMessage(models.User, input)
}

// CreateAssistantMessage creates a new assistant message with the given input string
func (companion *Companion) CreateAssistantMessage(input string) models.Message {
	return companion.CreateMessage(models.Assistant, input)
}

func (companion *Companion) SetVectorDBClient(vectorDbClient *rag.VectorDbClient) {
	companion.VectorDbClient = vectorDbClient
}

func (companion *Companion) GetVectorDBClient() *rag.VectorDbClient {
	return companion.VectorDbClient
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

// prepareConversation prepares the conversation by appending system role and current conversation messages.
func (companion *Companion) PrepareConversation() []models.Message {
	messages := append([]models.Message{companion.SystemRole}, companion.Conversation...)
	if len(messages) > companion.Config.MaxMessages {
		messages = messages[len(messages)-companion.Config.MaxMessages:]
	}

	return messages
}

// createMessage creates a new message with the given role and content.
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

// addMessage adds the given message to the conversation history.
func (companion *Companion) AddMessage(message models.Message) {
	companion.Conversation = append(companion.Conversation, message)
}

// ClearLine clears the current line if output is enabled in the configuration.
func (companion *Companion) ClearLine() {
	if companion.Config.Output {
		fmt.Print(terminal.ClearLine)
	}
}

// Print prints the given content with the specified color and reset code if output is enabled in the configuration.
func (companion *Companion) Print(content string) {
	if companion.Config.Output {
		fmt.Printf("%s%s%s", companion.Config.Color, content, terminal.Reset)
	}
}

// Println prints the given content with the specified color and reset code followed by a newline character if output is enabled in the configuration.
func (companion *Companion) Println(content string) {
	if companion.Config.Output {
		fmt.Printf("%s%s%s\n", companion.Config.Color, content, terminal.Reset)
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
	if companion.Config.Output {
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
func (companion *Companion) SendChatRequest(message models.Message, streaming bool, callback func(m models.Message) error) (models.Message, error) {
	companion.AddMessage(message)
	var result models.Message
	var payload CompletionRequest = CompletionRequest{
		Model:    string(companion.Config.AiModels.ChatModel),
		Messages: companion.PrepareConversation(),
		Stream:   streaming,
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
	req, err := http.NewRequestWithContext(context.Background(), "POST", companion.Config.ApiEndpoints.ApiChatURL, bytes.NewBuffer(payloadBytes))
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

		var completionResponse CompletionResponse
		err = json.Unmarshal(bodyBytes, &completionResponse)
		if err != nil {
			companion.PrintError(err)
			return result, nil
		}

		result = completionResponse.Message
	}

	companion.Conversation = append(companion.Conversation, result)

	return result, nil
}

// ProcessUserInput processes the user input by sending it to the API and handling the response.
func (companion *Companion) SendGenerateRequest(message models.Message, streaming bool, callback func(m models.Message) error) (models.Message, error) {
	companion.AddMessage(message)
	var result models.Message
	var payload CompletionRequest = CompletionRequest{
		Model:  string(companion.Config.AiModels.ChatModel),
		Images: message.Images,
		Prompt: message.Content,
		Stream: streaming,
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
	req, err := http.NewRequestWithContext(context.Background(), "POST", companion.Config.ApiEndpoints.ApiGenerateURL, bytes.NewBuffer(payloadBytes))
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

		var completionResponse CompletionResponse
		err = json.Unmarshal(bodyBytes, &completionResponse)
		if err != nil {
			companion.PrintError(err)
			return result, nil
		}

		result = completionResponse.Message
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

	buffer := make([]byte, companion.Config.HttpConfig.BufferSize)
	companion.Print("> ")

	// handle response
	for {
		n, err := resp.Body.Read(buffer) // Read data from the response body into a buffer
		if n > 0 {
			lines := strings.Split(string(buffer[:n]), "\n") // Split the buffer content by newline characters to get individual lines
			for _, line := range lines {
				line = strings.TrimSpace(line) // Remove leading and trailing whitespace from each line
				if len(line) == 0 {
					continue
				}

				var responseObject CompletionResponse
				if err := json.Unmarshal([]byte(line), &responseObject); err != nil {
					companion.PrintError(err)
					companion.Println(line)
					break
				}

				switch streamType {
				case models.Chat:
					// Print the content from each choice in the chunk
					message.WriteString(responseObject.Message.Content)
					if callback != nil {
						if err := callback(responseObject.Message); err != nil {
							companion.PrintError(err)
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
						}
					}
					companion.Print(responseObject.Response)
				}

				if responseObject.Done {
					result = companion.CreateAssistantMessage(message.String())
					companion.Println("")
					break
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
