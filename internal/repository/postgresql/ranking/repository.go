package ranking

import (
	"context"
	"fmt"
	"time"

	"QuanPhotos/internal/model"
	"QuanPhotos/internal/repository/postgresql"

	"github.com/jmoiron/sqlx"
)

// RankingRepository handles ranking database operations
type RankingRepository struct {
	*postgresql.BaseRepository
}

// NewRankingRepository creates a new ranking repository
func NewRankingRepository(db *sqlx.DB) *RankingRepository {
	return &RankingRepository{
		BaseRepository: postgresql.NewBaseRepository(db),
	}
}

// PhotoRankingItem represents a photo in ranking
type PhotoRankingItem struct {
	ID            int64   `db:"id" json:"id"`
	Title         string  `db:"title" json:"title"`
	UserID        int64   `db:"user_id" json:"user_id"`
	ThumbnailPath *string `db:"thumbnail_path" json:"thumbnail_path"`
	LikeCount     int     `db:"like_count" json:"like_count"`
	ViewCount     int     `db:"view_count" json:"view_count"`
	CommentCount  int     `db:"comment_count" json:"comment_count"`
	ShareCount    int     `db:"share_count" json:"share_count"`
}

// UserRankingItem represents a user in ranking
type UserRankingItem struct {
	ID          int64   `db:"id" json:"id"`
	Username    string  `db:"username" json:"username"`
	Avatar      *string `db:"avatar" json:"avatar"`
	PhotoCount  int     `db:"photo_count" json:"photo_count"`
	TotalLikes  int     `db:"total_likes" json:"total_likes"`
	TotalViews  int     `db:"total_views" json:"total_views"`
}

// RankingParams contains parameters for ranking queries
type RankingParams struct {
	Period   string // day, week, month, all
	Limit    int
	SortBy   string // like_count, view_count, comment_count for photos; photo_count, total_likes for users
}

// GetTimeRange returns start time based on period
func GetTimeRange(period string) *time.Time {
	now := time.Now()
	var start time.Time

	switch period {
	case "day":
		start = now.AddDate(0, 0, -1)
	case "week":
		start = now.AddDate(0, 0, -7)
	case "month":
		start = now.AddDate(0, -1, 0)
	default:
		return nil // all time
	}

	return &start
}

// GetPhotoRanking retrieves photo ranking
func (r *RankingRepository) GetPhotoRanking(ctx context.Context, params RankingParams) ([]*PhotoRankingItem, error) {
	if params.Limit < 1 {
		params.Limit = 20
	}
	if params.Limit > 100 {
		params.Limit = 100
	}

	if params.SortBy == "" {
		params.SortBy = "like_count"
	}

	// Validate sort column
	validSortColumns := map[string]bool{
		"like_count":    true,
		"view_count":    true,
		"comment_count": true,
		"share_count":   true,
	}
	if !validSortColumns[params.SortBy] {
		params.SortBy = "like_count"
	}

	// Build query
	var whereClause string
	var args []interface{}
	argIndex := 1

	whereClause = "WHERE status = 'approved'"

	startTime := GetTimeRange(params.Period)
	if startTime != nil {
		whereClause += fmt.Sprintf(" AND created_at >= $%d", argIndex)
		args = append(args, *startTime)
		argIndex++
	}

	query := fmt.Sprintf(`
		SELECT id, title, user_id, thumbnail_path, like_count, view_count, comment_count, share_count
		FROM photos
		%s
		ORDER BY %s DESC
		LIMIT $%d
	`, whereClause, params.SortBy, argIndex)

	args = append(args, params.Limit)

	var items []*PhotoRankingItem
	err := r.DB().SelectContext(ctx, &items, query, args...)
	if err != nil {
		return nil, err
	}

	return items, nil
}

// GetUserRanking retrieves user ranking
func (r *RankingRepository) GetUserRanking(ctx context.Context, params RankingParams) ([]*UserRankingItem, error) {
	if params.Limit < 1 {
		params.Limit = 20
	}
	if params.Limit > 100 {
		params.Limit = 100
	}

	if params.SortBy == "" {
		params.SortBy = "total_likes"
	}

	// Validate sort column
	validSortColumns := map[string]bool{
		"photo_count": true,
		"total_likes": true,
		"total_views": true,
	}
	if !validSortColumns[params.SortBy] {
		params.SortBy = "total_likes"
	}

	// Build query
	var whereClause string
	var args []interface{}
	argIndex := 1

	startTime := GetTimeRange(params.Period)
	if startTime != nil {
		whereClause = fmt.Sprintf("WHERE p.created_at >= $%d", argIndex)
		args = append(args, *startTime)
		argIndex++
	}

	query := fmt.Sprintf(`
		SELECT
			u.id,
			u.username,
			u.avatar,
			COUNT(DISTINCT p.id) as photo_count,
			COALESCE(SUM(p.like_count), 0) as total_likes,
			COALESCE(SUM(p.view_count), 0) as total_views
		FROM users u
		LEFT JOIN photos p ON u.id = p.user_id AND p.status = 'approved' %s
		WHERE u.status = 'active' AND u.role != 'guest'
		GROUP BY u.id, u.username, u.avatar
		HAVING COUNT(DISTINCT p.id) > 0
		ORDER BY %s DESC
		LIMIT $%d
	`, func() string {
		if whereClause != "" {
			return "AND p.created_at >= $1"
		}
		return ""
	}(), params.SortBy, argIndex)

	args = append(args, params.Limit)

	var items []*UserRankingItem
	err := r.DB().SelectContext(ctx, &items, query, args...)
	if err != nil {
		return nil, err
	}

	return items, nil
}

// GetUserMap retrieves a map of users by their IDs
func (r *RankingRepository) GetUserMap(ctx context.Context, userIDs []int64) (map[int64]*model.User, error) {
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
