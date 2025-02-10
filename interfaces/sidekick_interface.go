package sidekick

import (
	"image"
	"net/http"

	"github.com/ghmer/aicompanion/impl/sidekick"
	"github.com/ghmer/aicompanion/models"
)

type SideKickInterface interface {
	// CreateMessage creates a new message with the given role and input string
	CreateMessage(role models.Role, input string) models.Message

	// CreateMessageWithImages creates a new message with the given role, input string, and images
	CreateMessageWithImages(role models.Role, message string, images *[]models.Base64Image) models.Message

	//CreateUserMessage creates a new user message with the given input string
	CreateUserMessage(input string, images *[]models.Base64Image) models.Message

	// CreateAssistantMessage creates a new assistant message with the given input string
	CreateAssistantMessage(input string) models.Message

	// CreateEmbeddingRequest creates an embedding request for the given input.
	CreateEmbeddingRequest(model models.Model, input []string) models.EmbeddingRequest

	// CreateModerationRequest
	CreateModerationRequest(input string) models.ModerationRequest

	// ResizeImage resizes an image to the specified maximum dimension while maintaining its aspect ratio.
	ResizeImage(imageBytes []byte, maxSize int) ([]byte, error)

	// DecodeImage decodes image bytes into an image.Image and detects the format.
	DecodeImage(imageBytes []byte) (image.Image, string, error)

	// Resize resizes an image to the specified width and height using high-quality scaling.
	Resize(img image.Image, newWidth, newHeight int) image.Image

	// CalculateNewDimensions calculates the new width and height while maintaining the aspect ratio.
	CalculateNewDimensions(bounds image.Rectangle, maxSize int) (int, int)

	// EncodeImage encodes an image into a specific format (JPEG, PNG, etc.).
	EncodeImage(img image.Image, format string) ([]byte, error)

	// ReadFile reads a file and returns its base64 encoded content.
	ReadFile(filepath string) ([]byte, error)

	// RunFunction runs a function and returns the response
	RunFunction(httpClient *http.Client, function models.Function, payload models.FunctionPayload, debug, trace bool) (models.FunctionResponse, error)

	// Debug logs a debug message.
	Debug(payload string, termconfig models.Terminal)

	// Trace logs a trace message.
	Trace(payload string, termconfig models.Terminal)

	// Error logs an error message.
	Error(err error)

	// ClearLine clears the current line of the terminal
	ClearLine(termconfig models.Terminal)

	// Print prints the content to the terminal.
	Print(content string, termconfig models.Terminal)

	// Println prints the content to the terminal.
	Println(content string, termconfig models.Terminal)

	// PrepareArray filters and limits messages based on the includeStrategy.
	PrepareArray(messages []models.Message, includeStrategy models.IncludeStrategy, maxMessages int) []models.Message
}

func NewSideKick() SideKickInterface {
	return &sidekick.SideKick{}
}
