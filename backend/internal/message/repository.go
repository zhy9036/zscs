package message

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = fmt.Errorf("message not found")

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, m Message) (Message, error) {
	const q = `
		INSERT INTO messages (project_id, role, content)
		VALUES ($1, $2, $3)
		RETURNING id, project_id, role, content, created_at
	`
	var out Message
	err := r.pool.QueryRow(ctx, q, m.ProjectID, m.Role, m.Content).
		Scan(&out.ID, &out.ProjectID, &out.Role, &out.Content, &out.CreatedAt)
	if err != nil {
		return Message{}, fmt.Errorf("create message: %w", err)
	}
	return out, nil
}

func (r *Repository) ListByProject(ctx context.Context, projectID string) ([]Message, error) {
	const q = `
		SELECT id, project_id, role, content, created_at
		FROM messages WHERE project_id = $1
		ORDER BY created_at ASC
	`
	rows, err := r.pool.Query(ctx, q, projectID)
	if err != nil {
		return nil, fmt.Errorf("list messages: %w", err)
	}
	defer rows.Close()

	var items []Message
	for rows.Next() {
		var m Message
		if err := rows.Scan(&m.ID, &m.ProjectID, &m.Role, &m.Content, &m.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan message: %w", err)
		}
		items = append(items, m)
	}
	return items, rows.Err()
}
