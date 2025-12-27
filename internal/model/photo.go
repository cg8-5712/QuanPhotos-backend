package model

import (
	"database/sql"
	"time"
)

// PhotoStatus represents photo review status
type PhotoStatus string

const (
	PhotoStatusPending    PhotoStatus = "pending"
	PhotoStatusAIPassed   PhotoStatus = "ai_passed"
	PhotoStatusAIRejected PhotoStatus = "ai_rejected"
	PhotoStatusApproved   PhotoStatus = "approved"
	PhotoStatusRejected   PhotoStatus = "rejected"
)

// Photo represents a photo in the system
type Photo struct {
	ID            int64          `db:"id" json:"id"`
	UserID        int64          `db:"user_id" json:"user_id"`
	CategoryID    sql.NullInt32  `db:"category_id" json:"-"`
	Title         string         `db:"title" json:"title"`
	Description   sql.NullString `db:"description" json:"-"`
	FilePath      string         `db:"file_path" json:"-"`
	ThumbnailPath sql.NullString `db:"thumbnail_path" json:"-"`
	RawFilePath   sql.NullString `db:"raw_file_path" json:"-"`
	FileSize      sql.NullInt64  `db:"file_size" json:"-"`
	Status        PhotoStatus    `db:"status" json:"status"`
	ViewCount     int            `db:"view_count" json:"view_count"`
	LikeCount     int            `db:"like_count" json:"like_count"`
	FavoriteCount int            `db:"favorite_count" json:"favorite_count"`
	CommentCount  int            `db:"comment_count" json:"comment_count"`
	ShareCount    int            `db:"share_count" json:"share_count"`

	// Aviation info
	AircraftType sql.NullString `db:"aircraft_type" json:"-"`
	Airline      sql.NullString `db:"airline" json:"-"`
	Registration sql.NullString `db:"registration" json:"-"`
	Airport      sql.NullString `db:"airport" json:"-"`

	// EXIF Camera info
	ExifCameraMake   sql.NullString `db:"exif_camera_make" json:"-"`
	ExifCameraModel  sql.NullString `db:"exif_camera_model" json:"-"`
	ExifSerialNumber sql.NullString `db:"exif_serial_number" json:"-"`

	// EXIF Lens info
	ExifLensMake        sql.NullString `db:"exif_lens_make" json:"-"`
	ExifLensModel       sql.NullString `db:"exif_lens_model" json:"-"`
	ExifFocalLength     sql.NullString `db:"exif_focal_length" json:"-"`
	ExifFocalLength35mm sql.NullString `db:"exif_focal_length_35mm" json:"-"`

	// EXIF Shooting parameters
	ExifAperture        sql.NullString `db:"exif_aperture" json:"-"`
	ExifShutterSpeed    sql.NullString `db:"exif_shutter_speed" json:"-"`
	ExifISO             sql.NullInt32  `db:"exif_iso" json:"-"`
	ExifExposureMode    sql.NullString `db:"exif_exposure_mode" json:"-"`
	ExifExposureProgram sql.NullString `db:"exif_exposure_program" json:"-"`
	ExifMeteringMode    sql.NullString `db:"exif_metering_mode" json:"-"`
	ExifWhiteBalance    sql.NullString `db:"exif_white_balance" json:"-"`
	ExifFlash           sql.NullString `db:"exif_flash" json:"-"`
	ExifExposureBias    sql.NullString `db:"exif_exposure_bias" json:"-"`

	// EXIF Time and location
	ExifTakenAt      sql.NullTime    `db:"exif_taken_at" json:"-"`
	ExifGPSLatitude  sql.NullFloat64 `db:"exif_gps_latitude" json:"-"`
	ExifGPSLongitude sql.NullFloat64 `db:"exif_gps_longitude" json:"-"`
	ExifGPSAltitude  sql.NullFloat64 `db:"exif_gps_altitude" json:"-"`

	// EXIF Image info
	ExifImageWidth  sql.NullInt32  `db:"exif_image_width" json:"-"`
	ExifImageHeight sql.NullInt32  `db:"exif_image_height" json:"-"`
	ExifOrientation sql.NullInt32  `db:"exif_orientation" json:"-"`
	ExifColorSpace  sql.NullString `db:"exif_color_space" json:"-"`
	ExifSoftware    sql.NullString `db:"exif_software" json:"-"`

	// Timestamps
	ApprovedAt sql.NullTime `db:"approved_at" json:"-"`
	CreatedAt  time.Time    `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time    `db:"updated_at" json:"updated_at"`
}

// PhotoListItem represents a photo in list view
type PhotoListItem struct {
	ID            int64      `json:"id"`
	Title         string     `json:"title"`
	ThumbnailURL  string     `json:"thumbnail_url"`
	AircraftType  *string    `json:"aircraft_type,omitempty"`
	Airline       *string    `json:"airline,omitempty"`
	Airport       *string    `json:"airport,omitempty"`
	Registration  *string    `json:"registration,omitempty"`
	ViewCount     int        `json:"view_count"`
	LikeCount     int        `json:"like_count"`
	FavoriteCount int        `json:"favorite_count"`
	CommentCount  int        `json:"comment_count"`
	CreatedAt     string     `json:"created_at"`
	User          *UserBrief `json:"user"`
}

// UserBrief represents brief user info for photo list
type UserBrief struct {
	ID       int64   `json:"id"`
	Username string  `json:"username"`
	Avatar   *string `json:"avatar,omitempty"`
}

// PhotoDetail represents detailed photo information
type PhotoDetail struct {
	ID            int64          `json:"id"`
	Title         string         `json:"title"`
	Description   *string        `json:"description,omitempty"`
	ImageURL      string         `json:"image_url"`
	ThumbnailURL  string         `json:"thumbnail_url"`
	HasRAW        bool           `json:"has_raw"`
	Status        PhotoStatus    `json:"status"`
	AircraftType  *string        `json:"aircraft_type,omitempty"`
	Airline       *string        `json:"airline,omitempty"`
	Registration  *string        `json:"registration,omitempty"`
	Airport       *string        `json:"airport,omitempty"`
	Category      *CategoryBrief `json:"category,omitempty"`
	Tags          []string       `json:"tags"`
	EXIF          *PhotoEXIF     `json:"exif,omitempty"`
	ViewCount     int            `json:"view_count"`
	LikeCount     int            `json:"like_count"`
	FavoriteCount int            `json:"favorite_count"`
	CommentCount  int            `json:"comment_count"`
	IsFavorited   bool           `json:"is_favorited"`
	IsLiked       bool           `json:"is_liked"`
	CreatedAt     string         `json:"created_at"`
	ApprovedAt    *string        `json:"approved_at,omitempty"`
	User          *UserBrief     `json:"user"`
}

// CategoryBrief represents brief category info
type CategoryBrief struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
}

// PhotoEXIF represents EXIF information
type PhotoEXIF struct {
	CameraMake      *string  `json:"camera_make,omitempty"`
	CameraModel     *string  `json:"camera_model,omitempty"`
	LensModel       *string  `json:"lens_model,omitempty"`
	FocalLength     *string  `json:"focal_length,omitempty"`
	FocalLength35mm *string  `json:"focal_length_35mm,omitempty"`
	Aperture        *string  `json:"aperture,omitempty"`
	ShutterSpeed    *string  `json:"shutter_speed,omitempty"`
	ISO             *int32   `json:"iso,omitempty"`
	ExposureMode    *string  `json:"exposure_mode,omitempty"`
	MeteringMode    *string  `json:"metering_mode,omitempty"`
	WhiteBalance    *string  `json:"white_balance,omitempty"`
	Flash           *string  `json:"flash,omitempty"`
	TakenAt         *string  `json:"taken_at,omitempty"`
	GPSLatitude     *float64 `json:"gps_latitude,omitempty"`
	GPSLongitude    *float64 `json:"gps_longitude,omitempty"`
	ImageWidth      *int32   `json:"image_width,omitempty"`
	ImageHeight     *int32   `json:"image_height,omitempty"`
}

// ToListItem converts Photo to PhotoListItem
func (p *Photo) ToListItem(user *UserBrief, baseURL string) *PhotoListItem {
	item := &PhotoListItem{
		ID:            p.ID,
		Title:         p.Title,
		ViewCount:     p.ViewCount,
		LikeCount:     p.LikeCount,
		FavoriteCount: p.FavoriteCount,
		CommentCount:  p.CommentCount,
		CreatedAt:     p.CreatedAt.Format(time.RFC3339),
		User:          user,
	}

	if p.ThumbnailPath.Valid {
		item.ThumbnailURL = baseURL + p.ThumbnailPath.String
	}
	if p.AircraftType.Valid {
		item.AircraftType = &p.AircraftType.String
	}
	if p.Airline.Valid {
		item.Airline = &p.Airline.String
	}
	if p.Airport.Valid {
		item.Airport = &p.Airport.String
	}
	if p.Registration.Valid {
		item.Registration = &p.Registration.String
	}

	return item
}

// ToDetail converts Photo to PhotoDetail
func (p *Photo) ToDetail(user *UserBrief, category *CategoryBrief, tags []string, baseURL string, isFavorited, isLiked bool) *PhotoDetail {
	detail := &PhotoDetail{
		ID:            p.ID,
		Title:         p.Title,
		HasRAW:        p.RawFilePath.Valid,
		Status:        p.Status,
		Tags:          tags,
		ViewCount:     p.ViewCount,
		LikeCount:     p.LikeCount,
		FavoriteCount: p.FavoriteCount,
		CommentCount:  p.CommentCount,
		IsFavorited:   isFavorited,
		IsLiked:       isLiked,
		CreatedAt:     p.CreatedAt.Format(time.RFC3339),
		User:          user,
		Category:      category,
	}

	detail.ImageURL = baseURL + p.FilePath
	if p.ThumbnailPath.Valid {
		detail.ThumbnailURL = baseURL + p.ThumbnailPath.String
	}
	if p.Description.Valid {
		detail.Description = &p.Description.String
	}
	if p.AircraftType.Valid {
		detail.AircraftType = &p.AircraftType.String
	}
	if p.Airline.Valid {
		detail.Airline = &p.Airline.String
	}
	if p.Registration.Valid {
		detail.Registration = &p.Registration.String
	}
	if p.Airport.Valid {
		detail.Airport = &p.Airport.String
	}
	if p.ApprovedAt.Valid {
		approvedAt := p.ApprovedAt.Time.Format(time.RFC3339)
		detail.ApprovedAt = &approvedAt
	}

	// Build EXIF
	detail.EXIF = p.buildEXIF()

	return detail
}

func (p *Photo) buildEXIF() *PhotoEXIF {
	exif := &PhotoEXIF{}
	hasData := false

	if p.ExifCameraMake.Valid {
		exif.CameraMake = &p.ExifCameraMake.String
		hasData = true
	}
	if p.ExifCameraModel.Valid {
		exif.CameraModel = &p.ExifCameraModel.String
		hasData = true
	}
	if p.ExifLensModel.Valid {
		exif.LensModel = &p.ExifLensModel.String
		hasData = true
	}
	if p.ExifFocalLength.Valid {
		exif.FocalLength = &p.ExifFocalLength.String
		hasData = true
	}
	if p.ExifFocalLength35mm.Valid {
		exif.FocalLength35mm = &p.ExifFocalLength35mm.String
		hasData = true
	}
	if p.ExifAperture.Valid {
		exif.Aperture = &p.ExifAperture.String
		hasData = true
	}
	if p.ExifShutterSpeed.Valid {
		exif.ShutterSpeed = &p.ExifShutterSpeed.String
		hasData = true
	}
	if p.ExifISO.Valid {
		exif.ISO = &p.ExifISO.Int32
		hasData = true
	}
	if p.ExifExposureMode.Valid {
		exif.ExposureMode = &p.ExifExposureMode.String
		hasData = true
	}
	if p.ExifMeteringMode.Valid {
		exif.MeteringMode = &p.ExifMeteringMode.String
		hasData = true
	}
	if p.ExifWhiteBalance.Valid {
		exif.WhiteBalance = &p.ExifWhiteBalance.String
		hasData = true
	}
	if p.ExifFlash.Valid {
		exif.Flash = &p.ExifFlash.String
		hasData = true
	}
	if p.ExifTakenAt.Valid {
		takenAt := p.ExifTakenAt.Time.Format(time.RFC3339)
		exif.TakenAt = &takenAt
		hasData = true
	}
	if p.ExifGPSLatitude.Valid {
		exif.GPSLatitude = &p.ExifGPSLatitude.Float64
		hasData = true
	}
	if p.ExifGPSLongitude.Valid {
		exif.GPSLongitude = &p.ExifGPSLongitude.Float64
		hasData = true
	}
	if p.ExifImageWidth.Valid {
		exif.ImageWidth = &p.ExifImageWidth.Int32
		hasData = true
	}
	if p.ExifImageHeight.Valid {
		exif.ImageHeight = &p.ExifImageHeight.Int32
		hasData = true
	}

	if !hasData {
		return nil
	}
	return exif
}

// Favorite represents a user's favorite photo
type Favorite struct {
	UserID    int64     `db:"user_id"`
	PhotoID   int64     `db:"photo_id"`
	CreatedAt time.Time `db:"created_at"`
}

// PhotoLike represents a user's like on a photo
type PhotoLike struct {
	UserID    int64     `db:"user_id"`
	PhotoID   int64     `db:"photo_id"`
	CreatedAt time.Time `db:"created_at"`
}

// Category represents a photo category
type Category struct {
	ID          int32          `db:"id" json:"id"`
	Name        string         `db:"name" json:"name"`
	NameEN      string         `db:"name_en" json:"name_en"`
	Description sql.NullString `db:"description" json:"-"`
	SortOrder   int            `db:"sort_order" json:"sort_order"`
	PhotoCount  int            `db:"photo_count" json:"photo_count,omitempty"`
	CreatedAt   time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time      `db:"updated_at" json:"updated_at"`
}

// Tag represents a photo tag
type Tag struct {
	ID         int32     `db:"id" json:"id"`
	Name       string    `db:"name" json:"name"`
	PhotoCount int       `db:"photo_count" json:"photo_count"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
}
