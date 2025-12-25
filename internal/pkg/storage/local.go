package storage

import (
	"context"
	"io"
	"os"
	"path/filepath"
)

// LocalStorage implements Storage interface for local file system
type LocalStorage struct {
	basePath string
	baseURL  string
}

// NewLocalStorage creates a new local storage instance
func NewLocalStorage(basePath, baseURL string) (*LocalStorage, error) {
	// Ensure base path exists
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, err
	}

	// Create required subdirectories
	subdirs := []string{"photos", "thumbnails", "raw", "temp"}
	for _, dir := range subdirs {
		if err := os.MkdirAll(filepath.Join(basePath, dir), 0755); err != nil {
			return nil, err
		}
	}

	return &LocalStorage{
		basePath: basePath,
		baseURL:  baseURL,
	}, nil
}

// Upload saves a file to the specified path
func (s *LocalStorage) Upload(ctx context.Context, file io.Reader, path string) error {
	fullPath := s.getFullPath(path)

	// Ensure directory exists
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return ErrCreateDirectory
	}

	// Create file
	dst, err := os.Create(fullPath)
	if err != nil {
		return ErrWriteFile
	}
	defer dst.Close()

	// Copy content
	if _, err := io.Copy(dst, file); err != nil {
		return ErrWriteFile
	}

	return nil
}

// Delete removes a file at the specified path
func (s *LocalStorage) Delete(ctx context.Context, path string) error {
	fullPath := s.getFullPath(path)

	if err := os.Remove(fullPath); err != nil {
		if os.IsNotExist(err) {
			return ErrFileNotFound
		}
		return ErrDeleteFile
	}

	return nil
}

// Move moves a file from one path to another
func (s *LocalStorage) Move(ctx context.Context, from, to string) error {
	fromPath := s.getFullPath(from)
	toPath := s.getFullPath(to)

	// Ensure destination directory exists
	dir := filepath.Dir(toPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return ErrCreateDirectory
	}

	// Move file
	if err := os.Rename(fromPath, toPath); err != nil {
		return ErrMoveFile
	}

	return nil
}

// GetURL returns the URL for accessing a file
func (s *LocalStorage) GetURL(path string) string {
	if s.baseURL == "" {
		return path
	}
	return s.baseURL + path
}

// Exists checks if a file exists at the specified path
func (s *LocalStorage) Exists(ctx context.Context, path string) bool {
	fullPath := s.getFullPath(path)
	_, err := os.Stat(fullPath)
	return err == nil
}

// EnsureDir ensures the directory exists
func (s *LocalStorage) EnsureDir(ctx context.Context, path string) error {
	fullPath := s.getFullPath(path)
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return ErrCreateDirectory
	}
	return nil
}

// getFullPath returns the full file system path
func (s *LocalStorage) getFullPath(path string) string {
	// If path starts with /, it's a relative path from base
	if len(path) > 0 && path[0] == '/' {
		return s.basePath + path
	}
	return filepath.Join(s.basePath, path)
}

// SaveToTemp saves a file to the temp directory and returns the temp path
func (s *LocalStorage) SaveToTemp(ctx context.Context, file io.Reader, filename string) (string, error) {
	tempPath := "/temp/" + filename
	if err := s.Upload(ctx, file, tempPath); err != nil {
		return "", err
	}
	return tempPath, nil
}

// GetBasePath returns the base path
func (s *LocalStorage) GetBasePath() string {
	return s.basePath
}

// GetAbsolutePath returns the absolute file system path for a relative path
func (s *LocalStorage) GetAbsolutePath(relativePath string) string {
	return s.getFullPath(relativePath)
}
