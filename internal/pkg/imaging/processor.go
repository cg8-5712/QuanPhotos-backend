package imaging

import (
	"fmt"
	"image"
	"image/jpeg"
	"os"
	"path/filepath"

	"github.com/disintegration/imaging"
)

// Processor handles image processing operations
type Processor struct {
	config ProcessorConfig
}

// NewProcessor creates a new image processor
func NewProcessor(config ProcessorConfig) *Processor {
	return &Processor{config: config}
}

// ProcessResult contains the results of image processing
type ProcessResult struct {
	// MainImagePath is the path to the processed main image
	MainImagePath string
	// ThumbnailPaths maps size name to thumbnail path
	ThumbnailPaths map[string]string
	// Width is the width of the processed main image
	Width int
	// Height is the height of the processed main image
	Height int
}

// Process processes an image file: auto-rotates, resizes if needed, and generates thumbnails
func (p *Processor) Process(srcPath, destDir, baseName string, orientation int) (*ProcessResult, error) {
	// Load the source image
	src, err := imaging.Open(srcPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open image: %w", err)
	}

	// Auto-rotate based on EXIF orientation
	src = p.autoRotate(src, orientation)

	// Resize if needed
	src = p.resizeIfNeeded(src)

	// Ensure destination directory exists
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// Save main image
	mainPath := filepath.Join(destDir, baseName+".jpg")
	if err := p.saveJPEG(src, mainPath, p.config.Quality); err != nil {
		return nil, fmt.Errorf("failed to save main image: %w", err)
	}

	// Generate thumbnails
	thumbnailPaths := make(map[string]string)
	for _, size := range p.config.ThumbnailSizes {
		thumbPath := filepath.Join(destDir, fmt.Sprintf("%s_%s.jpg", baseName, size.Name))
		if err := p.generateThumbnail(src, thumbPath, size); err != nil {
			return nil, fmt.Errorf("failed to generate %s thumbnail: %w", size.Name, err)
		}
		thumbnailPaths[size.Name] = thumbPath
	}

	bounds := src.Bounds()
	return &ProcessResult{
		MainImagePath:  mainPath,
		ThumbnailPaths: thumbnailPaths,
		Width:          bounds.Dx(),
		Height:         bounds.Dy(),
	}, nil
}

// ProcessToSeparateDirs processes an image and saves main image and thumbnails to separate directories
func (p *Processor) ProcessToSeparateDirs(srcPath, photoDir, thumbnailDir, baseName string, orientation int) (*ProcessResult, error) {
	// Load the source image
	src, err := imaging.Open(srcPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open image: %w", err)
	}

	// Auto-rotate based on EXIF orientation
	src = p.autoRotate(src, orientation)

	// Resize if needed
	src = p.resizeIfNeeded(src)

	// Ensure directories exist
	if err := os.MkdirAll(photoDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create photo directory: %w", err)
	}
	if err := os.MkdirAll(thumbnailDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create thumbnail directory: %w", err)
	}

	// Save main image
	mainPath := filepath.Join(photoDir, baseName+".jpg")
	if err := p.saveJPEG(src, mainPath, p.config.Quality); err != nil {
		return nil, fmt.Errorf("failed to save main image: %w", err)
	}

	// Generate thumbnails
	thumbnailPaths := make(map[string]string)
	for _, size := range p.config.ThumbnailSizes {
		thumbPath := filepath.Join(thumbnailDir, fmt.Sprintf("%s_%s.jpg", baseName, size.Name))
		if err := p.generateThumbnail(src, thumbPath, size); err != nil {
			return nil, fmt.Errorf("failed to generate %s thumbnail: %w", size.Name, err)
		}
		thumbnailPaths[size.Name] = thumbPath
	}

	bounds := src.Bounds()
	return &ProcessResult{
		MainImagePath:  mainPath,
		ThumbnailPaths: thumbnailPaths,
		Width:          bounds.Dx(),
		Height:         bounds.Dy(),
	}, nil
}

// autoRotate rotates the image based on EXIF orientation
func (p *Processor) autoRotate(img image.Image, orientation int) image.Image {
	switch orientation {
	case 2:
		return imaging.FlipH(img)
	case 3:
		return imaging.Rotate180(img)
	case 4:
		return imaging.FlipV(img)
	case 5:
		return imaging.Transpose(img)
	case 6:
		return imaging.Rotate270(img)
	case 7:
		return imaging.Transverse(img)
	case 8:
		return imaging.Rotate90(img)
	default:
		return img
	}
}

// resizeIfNeeded resizes the image if it exceeds the maximum dimension
func (p *Processor) resizeIfNeeded(img image.Image) image.Image {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	maxDim := p.config.MaxDimension
	if width <= maxDim && height <= maxDim {
		return img
	}

	// Calculate new dimensions while maintaining aspect ratio
	if width > height {
		return imaging.Resize(img, maxDim, 0, imaging.Lanczos)
	}
	return imaging.Resize(img, 0, maxDim, imaging.Lanczos)
}

// generateThumbnail creates a thumbnail of the specified size
func (p *Processor) generateThumbnail(img image.Image, destPath string, size ThumbnailSize) error {
	// Use Fill to maintain aspect ratio and crop from center
	thumb := imaging.Fill(img, size.Width, size.Height, imaging.Center, imaging.Lanczos)
	return p.saveJPEG(thumb, destPath, size.Quality)
}

// saveJPEG saves an image as JPEG with the specified quality
func (p *Processor) saveJPEG(img image.Image, path string, quality int) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	return jpeg.Encode(f, img, &jpeg.Options{Quality: quality})
}

// GetImageDimensions returns the dimensions of an image file
func GetImageDimensions(path string) (width, height int, err error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, 0, err
	}
	defer f.Close()

	config, _, err := image.DecodeConfig(f)
	if err != nil {
		return 0, 0, err
	}

	return config.Width, config.Height, nil
}
