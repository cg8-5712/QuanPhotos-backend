package storage

import "errors"

var (
	// ErrFileNotFound indicates the file was not found
	ErrFileNotFound = errors.New("file not found")

	// ErrFileTooLarge indicates the file exceeds the maximum size
	ErrFileTooLarge = errors.New("file too large")

	// ErrInvalidFileType indicates the file type is not allowed
	ErrInvalidFileType = errors.New("invalid file type")

	// ErrCreateDirectory indicates failed to create directory
	ErrCreateDirectory = errors.New("failed to create directory")

	// ErrWriteFile indicates failed to write file
	ErrWriteFile = errors.New("failed to write file")

	// ErrReadFile indicates failed to read file
	ErrReadFile = errors.New("failed to read file")

	// ErrDeleteFile indicates failed to delete file
	ErrDeleteFile = errors.New("failed to delete file")

	// ErrMoveFile indicates failed to move file
	ErrMoveFile = errors.New("failed to move file")
)

// FileType represents supported file types
type FileType string

const (
	FileTypeJPEG FileType = "jpeg"
	FileTypePNG  FileType = "png"
	FileTypeCR2  FileType = "cr2"
	FileTypeCR3  FileType = "cr3"
	FileTypeNEF  FileType = "nef"
	FileTypeARW  FileType = "arw"
	FileTypeRAF  FileType = "raf"
	FileTypeORF  FileType = "orf"
	FileTypeRW2  FileType = "rw2"
	FileTypeDNG  FileType = "dng"
)

// IsImageType checks if the file type is a standard image
func (ft FileType) IsImageType() bool {
	return ft == FileTypeJPEG || ft == FileTypePNG
}

// IsRAWType checks if the file type is a RAW format
func (ft FileType) IsRAWType() bool {
	switch ft {
	case FileTypeCR2, FileTypeCR3, FileTypeNEF, FileTypeARW,
		FileTypeRAF, FileTypeORF, FileTypeRW2, FileTypeDNG:
		return true
	default:
		return false
	}
}

// MagicBytes contains file type magic bytes for validation
var MagicBytes = map[FileType][]byte{
	FileTypeJPEG: {0xFF, 0xD8, 0xFF},
	FileTypePNG:  {0x89, 0x50, 0x4E, 0x47},
}

// ValidateFileType validates file type by checking magic bytes
func ValidateFileType(header []byte, allowedTypes []string) (FileType, error) {
	// Check JPEG
	if len(header) >= 3 && header[0] == 0xFF && header[1] == 0xD8 && header[2] == 0xFF {
		if isAllowed("jpg", allowedTypes) || isAllowed("jpeg", allowedTypes) {
			return FileTypeJPEG, nil
		}
	}

	// Check PNG
	if len(header) >= 4 && header[0] == 0x89 && header[1] == 0x50 && header[2] == 0x4E && header[3] == 0x47 {
		if isAllowed("png", allowedTypes) {
			return FileTypePNG, nil
		}
	}

	// For RAW files, we'll rely on extension check since magic bytes vary
	return "", ErrInvalidFileType
}

func isAllowed(ext string, allowedTypes []string) bool {
	for _, t := range allowedTypes {
		if t == ext {
			return true
		}
	}
	return false
}

// GetExtension returns the file extension for a FileType
func (ft FileType) GetExtension() string {
	switch ft {
	case FileTypeJPEG:
		return "jpg"
	case FileTypePNG:
		return "png"
	case FileTypeCR2:
		return "cr2"
	case FileTypeCR3:
		return "cr3"
	case FileTypeNEF:
		return "nef"
	case FileTypeARW:
		return "arw"
	case FileTypeRAF:
		return "raf"
	case FileTypeORF:
		return "orf"
	case FileTypeRW2:
		return "rw2"
	case FileTypeDNG:
		return "dng"
	default:
		return ""
	}
}
