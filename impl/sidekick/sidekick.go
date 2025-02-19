package sidekick

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"net/http"
	"os"

	_ "image/gif" // Support for GIF decoding

	"github.com/ghmer/aicompanion/models"
	"github.com/ghmer/aicompanion/terminal"
	"golang.org/x/image/draw"
)

// Resolution represents different image resolutions.
type Resolution int

const (
	Res4K     Resolution = 3840
	Res2K     Resolution = 2048
	Res1080p  Resolution = 1920
	Res720p   Resolution = 1280
	Res480p   Resolution = 640
	Res360p   Resolution = 480
	Res320p   Resolution = 320
	Res240p   Resolution = 320
	Res144p   Resolution = 256
	Pixel1024 Resolution = 1024
	Pixel512  Resolution = 512
)

type SideKick struct {
}

// ResizeImage resizes an image to the specified maximum dimension while maintaining its aspect ratio.
// Larger dimension (width or height) will be resized to maxSize.
// Input: imageBytes []byte (image data), maxSize int (max dimension).
// Output: Resized image as []byte, error if any issue occurs.
func (utility *SideKick) ResizeImage(imageBytes []byte, maxSize int) ([]byte, error) {
	// Validate input
	if len(imageBytes) == 0 {
		return nil, errors.New("input image data is empty")
	}
	if maxSize <= 0 {
		return nil, errors.New("maxSize must be greater than zero")
	}

	// Decode image
	img, format, err := utility.DecodeImage(imageBytes)
	if err != nil {
		return nil, err
	}

	// Calculate new dimensions
	newWidth, newHeight := utility.CalculateNewDimensions(img.Bounds(), maxSize)

	// Resize the image
	resizedImg := utility.Resize(img, newWidth, newHeight)

	// Encode resized image back to original format
	resizedBytes, err := utility.EncodeImage(resizedImg, format)
	if err != nil {
		return nil, err
	}

	return resizedBytes, nil
}

// DecodeImage decodes image bytes into an image.Image and detects the format.
func (utility *SideKick) DecodeImage(imageBytes []byte) (image.Image, string, error) {
	img, format, err := image.Decode(bytes.NewReader(imageBytes))
	if err != nil {
		return nil, "", errors.New("failed to decode image: " + err.Error())
	}
	return img, format, nil
}

// CalculateNewDimensions calculates the new width and height while maintaining the aspect ratio.
func (utility *SideKick) CalculateNewDimensions(bounds image.Rectangle, maxSize int) (int, int) {
	width := bounds.Dx()
	height := bounds.Dy()

	var scale float64
	if width > height {
		scale = float64(maxSize) / float64(width)
	} else {
		scale = float64(maxSize) / float64(height)
	}

	newWidth := int(float64(width) * scale)
	newHeight := int(float64(height) * scale)

	return newWidth, newHeight
}

// Resize resizes an image to the specified width and height using high-quality scaling.
func (utility *SideKick) Resize(img image.Image, newWidth, newHeight int) image.Image {
	dst := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
	draw.CatmullRom.Scale(dst, dst.Rect, img, img.Bounds(), draw.Over, nil)
	return dst
}

// EncodeImage encodes an image into a specific format (JPEG, PNG, etc.).
func (utility *SideKick) EncodeImage(img image.Image, format string) ([]byte, error) {
	var buf bytes.Buffer
	var err error

	switch format {
	case "jpeg":
		err = jpeg.Encode(&buf, img, nil)
	case "png":
		err = png.Encode(&buf, img)
	default:
		err = errors.New("unsupported image format: " + format)
	}

	if err != nil {
		return nil, errors.New("failed to encode image: " + err.Error())
	}

	return buf.Bytes(), nil
}

// ReadFile reads a file and returns its base64 encoded content.
func (utility *SideKick) ReadFile(filepath string) ([]byte, error) {
	return os.ReadFile(filepath)
}

// createMessage creates a new message with the given role and content.
func (utility *SideKick) CreateMessage(role models.Role, input string) models.Message {
	var message models.Message = models.Message{
		Role:    role,
		Content: input,
		Images:  nil,
	}

	return message
}

// CreateMessageWithImages creates a new message with the given role, content and images
func (utility *SideKick) CreateMessageWithImages(role models.Role, input string, images *[]models.Base64Image) models.Message {
	var message models.Message = models.Message{
		Role:    role,
		Content: input,
		Images:  images,
	}

	return message
}

// CreateUserMessage creates a new user message with the given input string
func (utility *SideKick) CreateUserMessage(input string, images *[]models.Base64Image) models.Message {
	if images != nil && len(*images) > 0 {
		return utility.CreateMessageWithImages(models.User, input, images)
	}
	return utility.CreateMessage(models.User, input)
}

// CreateAssistantMessage creates a new assistant message with the given input string
func (utility *SideKick) CreateAssistantMessage(input string) models.Message {
	return utility.CreateMessage(models.Assistant, input)
}

func (utility *SideKick) CreateEmbeddingRequest(model models.Model, input []string) models.EmbeddingRequest {
	return models.EmbeddingRequest{
		Model: model.Model,
		Input: input,
	}
}

func (utility *SideKick) CreateModerationRequest(input string) models.ModerationRequest {
	return models.ModerationRequest{
		Input: input,
	}
}

func (utility *SideKick) RunFunction(httpClient *http.Client, tool models.Tool, payload models.FunctionPayload, debug, trace bool) (models.FunctionResponse, error) {
	result := models.FunctionResponse{}

	payloadBytes, err := json.Marshal(payload.Parameters)
	if err != nil {
		log.Println(err)
		return result, err
	}

	// Create and configure the HTTP request
	req, err := http.NewRequestWithContext(context.Background(), "POST", tool.Endpoint, bytes.NewBuffer(payloadBytes))
	if err != nil {
		log.Println(err)
		return result, err
	}
	req.Header.Set("Authorization", "Bearer "+tool.ApiKey)
	req.Header.Set("Content-Type", "application/json")

	// Execute the HTTP request
	resp, err := httpClient.Do(req)
	if err != nil {
		log.Println(err)
		return result, err
	}
	defer resp.Body.Close()
	if debug {
		log.Printf("RunFunction: StatusCode %d, Status %s\n", resp.StatusCode, resp.Status)
	}
	if trace {
		log.Printf("RunFunction: payload %s\n", string(payloadBytes))
	}

	responseBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return result, err
	}
	if trace {
		log.Printf("RunFunction: responseBytes %s\n", string(responseBytes))
	}

	err = json.Unmarshal(responseBytes, &result)
	if err != nil {
		log.Println(err)
		return result, err
	}
	return result, nil
}

// ClearLine clears the current line if output is enabled in the configuration
func (utility *SideKick) ClearLine(termconfig models.Terminal) {
	if termconfig.Output {
		fmt.Print(terminal.ClearLine)
	}
}

// Print prints the given content to the console with color and reset.
func (utility *SideKick) Print(content string, termconfig models.Terminal) {
	if termconfig.Output {
		fmt.Printf("%s%s%s", termconfig.Color, content, terminal.Reset)
	}
}

// Println prints the given content to the console with color and a newline character, then resets the color.
func (utility *SideKick) Println(content string, termconfig models.Terminal) {
	if termconfig.Output {
		fmt.Printf("%s%s%s\n", termconfig.Color, content, terminal.Reset)
	}
}

// PrintError prints an error message to the console in red.
func (utility *SideKick) Error(err error) {
	fmt.Printf("%s%v%s\n", terminal.Red, err, terminal.Reset)
}

func (utility *SideKick) Debug(payload string, termconfig models.Terminal) {
	if termconfig.Debug {
		fmt.Println(payload)
	}
}

func (utility *SideKick) Trace(payload string, termconfig models.Terminal) {
	if termconfig.Trace {
		fmt.Println(payload)
	}
}

// PrepareArray prepares an array of messages based on the includeStrategy.
func (utility *SideKick) PrepareArray(messages []models.Message, includeStrategy models.IncludeStrategy, maxMessages int) []models.Message {
	var newarray []models.Message
	for _, msg := range messages {
		switch includeStrategy {
		case models.IncludeAssistant:
			{
				if msg.Role == models.Assistant {
					newarray = append(newarray, msg)
				}
			}
		case models.IncludeUser:
			{
				if msg.Role == models.User {
					newarray = append(newarray, msg)
				}
			}
		case models.IncludeBoth:
			{
				newarray = append(newarray, msg)
			}
		default:
			{
				newarray = append(newarray, msg)
			}
		}
	}

	if len(newarray) > maxMessages {
		newarray = newarray[len(newarray)-maxMessages:]
	}

	return newarray
}

func (utility *SideKick) VerifyStatus(resp *http.Response) error {
	if resp.StatusCode >= 400 {
		return fmt.Errorf("unexpected status code: %d, status: %s", resp.StatusCode, resp.Status)
	}

	return nil
}
