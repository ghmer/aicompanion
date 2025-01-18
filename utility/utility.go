package utility

import (
	"bytes"
	"errors"
	"image"
	"image/jpeg"
	"image/png"

	_ "image/gif" // Support for GIF decoding

	"golang.org/x/image/draw"
)

// Predefined resolutions map
var Resolutions = map[string]int{
	"4K": 3840,
	"2K": 2048,
}

// ResizeImage resizes an image to the specified maximum dimension while maintaining its aspect ratio.
// Larger dimension (width or height) will be resized to maxSize.
// Input: imageBytes []byte (image data), maxSize int (max dimension).
// Output: Resized image as []byte, error if any issue occurs.
func ResizeImage(imageBytes []byte, maxSize int) ([]byte, error) {
	// Validate input
	if len(imageBytes) == 0 {
		return nil, errors.New("input image data is empty")
	}
	if maxSize <= 0 {
		return nil, errors.New("maxSize must be greater than zero")
	}

	// Decode image
	img, format, err := decodeImage(imageBytes)
	if err != nil {
		return nil, err
	}

	// Calculate new dimensions
	newWidth, newHeight := calculateNewDimensions(img.Bounds(), maxSize)

	// Resize the image
	resizedImg := resize(img, newWidth, newHeight)

	// Encode resized image back to original format
	resizedBytes, err := encodeImage(resizedImg, format)
	if err != nil {
		return nil, err
	}

	return resizedBytes, nil
}

// decodeImage decodes image bytes into an image.Image and detects the format.
func decodeImage(imageBytes []byte) (image.Image, string, error) {
	img, format, err := image.Decode(bytes.NewReader(imageBytes))
	if err != nil {
		return nil, "", errors.New("failed to decode image: " + err.Error())
	}
	return img, format, nil
}

// calculateNewDimensions calculates the new width and height while maintaining the aspect ratio.
func calculateNewDimensions(bounds image.Rectangle, maxSize int) (int, int) {
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

// resize resizes an image to the specified width and height using high-quality scaling.
func resize(img image.Image, newWidth, newHeight int) image.Image {
	dst := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
	draw.CatmullRom.Scale(dst, dst.Rect, img, img.Bounds(), draw.Over, nil)
	return dst
}

// encodeImage encodes an image into a specific format (JPEG, PNG, etc.).
func encodeImage(img image.Image, format string) ([]byte, error) {
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

/*
// Example usage: resize to predefined resolution or custom maxSize
func main() {
	// Example image data (replace with actual input)
	inputImage := []byte{} // Load your image file data here
	maxSize := 512         // Example: Resize to max dimension of 512 pixels

	// Attempt to resize
	resizedImage, err := ResizeImage(inputImage, maxSize)
	if err != nil {
		panic("error resizing image: " + err.Error())
	}

	// Use resizedImage as needed (e.g., save to a file, return from an API, etc.)
	println("Image resized successfully")
}
*/
