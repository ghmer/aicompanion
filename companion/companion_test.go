package companion_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ghmer/aicompanion/companion"
	"github.com/ghmer/aicompanion/models"
	"github.com/ghmer/aicompanion/openai"
)

func TestAICompanion(t *testing.T) {
	// Prepare configuration for the AI Companion
	config := models.Configuration{
		AIType:            models.Chat,
		AiModel:           "mock-model",
		HTTPClientTimeout: 5,
		ApiProvider:       models.OpenAI,
	}

	// Create a new AI Companion instance
	companion := companion.NewCompanion(config)

	t.Run("Test PrepareConversation", func(t *testing.T) {
		messages := companion.PrepareConversation()
		if len(messages) != 0 {
			t.Errorf("Expected empty conversation, got %d messages", len(messages))
		}
	})

	t.Run("Test CreateMessage", func(t *testing.T) {
		message := companion.CreateMessage(models.User, "Hello!")
		if message.Role != models.User {
			t.Errorf("Expected role 'user', got '%s'", message.Role)
		}
		if message.Content != "Hello!" {
			t.Errorf("Expected content 'Hello!', got '%s'", message.Content)
		}
	})

	t.Run("Test CreateMessageWithImages", func(t *testing.T) {
		images := []models.Base64Image{{Data: "mockImageData"}}
		message := companion.CreateMessageWithImages(models.User, "Hello with images!", images)
		if len(message.Images) != 1 {
			t.Errorf("Expected 1 image, got %d", len(message.Images))
		}
	})

	t.Run("Test ReadFile", func(t *testing.T) {
		content := companion.ReadFile("../README.md")
		if content == "" {
			t.Errorf("Expected file content, got empty string")
		}
	})

	t.Run("Test AddMessage", func(t *testing.T) {
		message := models.Message{Role: models.User, Content: "New message"}
		companion.AddMessage(message)
		conversation := companion.GetConversation()
		if len(conversation) != 1 {
			t.Errorf("Expected 1 message in conversation, got %d", len(conversation))
		}
	})

	t.Run("Test GetConfig", func(t *testing.T) {
		if companion.GetConfig().AiModel != config.AiModel {
			t.Errorf("Expected AI model '%s', got '%s'", config.AiModel, companion.GetConfig().AiModel)
		}
	})

	t.Run("Test SetConfig", func(t *testing.T) {
		newConfig := config
		newConfig.AiModel = "updated-model"
		companion.SetConfig(newConfig)
		if companion.GetConfig().AiModel != "updated-model" {
			t.Errorf("Expected updated model 'updated-model', got '%s'", companion.GetConfig().AiModel)
		}
	})

	t.Run("Test GetSystemRole", func(t *testing.T) {
		role := companion.GetSystemRole()
		if role.Content != "You are a helpful assistant" {
			t.Errorf("Expected system role content 'You are a helpful assistant', got '%s'", role.Content)
		}
	})

	t.Run("Test SetSystemRole", func(t *testing.T) {
		companion.SetSystemRole("New system role")
		if companion.GetSystemRole().Content != "New system role" {
			t.Errorf("Expected system role 'New system role', got '%s'", companion.GetSystemRole().Content)
		}
	})

	t.Run("Test SetConversation", func(t *testing.T) {
		newConversation := []models.Message{
			{Role: models.User, Content: "Message 1"},
			{Role: models.Assistant, Content: "Message 2"},
		}
		companion.SetConversation(newConversation)
		conversation := companion.GetConversation()
		if len(conversation) != 2 {
			t.Errorf("Expected conversation length 2, got %d", len(conversation))
		}
		if conversation[0].Content != "Message 1" {
			t.Errorf("Expected first message 'Message 1', got '%s'", conversation[0].Content)
		}
	})

	t.Run("Test GetClient", func(t *testing.T) {
		client := companion.GetClient()
		if client == nil {
			t.Errorf("Expected non-nil HTTP client")
		}
	})

	t.Run("Test SetClient", func(t *testing.T) {
		newClient := &http.Client{Timeout: 10 * time.Second}
		companion.SetClient(newClient)
		if companion.GetClient() != newClient {
			t.Errorf("Expected updated HTTP client")
		}
	})

	t.Run("Test SendCompletionRequest", func(t *testing.T) {
		mockCompletionRequest := models.Message{
			Role:    models.User,
			Content: "generate this prompt",
		}

		mockCompletionResponse := openai.CompletionsResponse{
			ID:      "1",
			Created: 12345678,
			Model:   "mock-model",
			Choices: []openai.Choice{
				{
					Delta: openai.Delta{
						Content: "the generated prompt",
					},
				},
			},
		}

		// Mock server for completion API
		mockCompletionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
				return
			}

			// Log incoming request for debugging
			body := new(bytes.Buffer)
			_, err := body.ReadFrom(r.Body)
			if err != nil {
				t.Errorf("Error reading request body: %v", err)
				http.Error(w, "Failed to read request", http.StatusInternalServerError)
				return
			}
			t.Logf("Received request: %s", body.String())

			bytes, err := json.Marshal(mockCompletionResponse)
			if err != nil {
				t.Errorf("Error marshalling mock response: %v", err)
				http.Error(w, "Failed to generate response", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(bytes)
		}))

		defer mockCompletionServer.Close()

		config.ApiGenerateURL = mockCompletionServer.URL
		companion.SetConfig(config)

		response, err := companion.SendCompletionRequest(mockCompletionRequest)
		if err != nil {
			t.Fatalf("SendCompletionRequest returned an error: %v", err)
		}

		if response.Content != "the generated prompt" {
			t.Errorf("Expected 'the generated prompt', got %s", response.Content)
		}
	})

	t.Run("Test SendChatRequest", func(t *testing.T) {
		mockChatResponse := models.Message{
			Role:    models.Assistant,
			Content: "the generated chat response",
		}

		mockChatRequest := models.Message{
			Role:    models.User,
			Content: "the generated chat request",
		}

		mockChatServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(mockChatResponse)
		}))
		defer mockChatServer.Close()

		config.ApiChatURL = mockChatServer.URL
		companion.SetConfig(config)

		response, err := companion.SendChatRequest(mockChatRequest)
		if err != nil {
			t.Fatalf("SendChatRequest returned an error: %v", err)
		}

		if response.Content != "the generated chat response" {
			t.Errorf("Expected 'the generated chat response', got %s", response.Content)
		}
	})

	t.Run("Test SendEmbeddingRequest", func(t *testing.T) {
		embeddingRequest := models.EmbeddingRequest{
			Model: "mock-model",
			Input: []string{"test input"},
		}
		mockEmbeddingResponse := openai.EmbeddingResponse{
			Model: "mock-model",
			Data: []openai.Embedding{
				{Embedding: []float64{1.0, 2.0, 3.0}},
			},
		}
		mockEmbeddingServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(mockEmbeddingResponse)
		}))
		defer mockEmbeddingServer.Close()

		config.ApiEmbedURL = mockEmbeddingServer.URL
		companion.SetConfig(config)

		response, err := companion.SendEmbeddingRequest(embeddingRequest)
		if err != nil {
			t.Fatalf("SendEmbeddingRequest returned an error: %v", err)
		}
		if len(response.Embeddings) != 1 {
			t.Errorf("Expected 1 embedding, got %d", len(response.Embeddings))
		}
	})
}
