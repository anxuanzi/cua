package screen

import (
	"image"

	"golang.org/x/image/draw"
)

// Default maximum dimensions for screenshots sent to LLMs.
// These values are optimized for Claude and GPT-4V vision models.
const (
	DefaultMaxWidth  = 1280
	DefaultMaxHeight = 800
)

// Resize scales an image to fit within maxWidth x maxHeight while preserving aspect ratio.
// Uses high-quality CatmullRom interpolation for best results.
// Returns the resized image and the new dimensions.
func Resize(img image.Image, maxWidth, maxHeight int) (image.Image, int, int) {
	bounds := img.Bounds()
	origW := bounds.Dx()
	origH := bounds.Dy()

	// Calculate scaled dimensions
	newW, newH := CalculateScaledDimensions(origW, origH, maxWidth, maxHeight)

	// If no resize needed, return original
	if newW == origW && newH == origH {
		return img, origW, origH
	}

	// Create destination image
	resized := image.NewRGBA(image.Rect(0, 0, newW, newH))

	// Use high-quality CatmullRom interpolation
	draw.CatmullRom.Scale(resized, resized.Bounds(), img, bounds, draw.Over, nil)

	return resized, newW, newH
}

// ResizeWithQuality provides control over interpolation quality.
type Quality int

const (
	QualityNearest    Quality = iota // Fastest, lowest quality
	QualityBilinear                  // Good balance of speed and quality
	QualityCatmullRom                // High quality, slower
)

// ResizeWithQuality scales an image with the specified interpolation quality.
func ResizeWithQuality(img image.Image, maxWidth, maxHeight int, quality Quality) (image.Image, int, int) {
	bounds := img.Bounds()
	origW := bounds.Dx()
	origH := bounds.Dy()

	newW, newH := CalculateScaledDimensions(origW, origH, maxWidth, maxHeight)

	if newW == origW && newH == origH {
		return img, origW, origH
	}

	resized := image.NewRGBA(image.Rect(0, 0, newW, newH))

	var interpolator draw.Interpolator
	switch quality {
	case QualityNearest:
		interpolator = draw.NearestNeighbor
	case QualityBilinear:
		interpolator = draw.BiLinear
	case QualityCatmullRom:
		interpolator = draw.CatmullRom
	default:
		interpolator = draw.CatmullRom
	}

	interpolator.Scale(resized, resized.Bounds(), img, bounds, draw.Over, nil)
	return resized, newW, newH
}

// CalculateScaledDimensions computes new dimensions that fit within maxWidth x maxHeight
// while preserving the original aspect ratio.
func CalculateScaledDimensions(origW, origH, maxWidth, maxHeight int) (newW, newH int) {
	// If image already fits, return original dimensions
	if origW <= maxWidth && origH <= maxHeight {
		return origW, origH
	}

	// Calculate aspect ratios
	aspectRatio := float64(origW) / float64(origH)
	targetAspect := float64(maxWidth) / float64(maxHeight)

	if aspectRatio > targetAspect {
		// Width is the limiting factor
		newW = maxWidth
		newH = int(float64(maxWidth) / aspectRatio)
	} else {
		// Height is the limiting factor
		newH = maxHeight
		newW = int(float64(maxHeight) * aspectRatio)
	}

	// Ensure minimum dimensions
	if newW < 1 {
		newW = 1
	}
	if newH < 1 {
		newH = 1
	}

	return newW, newH
}

// ResizeToExact resizes an image to exact dimensions, potentially distorting aspect ratio.
// Use this only when exact dimensions are required.
func ResizeToExact(img image.Image, width, height int) image.Image {
	resized := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.CatmullRom.Scale(resized, resized.Bounds(), img, img.Bounds(), draw.Over, nil)
	return resized
}

// Thumbnail creates a small thumbnail image, useful for memory management.
func Thumbnail(img image.Image, maxSize int) image.Image {
	resized, _, _ := Resize(img, maxSize, maxSize)
	return resized
}
