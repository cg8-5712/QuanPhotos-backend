package photo

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"

	"QuanPhotos/internal/config"
	exifPkg "QuanPhotos/internal/pkg/exif"
	"QuanPhotos/internal/pkg/imaging"
	"QuanPhotos/internal/pkg/storage"
	"QuanPhotos/internal/repository/postgresql/photo"
)

// UploadRequest represents the upload request parameters
type UploadRequest struct {
	UserID       int64
	File         *multipart.FileHeader
	RawFile      *multipart.FileHeader // Optional
	Title        string
	Description  string
	AircraftType string
	Airline      string
	Registration string
	Airport      string
	CategoryID   int32
	Tags         string // Comma-separated
}

// UploadResponse represents the upload response
type UploadResponse struct {
	ID     int64  `json:"id"`
	Status string `json:"status"`
	Title  string `json:"title"`
}

// Uploader handles photo upload logic
type Uploader struct {
	storage       *storage.LocalStorage
	pathGen       *storage.PathGenerator
	exifParser    *exifPkg.Parser
	imageProc     *imaging.Processor
	photoRepo     *photo.PhotoRepository
	config        *config.Config
	allowedTypes  []string
	maxUploadSize int64
}

// NewUploader creates a new photo uploader
func NewUploader(
	localStorage *storage.LocalStorage,
	photoRepo *photo.PhotoRepository,
	cfg *config.Config,
) *Uploader {
	// Create image processor with config
	procConfig := imaging.ProcessorConfig{
		MaxDimension: cfg.Image.MaxDimension,
		Quality:      cfg.Image.Quality,
		ThumbnailSizes: []imaging.ThumbnailSize{
			{Name: "sm", Width: cfg.Image.ThumbSmWidth, Height: cfg.Image.ThumbSmHeight, Quality: cfg.Image.ThumbSmQuality},
			{Name: "md", Width: cfg.Image.ThumbMdWidth, Height: cfg.Image.ThumbMdHeight, Quality: cfg.Image.ThumbMdQuality},
			{Name: "lg", Width: cfg.Image.ThumbLgWidth, Height: cfg.Image.ThumbLgHeight, Quality: cfg.Image.ThumbLgQuality},
		},
	}

	return &Uploader{
		storage:       localStorage,
		pathGen:       storage.NewPathGenerator(cfg.Storage.Path),
		exifParser:    exifPkg.NewParser(),
		imageProc:     imaging.NewProcessor(procConfig),
		photoRepo:     photoRepo,
		config:        cfg,
		allowedTypes:  cfg.Storage.AllowedTypes,
		maxUploadSize: cfg.Storage.MaxSize,
	}
}

// Upload handles the photo upload process
func (u *Uploader) Upload(ctx context.Context, req *UploadRequest) (*UploadResponse, error) {
	// 1. Validate file
	if err := u.validateFile(req.File); err != nil {
		return nil, err
	}

	// 2. Generate UUID for this upload
	fileUUID := uuid.New().String()

	// 3. Save to temp directory
	tempPath, err := u.saveToTemp(req.File, fileUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to save temp file: %w", err)
	}
	defer u.cleanupTemp(tempPath)

	// 4. Validate file type by magic bytes
	fileType, err := u.validateFileType(tempPath)
	if err != nil {
		return nil, err
	}

	// 5. Parse EXIF data
	exifData, err := u.parseEXIF(tempPath)
	if err != nil {
		// EXIF parsing failure is not critical
		exifData = &exifPkg.Data{}
	}

	// 6. Process image (rotate, resize, generate thumbnails)
	now := time.Now()
	photoDir := u.pathGen.PhotoPath(now, "")
	thumbnailDir := u.pathGen.ThumbnailPath(now, "")

	result, err := u.imageProc.ProcessToSeparateDirs(
		tempPath,
		photoDir,
		thumbnailDir,
		fileUUID,
		exifData.Orientation,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to process image: %w", err)
	}

	// 7. Get file size
	fileInfo, _ := os.Stat(result.MainImagePath)
	var fileSize int64
	if fileInfo != nil {
		fileSize = fileInfo.Size()
	}

	// 8. Prepare database record
	createParams := u.buildCreateParams(req, fileUUID, now, result, exifData, fileSize, fileType)

	// 9. Handle RAW file if present
	if req.RawFile != nil {
		rawPath, err := u.handleRawFile(ctx, req.RawFile, fileUUID, now)
		if err == nil {
			createParams.RawFilePath = &rawPath
		}
	}

	// 10. Save to database
	photoID, err := u.photoRepo.CreateWithTags(ctx, createParams)
	if err != nil {
		// Cleanup files on database error
		u.cleanupProcessedFiles(result)
		return nil, fmt.Errorf("failed to save photo: %w", err)
	}

	return &UploadResponse{
		ID:     photoID,
		Status: "pending",
		Title:  req.Title,
	}, nil
}

// validateFile validates the uploaded file
func (u *Uploader) validateFile(file *multipart.FileHeader) error {
	if file == nil {
		return fmt.Errorf("no file provided")
	}

	// Check file size
	if file.Size > u.maxUploadSize {
		return storage.ErrFileTooLarge
	}

	// Check extension
	ext := strings.ToLower(filepath.Ext(file.Filename))
	ext = strings.TrimPrefix(ext, ".")
	if !u.isAllowedExtension(ext) {
		return storage.ErrInvalidFileType
	}

	return nil
}

// isAllowedExtension checks if the file extension is allowed
func (u *Uploader) isAllowedExtension(ext string) bool {
	for _, allowed := range u.allowedTypes {
		if strings.ToLower(allowed) == ext {
			return true
		}
	}
	return false
}

// saveToTemp saves the uploaded file to temp directory
func (u *Uploader) saveToTemp(file *multipart.FileHeader, fileUUID string) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	ext := strings.ToLower(filepath.Ext(file.Filename))
	tempFileName := fileUUID + ext
	tempPath := u.pathGen.TempPath(tempFileName)

	// Ensure temp directory exists
	if err := os.MkdirAll(filepath.Dir(tempPath), 0755); err != nil {
		return "", err
	}

	dst, err := os.Create(tempPath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return "", err
	}

	return tempPath, nil
}

// validateFileType validates file type by checking magic bytes
func (u *Uploader) validateFileType(path string) (storage.FileType, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	header := make([]byte, 8)
	if _, err := f.Read(header); err != nil {
		return "", err
	}

	return storage.ValidateFileType(header, u.allowedTypes)
}

// parseEXIF parses EXIF data from the image
func (u *Uploader) parseEXIF(path string) (*exifPkg.Data, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Read file into buffer for EXIF parsing
	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, f); err != nil {
		return nil, err
	}

	return u.exifParser.Parse(bytes.NewReader(buf.Bytes()))
}

// buildCreateParams builds the photo creation parameters
func (u *Uploader) buildCreateParams(
	req *UploadRequest,
	fileUUID string,
	now time.Time,
	result *imaging.ProcessResult,
	exifData *exifPkg.Data,
	fileSize int64,
	_ storage.FileType, // fileType reserved for future use
) *photo.CreatePhotoParams {
	// Build relative paths for database storage
	filePath := u.pathGen.RelativePhotoPath(now, fileUUID+".jpg")
	thumbnailPath := u.pathGen.RelativeThumbnailPath(now, fileUUID)

	params := &photo.CreatePhotoParams{
		UserID:        req.UserID,
		Title:         req.Title,
		FilePath:      filePath,
		ThumbnailPath: &thumbnailPath,
		FileSize:      &fileSize,
	}

	// Optional fields
	if req.Description != "" {
		params.Description = &req.Description
	}
	if req.AircraftType != "" {
		params.AircraftType = &req.AircraftType
	}
	if req.Airline != "" {
		params.Airline = &req.Airline
	}
	if req.Registration != "" {
		params.Registration = &req.Registration
	}
	if req.Airport != "" {
		params.Airport = &req.Airport
	}
	if req.CategoryID > 0 {
		params.CategoryID = &req.CategoryID
	}

	// Parse tags
	if req.Tags != "" {
		tags := strings.Split(req.Tags, ",")
		for i, tag := range tags {
			tags[i] = strings.TrimSpace(tag)
		}
		params.Tags = tags
	}

	// EXIF data
	if exifData.CameraMake != "" {
		params.ExifCameraMake = &exifData.CameraMake
	}
	if exifData.CameraModel != "" {
		params.ExifCameraModel = &exifData.CameraModel
	}
	if exifData.SerialNumber != "" {
		params.ExifSerialNumber = &exifData.SerialNumber
	}
	if exifData.LensMake != "" {
		params.ExifLensMake = &exifData.LensMake
	}
	if exifData.LensModel != "" {
		params.ExifLensModel = &exifData.LensModel
	}
	if exifData.FocalLength != "" {
		params.ExifFocalLength = &exifData.FocalLength
	}
	if exifData.FocalLength35mm != "" {
		params.ExifFocalLength35mm = &exifData.FocalLength35mm
	}
	if exifData.Aperture != "" {
		params.ExifAperture = &exifData.Aperture
	}
	if exifData.ShutterSpeed != "" {
		params.ExifShutterSpeed = &exifData.ShutterSpeed
	}
	if exifData.ISO > 0 {
		iso := int32(exifData.ISO)
		params.ExifISO = &iso
	}
	if exifData.ExposureMode != "" {
		params.ExifExposureMode = &exifData.ExposureMode
	}
	if exifData.ExposureProgram != "" {
		params.ExifExposureProgram = &exifData.ExposureProgram
	}
	if exifData.MeteringMode != "" {
		params.ExifMeteringMode = &exifData.MeteringMode
	}
	if exifData.WhiteBalance != "" {
		params.ExifWhiteBalance = &exifData.WhiteBalance
	}
	if exifData.Flash != "" {
		params.ExifFlash = &exifData.Flash
	}
	if exifData.ExposureBias != "" {
		params.ExifExposureBias = &exifData.ExposureBias
	}
	if exifData.TakenAt != nil {
		takenAt := exifData.TakenAt.Format(time.RFC3339)
		params.ExifTakenAt = &takenAt
	}
	if exifData.GPSLatitude != nil {
		params.ExifGPSLatitude = exifData.GPSLatitude
	}
	if exifData.GPSLongitude != nil {
		params.ExifGPSLongitude = exifData.GPSLongitude
	}
	if exifData.GPSAltitude != nil {
		params.ExifGPSAltitude = exifData.GPSAltitude
	}
	if result.Width > 0 {
		width := int32(result.Width)
		params.ExifImageWidth = &width
	}
	if result.Height > 0 {
		height := int32(result.Height)
		params.ExifImageHeight = &height
	}
	if exifData.Orientation > 0 {
		orientation := int32(exifData.Orientation)
		params.ExifOrientation = &orientation
	}
	if exifData.ColorSpace != "" {
		params.ExifColorSpace = &exifData.ColorSpace
	}
	if exifData.Software != "" {
		params.ExifSoftware = &exifData.Software
	}

	return params
}

// handleRawFile handles RAW file upload
func (u *Uploader) handleRawFile(_ context.Context, rawFile *multipart.FileHeader, fileUUID string, now time.Time) (string, error) {
	ext := strings.ToLower(filepath.Ext(rawFile.Filename))

	// Save RAW file
	src, err := rawFile.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	rawPath := u.pathGen.RawPath(now, fileUUID+ext)
	if err := os.MkdirAll(filepath.Dir(rawPath), 0755); err != nil {
		return "", err
	}

	dst, err := os.Create(rawPath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return "", err
	}

	// Return relative path for database
	return u.pathGen.RelativeRawPath(now, fileUUID+ext), nil
}

// cleanupTemp removes temporary file
func (u *Uploader) cleanupTemp(path string) {
	if path != "" {
		os.Remove(path)
	}
}

// cleanupProcessedFiles removes processed files on error
func (u *Uploader) cleanupProcessedFiles(result *imaging.ProcessResult) {
	if result == nil {
		return
	}
	os.Remove(result.MainImagePath)
	for _, path := range result.ThumbnailPaths {
		os.Remove(path)
	}
}
