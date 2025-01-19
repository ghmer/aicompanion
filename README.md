# ai-companion

poc to communicate with LLMs via API

## Configuration

The configuration file is a JSON file that contains the following properties:

- `ai_type`: The type of AI model to use (e.g. chat, embed, moderation).
- `ai_model`: The name of the AI model to use.
- `api_provider`: The name of the AI model provider (e.g. OpenAI, Ollama).
- `api_key`: The API key to use for authentication with the AI model provider.
- `api_chat_url`: The URL for the API endpoint to use for chat requests.
- `api_generate_url`: The URL for the API endpoint to use for generate requests.
- `api_embed_url`: The URL for the API endpoint to use for embed requests.
- `api_moderation_url`: The URL for the API endpoint to use for moderation requests.
- `term_output`: whether to output chat responses to terminal.
- `term_color`: The color scheme to use in terminal output.
- `max_messages`: The maximum number of messages allowed in a chat conversation.
- `max_input_length`: The maximum length of allowed input.
- `http_client_timeout`: The timeout value for HTTP client requests, in seconds.
- `buffer_size`: The buffer size for reading and writing data, in bytes.

### Configuration Options

#### ai_type

|Allowed value|Description
|---|---
|chat|The AI model is used for chat.
|embed|The AI model is used for generating embeddings.
|moderation|The AI model is used for moderation tasks.

#### api_provider

|Allowed value|Description
|---|---
|Ollama|The API provider is Ollama.
|OpenAI|The API provider is OpenAI.

#### term_color

|Allowed value|Description
|---|---
|green|Green color scheme.
|red|Red color scheme.
|blue|Blue color scheme.
|yellow|Yellow color scheme.
|cyan|Cyan color scheme.
|magenta|Magenta color scheme.
|white|White color scheme.
|black|Black color scheme.

### Example Configuration File

```json
{
  "ai_type": "chat|embed|moderation",
  "ai_model": "llama3.2:3b",
  "api_provider": "Ollama|OpenAI",
  "api_key": "my-api-key",
  "api_chat_url": "https://<your.ollama.endpoint>/api/chat",
  "api_generate_url": "https://<your.ollama.endpoint>/api/generate",
  "api_embed_url": "https://<your.ollama.endpoint>/api/embed",
  "api_moderation_url": "https://<your.ollama.endpoint>/api/moderation",
  "term_color": "green",
  "term_output": true,
  "max_messages": 20,
  "max_input_length": 500,
  "http_client_timeout": 300,
  "buffer_size": 1024
}
```

## Examples

### basic usage

```golang
func main() {
    companion := aicompanion.NewCompanion(models.Configuration{
        ApiProvider:       "Ollama",
        ApiKey:            "",
        AiModel:           "llama3.1:8b",
        AIType:            "chat",
        ApiChatURL:        "http://localhost:11434/api/chat",
        ApiGenerateURL:    "http://localhost:11434/api/generate",
        ApiEmbedURL:       "http://localhost:11434/api/embed",
        ApiModerationURL:  "http://localhost:11434/api/moderation",
        MaxInputLength:    500,
        HTTPClientTimeout: 300,
        BufferSize:        1024,
        MaxMessages:       20,
        Color:             terminal.Green,
        Output:            true,
    })

    companion.SetSystemRole("you are a helpful assistant that only replies in haikus")
    companion.SendChatRequest(companion.CreateMessage(models.User, "hi! Who are you?"))
}
```

```console
% go run main.go 
> Assistant at hand
Helping you with care and ease
Friendly, here to aid
```

### image recognition

```golang
func main() {
    companion := aicompanion.NewCompanion(models.Configuration{
        AiModel:           "llama3.2-vision:latest",
        ApiChatURL:        "http://localhost:11434/api/chat",
        ApiGenerateURL:    "http://localhost:11434/api/generate",
        ApiEmbedURL:       "http://localhost:11434/api/embed",
        MaxInputLength:    500,
        HTTPClientTimeout: 300,
        BufferSize:        2048,
        ApiProvider:       "Ollama",
        ApiKey:            "",
        MaxMessages:       20,
        Color:             terminal.Green,
        Output:            true,
    })

    companion.SetSystemRole("you are a helpful assistant")

    image, err := aicompanion.ReadImageFromFile("/path/to/image.jpg")
    if err != nil {
        log.Fatal(err)
    }
    companion.SendGenerateRequest(companion.CreateMessageWithImages(models.User, "describe in detail what you can see in this image.", []models.Base64Image{image}))
}
```

### ai vs ai

```golang
func main() {
    companion1 := NewCompanion(*Config)
    companion2 := NewCompanion(*Config)
    companion1.SetSystemRole("you are Batman") 
    companion2.SetSystemRole("you are the Joker") 

    message := companion1.createMessage(User, "*the scene opens*")

    var assistant1 bool = true

    for {
        if assistant1 {
            message, _ = companion1.ProcessUserInput(message)
        } else {
            message, _ = companion2.ProcessUserInput(message)
        }
        message.Role = User
        assistant1 = !assistant1
    }
}
```
