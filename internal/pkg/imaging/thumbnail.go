package imaging

// ThumbnailSize represents a thumbnail size configuration
type ThumbnailSize struct {
	Name    string // sm, md, lg
	Width   int
	Height  int
	Quality int
}

// DefaultThumbnailSizes defines the default thumbnail sizes
var DefaultThumbnailSizes = []ThumbnailSize{
	{Name: "sm", Width: 300, Height: 200, Quality: 80},
	{Name: "md", Width: 800, Height: 533, Quality: 85},
	{Name: "lg", Width: 1600, Height: 1067, Quality: 90},
}

// ProcessorConfig holds configuration for image processing
type ProcessorConfig struct {
	// MaxDimension is the maximum width/height for the main image
	MaxDimension int
	// Quality is the JPEG quality for the main image (1-100)
	Quality int
	// ThumbnailSizes defines the thumbnail sizes to generate
	ThumbnailSizes []ThumbnailSize
}

// DefaultProcessorConfig returns the default processor configuration
func DefaultProcessorConfig() ProcessorConfig {
	return ProcessorConfig{
		MaxDimension:   4096,
		Quality:        92,
		ThumbnailSizes: DefaultThumbnailSizes,
	}
}
