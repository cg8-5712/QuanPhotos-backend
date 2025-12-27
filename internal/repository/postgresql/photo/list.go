package photo

import (
	"context"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"

	"QuanPhotos/internal/model"
)

// ListParams contains parameters for listing photos
type ListParams struct {
	Page         int
	PageSize     int
	Status       string
	CategoryID   int32
	UserID       int64
	AircraftType string
	Airline      string
	Airport      string
	Registration string
	Keyword      string
	TakenFrom    string // Date in format "2006-01-02"
	TakenTo      string // Date in format "2006-01-02"
	SortBy       string // created_at, view_count, like_count, favorite_count
	SortOrder    string // asc, desc
}

// ListResult contains the result of listing photos
type ListResult struct {
	Photos     []*model.Photo
	Total      int64
	Page       int
	PageSize   int
	TotalPages int
}

// List retrieves a paginated list of photos with optional filters
func (r *PhotoRepository) List(ctx context.Context, params ListParams) (*ListResult, error) {
	// Set defaults
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize < 1 {
		params.PageSize = 20
	}
	if params.PageSize > 100 {
		params.PageSize = 100
	}
	if params.SortBy == "" {
		params.SortBy = "created_at"
	}
	if params.SortOrder == "" {
		params.SortOrder = "desc"
	}

	// Validate sort fields
	validSortFields := map[string]bool{
		"created_at":     true,
		"view_count":     true,
		"like_count":     true,
		"favorite_count": true,
	}
	if !validSortFields[params.SortBy] {
		params.SortBy = "created_at"
	}
	if params.SortOrder != "asc" && params.SortOrder != "desc" {
		params.SortOrder = "desc"
	}

	// Build WHERE clause
	var conditions []string
	var args []interface{}
	argIndex := 1

	// Default: only show approved photos for public listing
	if params.Status == "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, model.PhotoStatusApproved)
		argIndex++
	} else if params.Status != "all" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, params.Status)
		argIndex++
	}

	if params.CategoryID > 0 {
		conditions = append(conditions, fmt.Sprintf("category_id = $%d", argIndex))
		args = append(args, params.CategoryID)
		argIndex++
	}

	if params.UserID > 0 {
		conditions = append(conditions, fmt.Sprintf("user_id = $%d", argIndex))
		args = append(args, params.UserID)
		argIndex++
	}

	if params.AircraftType != "" {
		conditions = append(conditions, fmt.Sprintf("aircraft_type ILIKE $%d", argIndex))
		args = append(args, "%"+params.AircraftType+"%")
		argIndex++
	}

	if params.Airline != "" {
		conditions = append(conditions, fmt.Sprintf("airline ILIKE $%d", argIndex))
		args = append(args, "%"+params.Airline+"%")
		argIndex++
	}

	if params.Airport != "" {
		conditions = append(conditions, fmt.Sprintf("airport ILIKE $%d", argIndex))
		args = append(args, "%"+params.Airport+"%")
		argIndex++
	}

	if params.Registration != "" {
		conditions = append(conditions, fmt.Sprintf("registration ILIKE $%d", argIndex))
		args = append(args, "%"+params.Registration+"%")
		argIndex++
	}

	if params.Keyword != "" {
		conditions = append(conditions, fmt.Sprintf("(title ILIKE $%d OR description ILIKE $%d OR aircraft_type ILIKE $%d OR registration ILIKE $%d)", argIndex, argIndex, argIndex, argIndex))
		args = append(args, "%"+params.Keyword+"%")
		argIndex++
	}

	if params.TakenFrom != "" {
		conditions = append(conditions, fmt.Sprintf("exif_taken_at >= $%d", argIndex))
		args = append(args, params.TakenFrom)
		argIndex++
	}

	if params.TakenTo != "" {
		conditions = append(conditions, fmt.Sprintf("exif_taken_at <= $%d::date + interval '1 day'", argIndex))
		args = append(args, params.TakenTo)
		argIndex++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM photos %s", whereClause)
	var total int64
	err := r.DB().GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, err
	}

	// Calculate pagination
	offset := (params.Page - 1) * params.PageSize
	totalPages := int(total) / params.PageSize
	if int(total)%params.PageSize > 0 {
		totalPages++
	}

	// Query photos
	query := fmt.Sprintf(`
		SELECT * FROM photos
		%s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d
	`, whereClause, params.SortBy, strings.ToUpper(params.SortOrder), argIndex, argIndex+1)

	args = append(args, params.PageSize, offset)

	var photos []*model.Photo
	err = r.DB().SelectContext(ctx, &photos, query, args...)
	if err != nil {
		return nil, err
	}

	return &ListResult{
		Photos:     photos,
		Total:      total,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: totalPages,
	}, nil
}

// ListUserFavorites retrieves user's favorite photos
func (r *PhotoRepository) ListUserFavorites(ctx context.Context, userID int64, page, pageSize int) (*ListResult, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	// Count total
	var total int64
	countQuery := `SELECT COUNT(*) FROM favorites WHERE user_id = $1`
	err := r.DB().GetContext(ctx, &total, countQuery, userID)
	if err != nil {
		return nil, err
	}

	// Calculate pagination
	offset := (page - 1) * pageSize
	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	// Query photos
	query := `
		SELECT p.* FROM photos p
		INNER JOIN favorites f ON f.photo_id = p.id
		WHERE f.user_id = $1
		ORDER BY f.created_at DESC
		LIMIT $2 OFFSET $3
	`

	var photos []*model.Photo
	err = r.DB().SelectContext(ctx, &photos, query, userID, pageSize, offset)
	if err != nil {
		return nil, err
	}

	return &ListResult{
		Photos:     photos,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// GetUserMap retrieves user info for a list of user IDs
func (r *PhotoRepository) GetUserMap(ctx context.Context, userIDs []int64) (map[int64]*model.User, error) {
	if len(userIDs) == 0 {
		return make(map[int64]*model.User), nil
	}

	query, args, err := sqlx.In(`SELECT * FROM users WHERE id IN (?)`, userIDs)
	if err != nil {
		return nil, err
	}
	query = r.DB().Rebind(query)

	var users []*model.User
	err = r.DB().SelectContext(ctx, &users, query, args...)
	if err != nil {
		return nil, err
	}

	result := make(map[int64]*model.User)
	for _, u := range users {
		result[u.ID] = u
	}

	return result, nil
}
