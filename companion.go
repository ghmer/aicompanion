package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type Companion struct {
	Config            Configuration
	AiModel           AIModel
	CurrentSystemRole Message
	Conversation      []Message
	Color             TermColor
	Client            *http.Client
}

func NewCompanion(config Configuration, aimodel AIModel, color TermColor) *Companion {
	return &Companion{
		Config:  config,
		AiModel: aimodel,
		CurrentSystemRole: Message{
			Role:    "system",
			Content: "You are a helpful assistant",
		},
		Conversation: make([]Message, 0),
		Color:        color,
		Client:       &http.Client{Timeout: time.Second * time.Duration(config.HTTPClientTimeout)},
	}
}

func (companion *Companion) prepareConversation() []Message {
	messages := append([]Message{companion.CurrentSystemRole}, companion.Conversation...)
	if len(messages) > 10 {
		messages = messages[len(messages)-10:]
	}
	return messages
}

// getUserInput prompts the user with the given message and returns sanitized input.
func (companion *Companion) createMessage(role Role, input string, images *[]string) Message {
	var message Message
	message = Message{
		Role:    string(role),
		Content: input,
	}

	if images != nil {
		message.Images = *images
	}

	return message
}

func (companion *Companion) readFile(filepath string) string {
	file, err := os.ReadFile(filepath)
	if err != nil {
		fmt.Println(err)
		return ""
	}

	return base64.StdEncoding.EncodeToString(file)
}

func (companion *Companion) addMessage(message Message) {
	companion.Conversation = append(companion.Conversation, message)
}

// processUserInput sends the user's input along with the assistant role to the API and handles the response.
func (companion *Companion) ProcessUserInput(message Message) (Message, error) {
	companion.addMessage(message)
	var result Message
	var payload RequestPayload = RequestPayload{
		Model:    string(companion.AiModel),
		Messages: companion.prepareConversation(),
		Stream:   true,
	}

	// Marshal payload into JSON
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("Error marshaling payload: %v\n", err)
		return result, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	cs := NewSpinningCharacter('?', 100, 10)
	cs.StartSpinning(ctx)
	defer cancel()

	// Create and configure the HTTP request
	req, err := http.NewRequestWithContext(context.Background(), "POST", companion.Config.DefaultAPIURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return result, err
	}
	req.Header.Set("Authorization", "Bearer "+companion.Config.BearerToken)
	req.Header.Set("Content-Type", "application/json")

	// Execute the HTTP request
	resp, err := companion.Client.Do(req)
	if err != nil {
		fmt.Printf("Error making HTTP request: %v\n", err)
		return Message{}, err
	}
	defer resp.Body.Close()

	cancel()
	fmt.Print(ClearLine)

	// Process the streaming response
	result, err = companion.handleStreamResponse(resp)
	if err != nil {
		fmt.Println(err)
	}
	companion.Conversation = append(companion.Conversation, result)

	return result, nil
}

func (companion *Companion) handleStreamResponse(resp *http.Response) (Message, error) {
	var message strings.Builder
	var result Message
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Unexpected HTTP status: %s\n", resp.Status)
		return Message{}, errors.New("unexpected HTTP status")
	}

	buffer := make([]byte, companion.Config.BufferSize)
	fmt.Print(companion.Color + "> " + Reset)
	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			lines := strings.Split(string(buffer[:n]), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if len(line) == 0 {
					continue
				}

				switch companion.Config.SelectedResponseType {
				case Ollama:
					companion.handleOllamaStreamResponse(line, &message, &result)
				case OpenAI:
					companion.handleOpenAIStreamResponse(line, &message, &result)
				}

			}
		}
		// Handle EOF and other errors during streaming
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Printf("Error reading response: %v\n", err)
			break
		}
	}

	return result, nil
}

func (companion *Companion) handleOllamaStreamResponse(line string, message *strings.Builder, result *Message) error {
	var responseObject ResponsePayload
	if err := json.Unmarshal([]byte(line), &responseObject); err != nil {
		fmt.Printf("Error parsing JSON: %v\n", err)
		fmt.Println(line)
		return err
	}

	// Print the content from each choice in the chunk
	message.WriteString(responseObject.Message.Content)
	fmt.Print(responseObject.Message.Content)
	if responseObject.Done {
		*result = companion.createMessage(Assistant, message.String(), nil)
		fmt.Println()
		return nil
	}

	return nil
}

func (companion *Companion) handleOpenAIStreamResponse(line string, message *strings.Builder, result *Message) error {
	var responseObject ResponsePayload
	if err := json.Unmarshal([]byte(line), &responseObject); err != nil {
		fmt.Printf("Error parsing JSON: %v\n", err)
		fmt.Println(line)
		return err
	}

	// Print the content from each choice in the chunk
	message.WriteString(responseObject.Message.Content)
	fmt.Print(responseObject.Message.Content)
	if responseObject.Done {
		*result = companion.createMessage(Assistant, message.String(), nil)
		fmt.Println()
		return nil
	}

	return nil
}

/*

func (companion *Companion) handleOllamaStreamResponse(resp *http.Response) (Message, error) {
	var message strings.Builder
	var result Message
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Unexpected HTTP status: %s\n", resp.Status)
		return Message{}, errors.New("unexpected HTTP status")
	}

	buffer := make([]byte, companion.Config.BufferSize)
	fmt.Print(companion.Color + "> " + Reset)
	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			lines := strings.Split(string(buffer[:n]), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if len(line) == 0 {
					continue
				}
				var responseObject ResponsePayload
				if err := json.Unmarshal([]byte(line), &responseObject); err != nil {
					fmt.Printf("Error parsing JSON: %v\n", err)
					fmt.Println(line)
					continue
				}

				// Print the content from each choice in the chunk
				message.WriteString(responseObject.Message.Content)
				fmt.Print(responseObject.Message.Content)
				if responseObject.Done {
					result = companion.createMessage(Assistant, message.String(), nil)
					companion.Conversation = append(companion.Conversation, result)
					fmt.Println()
					return result, err
				}

			}
		}
		// Handle EOF and other errors during streaming
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Printf("Error reading response: %v\n", err)
			break
		}
	}

	return result, nil
}

// handleStreamResponse processes the streaming response from the API.
func (companion *Companion) handleOpenAIStreamResponse(resp *http.Response) (Message, error) {
	var message strings.Builder
	var result Message
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Unexpected HTTP status: %s\n", resp.Status)
		return Message{}, errors.New("unexpected HTTP status")
	}

	buffer := make([]byte, companion.Config.BufferSize)
	fmt.Print(companion.Color + "> " + Reset)
	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			lines := strings.Split(string(buffer[:n]), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line == "" {
					continue
				}

				// Check for stream end signal
				if line == "data: [DONE]" {
					result = companion.createMessage(Assistant, message.String(), nil)
					companion.Conversation = append(companion.Conversation, result)
					fmt.Printf("\n\n")
					return result, err
				}

				// Process JSON data prefixed by "data: "
				if strings.HasPrefix(line, "data: ") {
					jsonData := strings.TrimPrefix(line, "data: ")
					var chunk ResponseChunk
					if err := json.Unmarshal([]byte(jsonData), &chunk); err != nil {
						fmt.Printf("Error parsing JSON: %v\n", err)
						continue
					}

					// Print the content from each choice in the chunk
					for _, choice := range chunk.Choices {
						message.WriteString(choice.Delta.Content)
						fmt.Print(choice.Delta.Content)
					}
				}
			}
		}
		// Handle EOF and other errors during streaming
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Printf("Error reading response: %v\n", err)
			break
		}
	}

	return result, nil
}

*/
