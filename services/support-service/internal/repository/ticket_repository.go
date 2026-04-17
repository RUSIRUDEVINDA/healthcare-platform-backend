package repository

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/lib/pq"
	"healthcare-platform/services/support-service/internal/model"
)

type TicketRepository struct {
	db *sql.DB
}

func NewTicketRepository(db *sql.DB) *TicketRepository {
	return &TicketRepository{db: db}
}

// Migrate creates the table if it doesn't exist
func (r *TicketRepository) Migrate() error {
	query := `
	CREATE TABLE IF NOT EXISTS reactivation_tickets (
		id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		name       VARCHAR(255) NOT NULL,
		email      VARCHAR(255) NOT NULL,
		reason     TEXT NOT NULL,
		status     VARCHAR(50) NOT NULL DEFAULT 'pending',
		created_at TIMESTAMPTZ DEFAULT NOW(),
		updated_at TIMESTAMPTZ DEFAULT NOW()
	);`
	_, err := r.db.Exec(query)
	return err
}

func (r *TicketRepository) Create(ctx context.Context, ticket *model.ReactivationTicket) error {
	now := time.Now()
	ticket.CreatedAt = now
	ticket.UpdatedAt = now
	ticket.Status = model.StatusPending

	query := `
		INSERT INTO reactivation_tickets (name, email, reason, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`

	return r.db.QueryRowContext(ctx, query,
		ticket.Name, ticket.Email, ticket.Reason,
		ticket.Status, ticket.CreatedAt, ticket.UpdatedAt,
	).Scan(&ticket.ID)
}

func (r *TicketRepository) List(ctx context.Context) ([]model.ReactivationTicket, error) {
	query := `
		SELECT id, name, email, reason, status, created_at, updated_at
		FROM reactivation_tickets
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tickets []model.ReactivationTicket
	for rows.Next() {
		var t model.ReactivationTicket
		if err := rows.Scan(&t.ID, &t.Name, &t.Email, &t.Reason, &t.Status, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		tickets = append(tickets, t)
	}
	return tickets, rows.Err()
}

func (r *TicketRepository) Resolve(ctx context.Context, id string) error {
	query := `UPDATE reactivation_tickets SET status = $1, updated_at = $2 WHERE id = $3`
	_, err := r.db.ExecContext(ctx, query, model.StatusResolved, time.Now(), id)
	return err
}
