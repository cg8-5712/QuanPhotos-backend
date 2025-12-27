package security

import (
	"bytes"
	"errors"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"
)

var (
	ErrInvalidFileType     = errors.New("invalid file type")
	ErrFileTooLarge        = errors.New("file too large")
	ErrMagicNumberMismatch = errors.New("file content does not match extension")
	ErrPathTraversal       = errors.New("path traversal detected")
)

// FileSignature represents a file's magic number signature
type FileSignature struct {
	Offset int
	Magic  []byte
}

// Known file signatures (magic numbers)
var fileSignatures = map[string][]FileSignature{
	".jpg": {
		{Offset: 0, Magic: []byte{0xFF, 0xD8, 0xFF}},
	},
	".jpeg": {
		{Offset: 0, Magic: []byte{0xFF, 0xD8, 0xFF}},
	},
	".png": {
		{Offset: 0, Magic: []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}},
	},
	".gif": {
		{Offset: 0, Magic: []byte{0x47, 0x49, 0x46, 0x38}},
	},
	".webp": {
		{Offset: 0, Magic: []byte{0x52, 0x49, 0x46, 0x46}}, // RIFF
	},
	".bmp": {
		{Offset: 0, Magic: []byte{0x42, 0x4D}},
	},
	".tiff": {
		{Offset: 0, Magic: []byte{0x49, 0x49, 0x2A, 0x00}}, // Little endian
		{Offset: 0, Magic: []byte{0x4D, 0x4D, 0x00, 0x2A}}, // Big endian
	},
	".tif": {
		{Offset: 0, Magic: []byte{0x49, 0x49, 0x2A, 0x00}},
		{Offset: 0, Magic: []byte{0x4D, 0x4D, 0x00, 0x2A}},
	},
	// RAW formats
	".cr2": {
		{Offset: 0, Magic: []byte{0x49, 0x49, 0x2A, 0x00}}, // TIFF header
	},
	".cr3": {
		{Offset: 4, Magic: []byte{0x66, 0x74, 0x79, 0x70}}, // ftyp
	},
	".nef": {
		{Offset: 0, Magic: []byte{0x4D, 0x4D, 0x00, 0x2A}}, // TIFF header
	},
	".arw": {
		{Offset: 0, Magic: []byte{0x49, 0x49, 0x2A, 0x00}}, // TIFF header
	},
	".raf": {
		{Offset: 0, Magic: []byte{0x46, 0x55, 0x4A, 0x49}}, // FUJI
	},
	".orf": {
		{Offset: 0, Magic: []byte{0x49, 0x49, 0x52, 0x4F}}, // IIRO
	},
	".rw2": {
		{Offset: 0, Magic: []byte{0x49, 0x49, 0x55, 0x00}},
	},
	".dng": {
		{Offset: 0, Magic: []byte{0x49, 0x49, 0x2A, 0x00}},
		{Offset: 0, Magic: []byte{0x4D, 0x4D, 0x00, 0x2A}},
	},
}

// AllowedImageExtensions contains allowed image file extensions
var AllowedImageExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
	".webp": true,
	".bmp":  true,
	".tiff": true,
	".tif":  true,
	// RAW formats
	".cr2": true,
	".cr3": true,
	".nef": true,
	".arw": true,
	".raf": true,
	".orf": true,
	".rw2": true,
	".dng": true,
}

// FileValidator provides file validation functionality
type FileValidator struct {
	maxSize            int64
	allowedExtensions  map[string]bool
	checkMagicNumbers  bool
}

// NewFileValidator creates a new file validator
func NewFileValidator(maxSize int64, allowedExtensions map[string]bool, checkMagicNumbers bool) *FileValidator {
	return &FileValidator{
		maxSize:           maxSize,
		allowedExtensions: allowedExtensions,
		checkMagicNumbers: checkMagicNumbers,
	}
}

// NewImageValidator creates a validator for image files
func NewImageValidator(maxSize int64) *FileValidator {
	return NewFileValidator(maxSize, AllowedImageExtensions, true)
}

// ValidateFile validates a file
func (v *FileValidator) ValidateFile(file *multipart.FileHeader) error {
	// Check file size
	if file.Size > v.maxSize {
		return ErrFileTooLarge
	}

	// Get and validate extension
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !v.allowedExtensions[ext] {
		return ErrInvalidFileType
	}

	// Check for path traversal in filename
	if err := v.checkPathTraversal(file.Filename); err != nil {
		return err
	}

	// Check magic numbers
	if v.checkMagicNumbers {
		if err := v.checkMagicNumber(file, ext); err != nil {
			return err
		}
	}

	return nil
}

// checkPathTraversal checks for path traversal attempts
func (v *FileValidator) checkPathTraversal(filename string) error {
	// Clean the filename
	cleaned := filepath.Clean(filename)

	// Check for path traversal patterns
	if strings.Contains(cleaned, "..") ||
		strings.Contains(cleaned, "/") ||
		strings.Contains(cleaned, "\\") ||
		strings.HasPrefix(cleaned, ".") {
		return ErrPathTraversal
	}

	return nil
}

// checkMagicNumber verifies the file's magic number matches its extension
func (v *FileValidator) checkMagicNumber(file *multipart.FileHeader, ext string) error {
	signatures, exists := fileSignatures[ext]
	if !exists {
		// No signature defined for this extension, skip check
		return nil
	}

	// Open file
	f, err := file.Open()
	if err != nil {
		return err
	}
	defer f.Close()

	// Read first 16 bytes (enough for most signatures)
	header := make([]byte, 16)
	n, err := f.Read(header)
	if err != nil && err != io.EOF {
		return err
	}
	header = header[:n]

	// Check against known signatures
	for _, sig := range signatures {
		if sig.Offset+len(sig.Magic) <= len(header) {
			if bytes.Equal(header[sig.Offset:sig.Offset+len(sig.Magic)], sig.Magic) {
				return nil
			}
		}
	}

	return ErrMagicNumberMismatch
}

// ValidateMIMEType validates the MIME type from the Content-Type header
func ValidateMIMEType(contentType string, allowedTypes []string) bool {
	// Extract MIME type (remove charset and other parameters)
	mimeType := contentType
	if idx := strings.Index(contentType, ";"); idx != -1 {
		mimeType = strings.TrimSpace(contentType[:idx])
	}

	for _, allowed := range allowedTypes {
		if mimeType == allowed {
			return true
		}
	}
	return false
}

// AllowedImageMIMETypes contains allowed image MIME types
var AllowedImageMIMETypes = []string{
	"image/jpeg",
	"image/png",
	"image/gif",
	"image/webp",
	"image/bmp",
	"image/tiff",
	// RAW formats often have generic MIME types
	"image/x-canon-cr2",
	"image/x-canon-cr3",
	"image/x-nikon-nef",
	"image/x-sony-arw",
	"image/x-fuji-raf",
	"image/x-olympus-orf",
	"image/x-panasonic-rw2",
	"image/x-adobe-dng",
	"application/octet-stream", // Fallback for unknown types
}

// SanitizeFilename removes potentially dangerous characters from filename
func SanitizeFilename(filename string) string {
	// Get base name and extension
	ext := filepath.Ext(filename)
	name := strings.TrimSuffix(filepath.Base(filename), ext)

	// Remove or replace dangerous characters
	replacer := strings.NewReplacer(
		"..", "",
		"/", "",
		"\\", "",
		"<", "",
		">", "",
		":", "",
		"\"", "",
		"|", "",
		"?", "",
		"*", "",
		"\x00", "",
	)

	name = replacer.Replace(name)

	// Limit length
	if len(name) > 200 {
		name = name[:200]
	}

	// Ensure name is not empty
	if name == "" {
		name = "file"
	}

	return name + ext
}
