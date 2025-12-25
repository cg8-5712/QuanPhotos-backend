package storage

import (
	"context"
	"io"
	"time"
)

// Storage defines the interface for file storage operations
type Storage interface {
	// Upload saves a file to the specified path
	Upload(ctx context.Context, file io.Reader, path string) error

	// Delete removes a file at the specified path
	Delete(ctx context.Context, path string) error

	// Move moves a file from one path to another
	Move(ctx context.Context, from, to string) error

	// GetURL returns the URL for accessing a file
	GetURL(path string) string

	// Exists checks if a file exists at the specified path
	Exists(ctx context.Context, path string) bool

	// EnsureDir ensures the directory exists
	EnsureDir(ctx context.Context, path string) error
}

// PathGenerator generates storage paths
type PathGenerator struct {
	basePath string
}

// NewPathGenerator creates a new path generator
func NewPathGenerator(basePath string) *PathGenerator {
	return &PathGenerator{basePath: basePath}
}

// TempPath generates a temporary file path
func (g *PathGenerator) TempPath(filename string) string {
	return g.basePath + "/temp/" + filename
}

// PhotoPath generates a photo storage path based on date
func (g *PathGenerator) PhotoPath(t time.Time, filename string) string {
	return g.basePath + "/photos/" + t.Format("2006/01/02") + "/" + filename
}

// ThumbnailPath generates a thumbnail storage path based on date
// Note: filename should NOT include the size suffix (_sm, _md, _lg)
func (g *PathGenerator) ThumbnailPath(t time.Time, filename string) string {
	return g.basePath + "/thumbnails/" + t.Format("2006/01/02") + "/" + filename
}

// RawPath generates a RAW file storage path based on date
func (g *PathGenerator) RawPath(t time.Time, filename string) string {
	return g.basePath + "/raw/" + t.Format("2006/01/02") + "/" + filename
}

// RelativePhotoPath generates a relative photo path (for database storage)
func (g *PathGenerator) RelativePhotoPath(t time.Time, filename string) string {
	return "/photos/" + t.Format("2006/01/02") + "/" + filename
}

// RelativeThumbnailPath generates a relative thumbnail path (for database storage)
// Note: filename should NOT include the size suffix (_sm, _md, _lg)
func (g *PathGenerator) RelativeThumbnailPath(t time.Time, filename string) string {
	return "/thumbnails/" + t.Format("2006/01/02") + "/" + filename
}

// RelativeRawPath generates a relative RAW file path (for database storage)
func (g *PathGenerator) RelativeRawPath(t time.Time, filename string) string {
	return "/raw/" + t.Format("2006/01/02") + "/" + filename
}
