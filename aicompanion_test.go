package aicompanion_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/ghmer/aicompanion"
	"github.com/ghmer/aicompanion/models"
)

func TestAICompanion(t *testing.T) {
	apiKey, vectorApiKey := "", ""
	config := aicompanion.NewDefaultConfig(models.Ollama, apiKey, "llama3.1:8b", "mxai-embed-large", models.SqlVectorDb, "vectorstore.db", vectorApiKey)
	config.Output = false
	companion := aicompanion.NewCompanion(*config)
	companion.SetSystemRole("you are a helpful assistant")

	t.Run("Test PrepareConversation", func(t *testing.T) {
		messages := companion.PrepareConversation()
		if len(messages) != 1 {
			t.Errorf("Expected 1 messages, got %d messages", len(messages))
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

		t.Logf("message: %v", message)
	})

	t.Run("Test SendMessage", func(t *testing.T) {
		msg := companion.CreateMessage(models.User, "Hello!")
		_, err := companion.SendChatRequest(msg, false, nil)
		if err != nil {
			t.Errorf("Failed to get AI response: %v", err)
		}
	})

	t.Run("Test PrepareConversation", func(t *testing.T) {
		messages := companion.PrepareConversation()
		if len(messages) != 3 {
			t.Errorf("Expected 3 messages, got %d messages", len(messages))
		}
	})

	t.Run("Test CreateMessageWithImages", func(t *testing.T) {
		images := []models.Base64Image{{Data: "mockImageData"}}
		message := companion.CreateMessageWithImages(models.User, "Hello with images!", &images)
		if len(*message.Images) != 1 {
			t.Errorf("Expected 1 image, got %d", len(*message.Images))
		}
	})

	t.Run("Test AddMessage", func(t *testing.T) {
		message := models.Message{Role: models.User, Content: "New message"}
		companion.AddMessage(message)
		conversation := companion.GetConversation()
		if len(conversation) != 3 {
			t.Errorf("Expected 3 message in conversation, got %d", len(conversation))
		}
	})

	t.Run("Test GetConfig", func(t *testing.T) {
		if companion.GetConfig().AiModels != config.AiModels {
			t.Errorf("Expected AI model '%s', got '%s'", config.AiModels, companion.GetConfig().AiModels)
		}
	})

	t.Run("Test SetConfig", func(t *testing.T) {
		newConfig := aicompanion.NewDefaultConfig(models.Ollama, "", "updated-model", "", models.WeaviateDb, "", "")
		newConfig.AiModels.ChatModel.Model = "updated-model"
		companion.SetConfig(*newConfig)
		if companion.GetConfig().AiModels.ChatModel.Model != "updated-model" {
			t.Errorf("Expected updated model 'updated-model', got '%s'", companion.GetConfig().AiModels.ChatModel)
		}
	})

	t.Run("Test GetModels", func(t *testing.T) {
		models, err := companion.GetModels()
		if err != nil {
			t.Errorf("Failed to get models: %v", err)
		}
		if len(models) == 0 {
			t.Errorf("Expected at least one model, got none")
		}
	})
	t.Run("Test GetSystemRole", func(t *testing.T) {
		role := companion.GetSystemRole()
		if role.Content != "you are a helpful assistant" {
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
}
