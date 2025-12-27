package ticket

import (
	"QuanPhotos/internal/repository/postgresql"

	"github.com/jmoiron/sqlx"
)

// TicketRepository handles ticket database operations
type TicketRepository struct {
	*postgresql.BaseRepository
}

// NewTicketRepository creates a new ticket repository
func NewTicketRepository(db *sqlx.DB) *TicketRepository {
	return &TicketRepository{
		BaseRepository: postgresql.NewBaseRepository(db),
	}
}
