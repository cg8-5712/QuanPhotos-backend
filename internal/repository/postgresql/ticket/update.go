package ticket

import (
	"context"

	"QuanPhotos/internal/model"
)

// UpdateStatus updates the status of a ticket
func (r *TicketRepository) UpdateStatus(ctx context.Context, id int64, status model.TicketStatus) error {
	query := `UPDATE tickets SET status = $1, updated_at = NOW() WHERE id = $2`

	result, err := r.DB().ExecContext(ctx, query, status, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return nil
	}

	return nil
}
