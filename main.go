package main

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var Config Configuration
var ChosenModel AIModel = DefaultModel
var CurrentSystemRole Message
var Conversation []Message
var AddImages bool = false

func init() {
	file, err := os.ReadFile("./config.json")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = json.Unmarshal(file, &Config)
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}

	CurrentSystemRole = Message{
		Role:    string(System),
		Content: "you are a helpful assistant",
	}

	Conversation = make([]Message, 0)

	clearConsole()
	fmt.Println("Welcome to " + Red + "GucciBot" + Reset)
	printLine('-')

	helpMessage()
	fmt.Println("Let's start by introducing your assistant to its new role.")
	setRole()
}

func main() {
	var images, imagePaths []string
	for {
		// Prompt user for input
		userContent := getUserInput()
		message := createMessage(User, userContent, nil)
		if AddImages {
			message.Images = images
			AddImages = false
		}
		Conversation = append(Conversation, message)

		// Handle special commands
		switch userContent {
		case "!quit":
			quitApplication()
			return
		case "!role":
			setRole()
		case "!help":
			helpMessage()
		case "!images":
			setImages(&images, &imagePaths)
		case "!addimages":
			AddImages = true
		case "!model":
			setModel()
		default:
			// Process user input
			processUserInput(ChosenModel)
		}
	}
}

func setModel() {
	fmt.Println("Current model: " + ChosenModel)
	fmt.Println("Select between following models:")
	fmt.Println(Yellow + "1 " + Reset + "Default Model")
	fmt.Println(Yellow + "2 " + Reset + "Vision Model")
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("To select a model, type the number, or hit enter to accept the current model: ")
	scanner.Scan()
	input := scanner.Text()

	if len(input) == 0 {
		return
	}

	switch strings.ToLower(input)[0] {
	case '1':
		ChosenModel = DefaultModel
	case '2':
		ChosenModel = VisionModel
	}
}

func setImages(images, imagePaths *[]string) {
	scanner := bufio.NewScanner(os.Stdin)

	if len(*imagePaths) > 0 {
		fmt.Println("The following file(s) are currently used in image mode:")
		for i, path := range *imagePaths {
			fmt.Printf(string(Yellow)+"%d. %s\n"+string(Reset), i, path)
		}
		fmt.Println()
		fmt.Print("Do you want to change that? This will remove all existing entries (y/N): ")
		scanner.Scan()
		input := scanner.Text()

		if len(input) == 0 || strings.ToLower(input)[0] == 'n' {
			return
		}

		if strings.ToLower(input)[0] == 'y' {
			*images = []string{}
			*imagePaths = []string{}
		}
	}

	for {
		fmt.Print("Enter the path to a file (or type '!finish' to stop): ")
		scanner.Scan()
		input := scanner.Text()

		// Check if the user wants to finish
		if input == "!finish" {
			break
		}

		// Check if the provided path is valid
		absPath, err := filepath.Abs(input)
		if err != nil {
			fmt.Println("Error resolving the path:", err)
			continue
		}

		if _, err := os.Stat(absPath); os.IsNotExist(err) {
			fmt.Println("File does not exist. Please try again.")
			continue
		}

		file, err := os.ReadFile(absPath)
		if err != nil {
			fmt.Println(err)
		}

		// Add the valid file path to the images slice
		*imagePaths = append(*imagePaths, absPath)
		*images = append(*images, base64.StdEncoding.EncodeToString(file))
		fmt.Println("Added:", absPath)
	}

	fmt.Println("File input finished.")
}

func helpMessage() {
	fmt.Println("The following commands are available:")
	fmt.Println(Yellow + "!help" + Reset + " displays this message")
	fmt.Println(Yellow + "!role" + Reset + " changes the assistant's role")
	fmt.Println(Yellow + "!model" + Reset + " changes the model")
	fmt.Println(Yellow + "!images" + Reset + " analyze an image")
	fmt.Println(Yellow + "!addimages" + Reset + " adds the images (defined via !images) to the next request to the AI")
	fmt.Println(Yellow + "!quit" + Reset + " ends the programm")
	fmt.Println()
}

// quitApplication handles quitting the application.
func quitApplication() {
	fmt.Println("Exiting chat. Goodbye!")
}

// setRole prompts the user to input the role of the assistant and returns it.
func setRole() {
	fmt.Println("Describe the new role of your assistant:")
	role := getUserInput()
	fmt.Println("Assistant understood the assignment. Please continue.")

	CurrentSystemRole = Message{
		Role:    string(System),
		Content: role,
	}
}

func prepareConversation() []Message {
	var messages []Message = make([]Message, 0)
	messages = append(messages, CurrentSystemRole)

	for _, message := range Conversation {
		messages = append(messages, message)
	}

	return messages
}

// getUserInput prompts the user with the given message and returns sanitized input.
func getUserInput() string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(Red + "> " + Reset)
	input, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("Error reading input: %v\n", err)
		return ""
	}
	return sanitizeInput(strings.TrimSpace(input))
}

// getUserInput prompts the user with the given message and returns sanitized input.
func createMessage(role Role, input string, images *[]string) Message {
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

// sanitizeInput validates and sanitizes user input.
func sanitizeInput(input string) string {
	if len(input) > Config.MaxInputLength {
		fmt.Println("Input exceeds maximum allowed length.")
		return ""
	}
	return input
}

func readFile(filepath string) string {
	file, err := os.ReadFile(filepath)
	if err != nil {
		fmt.Println(err)
		return ""
	}

	return base64.StdEncoding.EncodeToString(file)
}

// processUserInput sends the user's input along with the assistant role to the API and handles the response.
func processUserInput(model AIModel) {
	var payload RequestPayload = RequestPayload{
		Messages: prepareConversation(),
		Stream:   true,
	}

	switch model {
	case VisionModel:
		payload.Model = string(VisionModel)
	default:
		payload.Model = string(DefaultModel)
	}

	// Marshal payload into JSON
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("Error marshaling payload: %v\n", err)
		return
	}

	// Create and configure the HTTP request
	req, err := http.NewRequest("POST", Config.DefaultAPIURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return
	}
	req.Header.Set("Authorization", "Bearer "+Config.BearerToken)
	req.Header.Set("Content-Type", "application/json")

	// Execute the HTTP request
	client := &http.Client{Timeout: time.Second * time.Duration(Config.HTTPClientTimeout)}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error making HTTP request: %v\n", err)
		return
	}
	defer resp.Body.Close()

	// Process the streaming response
	switch Config.SelectedResponseType {
	case OpenAI:
		handleOpenAIStreamResponse(resp)
	case Ollama:
		handleOllamaStreamResponse(resp)
	}
}

func handleOllamaStreamResponse(resp *http.Response) {
	var message strings.Builder
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Unexpected HTTP status: %s\n", resp.Status)
		return
	}

	buffer := make([]byte, Config.BufferSize)
	fmt.Print(Green + "> " + Reset)
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
					message := createMessage(Assistant, message.String(), nil)
					Conversation = append(Conversation, message)
					fmt.Println()
					return
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
}

// handleStreamResponse processes the streaming response from the API.
func handleOpenAIStreamResponse(resp *http.Response) {
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Unexpected HTTP status: %s\n", resp.Status)
		return
	}

	buffer := make([]byte, Config.BufferSize)
	fmt.Print(Green + "> " + Reset)
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
					fmt.Printf("\n\n")
					return
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
}
