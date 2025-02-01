# AI Companion CLI Interface Documentation

## Overview

This document provides an overview and detailed explanation of the `AICompanion` interface, which is defined in the Go file `aicompanion_interface.go`. The `AICompanion` interface is designed to abstract interactions with various AI models, providing a unified way to interact with different API providers such as Ollama and OpenAI.

## Table of Contents

1. **Interface Definition**
2. **Configuration and Initialization**
3. **Utility Functions**
4. **Interactions**
5. **Example Usage**

## 1. Interface Definition

The `AICompanion` interface defines a set of methods that must be implemented by any AI companion implementation:

- **PrepareConversation(message models.Message) []models.Message**: Prepares the conversation with a new message.
- **CreateMessage(role models.Role, input string) models.Message**: Creates a new message with the specified role and content.
- **CreateMessageWithImages(role models.Role, message string, images \*[]models.Base64Image) models.Message**: Creates a new message with images.
- **CreateUserMessage(input string, images \*[]models.Base64Image) models.Message**: Creates a user message with the specified content and optional images.
- **CreateAssistantMessage(input string) models.Message**: Creates an assistant message with the specified content.
- **AddMessage(message models.Message)**: Adds a new message to the conversation.
- **GetConfig() models.Configuration**: Retrieves the current configuration for the AI companion.
- **SetConfig(config models.Configuration)**: Sets a new configuration for the AI companion.
- **GetSystemRole() models.Message**: Retrieves the current system role message.
- **SetSystemRole(prompt string)**: Sets a new system role prompt.
- **GetEnrichmentPrompt() string**: Retrieves the current enrichment prompt.
- **SetEnrichmentPrompt(prompt string)**: Sets a new enrichment prompt.
- **GetFunctionsPrompt() string**: Retrieves the current functions prompt.
- **SetFunctionsPrompt(prompt string)**: Sets a new functions prompt.
- **GetSummarizationPrompt() string**: Retrieves the current summarization prompt.
- **SetSummarizationPrompt(prompt string)**: Sets a new summarization prompt.
- **GetConversation() []models.Message**: Retrieves the current conversation.
- **SetConversation(conversation []models.Message)**: Sets the current conversation.
- **GetHttpClient() \*http.Client**: Retrieves the current HTTP client used for requests.
- **SetHttpClient(client *http.Client)**: Sets a new HTTP client for requests.
- **GetModels() ([]models.Model, error)**: Retrieves all models supported by the endpoint.
- **SendChatRequest(message models.MessageRequest, streaming bool, callback func(m models.Message) error) (models.Message, error)**: Sends a chat request to an AI model and handles the response.
- **SendGenerateRequest(message models.MessageRequest, streaming bool, callback func(m models.Message) error) (models.Message, error)**: Sends a generate request to an AI model and handles the response.
- **SendEmbeddingRequest(embedding models.EmbeddingRequest) (models.EmbeddingResponse, error)**: Sends an embedding request to an AI model and retrieves the response.
- **SendModerationRequest(moderationRequest models.ModerationRequest) (models.ModerationResponse, error)**: Sends a moderation request to an AI model and retrieves the response.
- **HandleStreamResponse(resp \*http.Response, streamType models.StreamType, callback func(m models.Message) error) (models.Message, error)**: Handles streaming responses from chat requests.
- **SetVectorDB(vectorDb \*vectordb.VectorDb)**: Sets the vector database for vectordb (Retrieval Augmented Generation).
- **GetVectorDB() \*vectordb.VectorDb**: Retrieves the vector database.
- **RunFunction(function models.Function) (models.FunctionResponse, error)**: Runs a function provided by an AI model and returns the response.

## 2. Configuration and Initialization

The `NewCompanion` function initializes a new companion instance with the provided configuration:

```go
func NewCompanion(config models.Configuration, vectordb *vectordb.VectorDb) AICompanion {
    // Implementation details...
}
```

The `NewDefaultConfig` function creates a new default configuration for the API provider, token, chat model, and embedding model:

```go
func NewDefaultConfig(apiProvider models.ApiProvider, apiToken, chatModel, embeddingModel string) *models.Configuration {
    // Implementation details...
}
```

## 3. Utility Functions

The `ReadImageFromFile` function reads an image from the specified filepath and returns a Base64 encoded image:

```go
func ReadImageFromFile(filepath string) (models.Base64Image, error) {
    // Implementation details...
}
```

## 4. Interactions

The `AICompanion` interface supports various interactions with AI models, including chat, generate requests, embedding, and moderation:

- **Chat Requests**: The `SendChatRequest` method sends a chat request to an AI model and handles the response.
- **Generate Requests**: The `SendGenerateRequest` method sends a generate request to an AI model and handles the response.
- **Embedding Requests**: The `SendEmbeddingRequest` method sends an embedding request to an AI model and retrieves the response.
- **Moderation Requests**: The `SendModerationRequest` method sends a moderation request to an AI model and retrieves the response.

## 5. Example Usage

Below is an example of how to use the `AICompanion` interface:

```go
func main() {
    config := NewDefaultConfig(models.Ollama, "your_api_token", "chat_model", "embedding_model")
    companion := NewCompanion(*config, nil)
    
    // Example usage of methods...
}
```
