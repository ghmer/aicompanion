package aicompanion_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/ghmer/aicompanion"
	"github.com/ghmer/aicompanion/models"
)

func TestAICompanion(t *testing.T) {
	// Prepare configuration for the AI Companion
	config := aicompanion.NewDefaultConfig(models.Ollama, "", "mock-model")

	// Create a new AI Companion instance
	companion := aicompanion.NewCompanion(*config)

	companion.SetSystemRole("you are a helpful assistant")

	t.Run("Test PrepareConversation", func(t *testing.T) {
		messages := companion.PrepareConversation()
		if len(messages) != 1 {
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
		newConfig := aicompanion.NewDefaultConfig(models.Ollama, "", "updated-model")
		newConfig.AiModel = "updated-model"
		companion.SetConfig(*newConfig)
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
}
