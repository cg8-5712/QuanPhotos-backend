package photo

import (
	"context"
	"database/sql"
	"strings"

	"QuanPhotos/internal/model"
)

// CreatePhotoParams contains parameters for creating a photo
type CreatePhotoParams struct {
	UserID        int64
	CategoryID    *int32
	Title         string
	Description   *string
	FilePath      string
	ThumbnailPath *string
	RawFilePath   *string
	FileSize      *int64
	AircraftType  *string
	Airline       *string
	Registration  *string
	Airport       *string

	// EXIF Camera info
	ExifCameraMake   *string
	ExifCameraModel  *string
	ExifSerialNumber *string

	// EXIF Lens info
	ExifLensMake        *string
	ExifLensModel       *string
	ExifFocalLength     *string
	ExifFocalLength35mm *string

	// EXIF Shooting parameters
	ExifAperture        *string
	ExifShutterSpeed    *string
	ExifISO             *int32
	ExifExposureMode    *string
	ExifExposureProgram *string
	ExifMeteringMode    *string
	ExifWhiteBalance    *string
	ExifFlash           *string
	ExifExposureBias    *string

	// EXIF Time and location
	ExifTakenAt      *string // RFC3339 format
	ExifGPSLatitude  *float64
	ExifGPSLongitude *float64
	ExifGPSAltitude  *float64

	// EXIF Image info
	ExifImageWidth  *int32
	ExifImageHeight *int32
	ExifOrientation *int32
	ExifColorSpace  *string
	ExifSoftware    *string

	// Tags
	Tags []string
}

// Create creates a new photo record
func (r *PhotoRepository) Create(ctx context.Context, params *CreatePhotoParams) (int64, error) {
	query := `
		INSERT INTO photos (
			user_id, category_id, title, description, file_path, thumbnail_path, raw_file_path, file_size,
			aircraft_type, airline, registration, airport,
			exif_camera_make, exif_camera_model, exif_serial_number,
			exif_lens_make, exif_lens_model, exif_focal_length, exif_focal_length_35mm,
			exif_aperture, exif_shutter_speed, exif_iso, exif_exposure_mode, exif_exposure_program,
			exif_metering_mode, exif_white_balance, exif_flash, exif_exposure_bias,
			exif_taken_at, exif_gps_latitude, exif_gps_longitude, exif_gps_altitude,
			exif_image_width, exif_image_height, exif_orientation, exif_color_space, exif_software,
			status
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8,
			$9, $10, $11, $12,
			$13, $14, $15,
			$16, $17, $18, $19,
			$20, $21, $22, $23, $24,
			$25, $26, $27, $28,
			$29, $30, $31, $32,
			$33, $34, $35, $36, $37,
			$38
		) RETURNING id
	`

	var id int64
	err := r.DB().QueryRowContext(ctx, query,
		params.UserID,
		toNullInt32(params.CategoryID),
		params.Title,
		toNullString(params.Description),
		params.FilePath,
		toNullString(params.ThumbnailPath),
		toNullString(params.RawFilePath),
		toNullInt64(params.FileSize),
		toNullString(params.AircraftType),
		toNullString(params.Airline),
		toNullString(params.Registration),
		toNullString(params.Airport),
		toNullString(params.ExifCameraMake),
		toNullString(params.ExifCameraModel),
		toNullString(params.ExifSerialNumber),
		toNullString(params.ExifLensMake),
		toNullString(params.ExifLensModel),
		toNullString(params.ExifFocalLength),
		toNullString(params.ExifFocalLength35mm),
		toNullString(params.ExifAperture),
		toNullString(params.ExifShutterSpeed),
		toNullInt32(params.ExifISO),
		toNullString(params.ExifExposureMode),
		toNullString(params.ExifExposureProgram),
		toNullString(params.ExifMeteringMode),
		toNullString(params.ExifWhiteBalance),
		toNullString(params.ExifFlash),
		toNullString(params.ExifExposureBias),
		toNullString(params.ExifTakenAt),
		toNullFloat64(params.ExifGPSLatitude),
		toNullFloat64(params.ExifGPSLongitude),
		toNullFloat64(params.ExifGPSAltitude),
		toNullInt32(params.ExifImageWidth),
		toNullInt32(params.ExifImageHeight),
		toNullInt32(params.ExifOrientation),
		toNullString(params.ExifColorSpace),
		toNullString(params.ExifSoftware),
		model.PhotoStatusPending,
	).Scan(&id)

	if err != nil {
		return 0, err
	}

	return id, nil
}

// CreateWithTags creates a photo with tags in a transaction
func (r *PhotoRepository) CreateWithTags(ctx context.Context, params *CreatePhotoParams) (int64, error) {
	tx, err := r.DB().BeginTxx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	// Create photo
	query := `
		INSERT INTO photos (
			user_id, category_id, title, description, file_path, thumbnail_path, raw_file_path, file_size,
			aircraft_type, airline, registration, airport,
			exif_camera_make, exif_camera_model, exif_serial_number,
			exif_lens_make, exif_lens_model, exif_focal_length, exif_focal_length_35mm,
			exif_aperture, exif_shutter_speed, exif_iso, exif_exposure_mode, exif_exposure_program,
			exif_metering_mode, exif_white_balance, exif_flash, exif_exposure_bias,
			exif_taken_at, exif_gps_latitude, exif_gps_longitude, exif_gps_altitude,
			exif_image_width, exif_image_height, exif_orientation, exif_color_space, exif_software,
			status
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8,
			$9, $10, $11, $12,
			$13, $14, $15,
			$16, $17, $18, $19,
			$20, $21, $22, $23, $24,
			$25, $26, $27, $28,
			$29, $30, $31, $32,
			$33, $34, $35, $36, $37,
			$38
		) RETURNING id
	`

	var photoID int64
	err = tx.QueryRowContext(ctx, query,
		params.UserID,
		toNullInt32(params.CategoryID),
		params.Title,
		toNullString(params.Description),
		params.FilePath,
		toNullString(params.ThumbnailPath),
		toNullString(params.RawFilePath),
		toNullInt64(params.FileSize),
		toNullString(params.AircraftType),
		toNullString(params.Airline),
		toNullString(params.Registration),
		toNullString(params.Airport),
		toNullString(params.ExifCameraMake),
		toNullString(params.ExifCameraModel),
		toNullString(params.ExifSerialNumber),
		toNullString(params.ExifLensMake),
		toNullString(params.ExifLensModel),
		toNullString(params.ExifFocalLength),
		toNullString(params.ExifFocalLength35mm),
		toNullString(params.ExifAperture),
		toNullString(params.ExifShutterSpeed),
		toNullInt32(params.ExifISO),
		toNullString(params.ExifExposureMode),
		toNullString(params.ExifExposureProgram),
		toNullString(params.ExifMeteringMode),
		toNullString(params.ExifWhiteBalance),
		toNullString(params.ExifFlash),
		toNullString(params.ExifExposureBias),
		toNullString(params.ExifTakenAt),
		toNullFloat64(params.ExifGPSLatitude),
		toNullFloat64(params.ExifGPSLongitude),
		toNullFloat64(params.ExifGPSAltitude),
		toNullInt32(params.ExifImageWidth),
		toNullInt32(params.ExifImageHeight),
		toNullInt32(params.ExifOrientation),
		toNullString(params.ExifColorSpace),
		toNullString(params.ExifSoftware),
		model.PhotoStatusPending,
	).Scan(&photoID)

	if err != nil {
		return 0, err
	}

	// Create tags if any
	if len(params.Tags) > 0 {
		for _, tagName := range params.Tags {
			tagName = strings.TrimSpace(tagName)
			if tagName == "" {
				continue
			}

			// Insert or get tag
			var tagID int32
			err = tx.QueryRowContext(ctx, `
				INSERT INTO tags (name) VALUES ($1)
				ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
				RETURNING id
			`, tagName).Scan(&tagID)
			if err != nil {
				return 0, err
			}

			// Link tag to photo
			_, err = tx.ExecContext(ctx, `
				INSERT INTO photo_tags (photo_id, tag_id) VALUES ($1, $2)
				ON CONFLICT DO NOTHING
			`, photoID, tagID)
			if err != nil {
				return 0, err
			}
		}
	}

	if err = tx.Commit(); err != nil {
		return 0, err
	}

	return photoID, nil
}

// Helper functions for converting pointers to sql.Null* types

func toNullString(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *s, Valid: true}
}

func toNullInt32(i *int32) sql.NullInt32 {
	if i == nil {
		return sql.NullInt32{}
	}
	return sql.NullInt32{Int32: *i, Valid: true}
}

func toNullInt64(i *int64) sql.NullInt64 {
	if i == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: *i, Valid: true}
}

func toNullFloat64(f *float64) sql.NullFloat64 {
	if f == nil {
		return sql.NullFloat64{}
	}
	return sql.NullFloat64{Float64: *f, Valid: true}
}
