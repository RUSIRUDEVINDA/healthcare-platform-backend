package repository

import (
	"context"
	"database/sql"

	"healthcare-platform/services/ai-symptom-service/internal/model"
)

// ChatRepository persists AI symptom check turns per user.
type ChatRepository struct {
	db *sql.DB
}

func NewChatRepository(db *sql.DB) *ChatRepository {
	return &ChatRepository{db: db}
}

func (r *ChatRepository) Save(ctx context.Context, userID string, req *model.SymptomCheckRequest, resp *model.SymptomCheckResponse) (int64, error) {
	var id int64
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO symptom_chat_history (user_id, symptoms, optional_context, suggested_specialty, preliminary_notes, disclaimer)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`,
		userID, req.Symptoms, req.OptionalContext, resp.SuggestedSpecialty, resp.PreliminaryNotes, resp.Disclaimer,
	).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *ChatRepository) ListByUser(ctx context.Context, userID string, limit, offset int) ([]model.SymptomChatHistoryItem, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, symptoms, optional_context, suggested_specialty, preliminary_notes, disclaimer, created_at
		FROM symptom_chat_history
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.SymptomChatHistoryItem
	for rows.Next() {
		var item model.SymptomChatHistoryItem
		if err := rows.Scan(
			&item.ID,
			&item.Symptoms,
			&item.OptionalContext,
			&item.SuggestedSpecialty,
			&item.PreliminaryNotes,
			&item.Disclaimer,
			&item.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

// Ping checks database connectivity.
func (r *ChatRepository) Ping(ctx context.Context) error {
	return r.db.PingContext(ctx)
}
