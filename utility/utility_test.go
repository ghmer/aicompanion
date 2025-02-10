package utility_test

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"testing"

	"github.com/ghmer/aicompanion/utility"
)

// createTestImage generates a simple test image with specified dimensions and color.
func createTestImage(width, height int, color color.RGBA) ([]byte, error) {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			img.Set(x, y, color)
		}
	}

	var buf bytes.Buffer
	err := jpeg.Encode(&buf, img, nil)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// createTestImage generates a simple test image with specified dimensions and color.
func TestReadFile(t *testing.T) {
	utility := utility.CompanionUtility{}
	t.Run("Test ReadFile", func(t *testing.T) {
		_, err := utility.ReadFile("../README.md")
		if err != nil {
			t.Error(err)
		}
	})
}

// TestResizeImage_ValidInput tests resizing with valid inputs.
func TestResizeImage_ValidInput(t *testing.T) {
	utility := utility.CompanionUtility{}
	originalWidth := 1920
	originalHeight := 1080
	maxSize := 512

	// Create a red test image
	testImage, err := createTestImage(originalWidth, originalHeight, color.RGBA{255, 0, 0, 255})
	if err != nil {
		t.Fatalf("failed to create test image: %v", err)
	}

	// Perform resize
	resizedImage, err := utility.ResizeImage(testImage, maxSize)
	if err != nil {
		t.Fatalf("failed to resize image: %v", err)
	}

	// Decode resized image
	resizedImg, _, err := image.Decode(bytes.NewReader(resizedImage))
	if err != nil {
		t.Fatalf("failed to decode resized image: %v", err)
	}

	// Validate dimensions
	resizedBounds := resizedImg.Bounds()
	newWidth := resizedBounds.Dx()
	newHeight := resizedBounds.Dy()

	if newWidth > maxSize || newHeight > maxSize {
		t.Errorf("resized dimensions exceed max size: got %dx%d, max %d", newWidth, newHeight, maxSize)
	}
}

// TestResizeImage_InvalidInputs tests invalid inputs to ResizeImage.
func TestResizeImage_InvalidInputs(t *testing.T) {
	utility := utility.CompanionUtility{}
	// Test with empty image bytes
	_, err := utility.ResizeImage([]byte{}, 512)
	if err == nil {
		t.Error("expected error for empty input, got nil")
	}

	// Test with invalid maxSize
	_, err = utility.ResizeImage([]byte{0xFF, 0xD8, 0xFF}, -100) // Valid JPEG header
	if err == nil {
		t.Error("expected error for invalid maxSize, got nil")
	}
}

// TestResizeImage_AspectRatio tests that aspect ratio is preserved during resizing.
func TestResizeImage_AspectRatio(t *testing.T) {
	utility := utility.CompanionUtility{}
	originalWidth := 1280
	originalHeight := 720
	maxSize := 256

	// Create a green test image
	testImage, err := createTestImage(originalWidth, originalHeight, color.RGBA{0, 255, 0, 255})
	if err != nil {
		t.Fatalf("failed to create test image: %v", err)
	}

	// Perform resize
	resizedImage, err := utility.ResizeImage(testImage, maxSize)
	if err != nil {
		t.Fatalf("failed to resize image: %v", err)
	}

	// Decode resized image
	resizedImg, _, err := image.Decode(bytes.NewReader(resizedImage))
	if err != nil {
		t.Fatalf("failed to decode resized image: %v", err)
	}

	// Check aspect ratio
	resizedBounds := resizedImg.Bounds()
	newWidth := resizedBounds.Dx()
	newHeight := resizedBounds.Dy()
	expectedAspectRatio := float64(originalWidth) / float64(originalHeight)
	actualAspectRatio := float64(newWidth) / float64(newHeight)

	if abs(expectedAspectRatio-actualAspectRatio) > 0.01 {
		t.Errorf("aspect ratio mismatch: expected %.2f, got %.2f", expectedAspectRatio, actualAspectRatio)
	}
}

// abs computes the absolute value of a float64.
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// TestResizeImage_PredefinedResolutions tests resizing using predefined resolutions.
func TestResizeImage_PredefinedResolutions(t *testing.T) {
	util := utility.CompanionUtility{}
	originalWidth := 3840
	originalHeight := 2160

	// Create a blue test image
	testImage, err := createTestImage(originalWidth, originalHeight, color.RGBA{0, 0, 255, 255})
	if err != nil {
		t.Fatalf("failed to create test image: %v", err)
	}

	for name, resolution := range []utility.Resolution{utility.Res4K, utility.Res2K, utility.Res1080p, utility.Res720p, utility.Res480p, utility.Res360p, utility.Res320p, utility.Res240p, utility.Res144p, utility.Pixel1024, utility.Pixel512} {
		resizedImage, err := util.ResizeImage(testImage, int(resolution))
		if err != nil {
			t.Errorf("failed to resize image to %d: %v", name, err)
			continue
		}

		// Decode resized image
		resizedImg, _, err := image.Decode(bytes.NewReader(resizedImage))
		if err != nil {
			t.Errorf("failed to decode resized %d image: %v", name, err)
			continue
		}

		// Validate dimensions
		resizedBounds := resizedImg.Bounds()
		newWidth := resizedBounds.Dx()
		newHeight := resizedBounds.Dy()

		if newWidth > int(resolution) || newHeight > int(resolution) {
			t.Errorf("%d resized dimensions exceed max size: got %dx%d, max %d", name, newWidth, newHeight, name)
		}
	}
}
