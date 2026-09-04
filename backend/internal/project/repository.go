package project

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("project not found")

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, p Project) (Project, error) {
	const q = `
		INSERT INTO projects (user_id, title)
		VALUES ($1, $2)
		RETURNING id, user_id, title, created_at, updated_at
	`
	var out Project
	err := r.pool.QueryRow(ctx, q, p.UserID, p.Title).
		Scan(&out.ID, &out.UserID, &out.Title, &out.CreatedAt, &out.UpdatedAt)
	if err != nil {
		return Project{}, fmt.Errorf("create project: %w", err)
	}
	return out, nil
}

func (r *Repository) GetByID(ctx context.Context, userID, id string) (Project, error) {
	const q = `
		SELECT id, user_id, title, created_at, updated_at
		FROM projects WHERE id = $1 AND user_id = $2
	`
	var p Project
	err := r.pool.QueryRow(ctx, q, id, userID).
		Scan(&p.ID, &p.UserID, &p.Title, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Project{}, ErrNotFound
		}
		return Project{}, fmt.Errorf("get project: %w", err)
	}
	return p, nil
}

func (r *Repository) List(ctx context.Context, userID string) ([]Project, error) {
	const q = `
		SELECT id, user_id, title, created_at, updated_at
		FROM projects WHERE user_id = $1
		ORDER BY updated_at DESC
	`
	rows, err := r.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}
	defer rows.Close()

	var items []Project
	for rows.Next() {
		var p Project
		if err := rows.Scan(&p.ID, &p.UserID, &p.Title, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan project: %w", err)
		}
		items = append(items, p)
	}
	return items, rows.Err()
}

func (r *Repository) Update(ctx context.Context, userID, id string, update ProjectUpdate) (Project, error) {
	const q = `
		UPDATE projects SET title = $3, updated_at = now()
		WHERE id = $1 AND user_id = $2
		RETURNING id, user_id, title, created_at, updated_at
	`
	var p Project
	err := r.pool.QueryRow(ctx, q, id, userID, update.Title).
		Scan(&p.ID, &p.UserID, &p.Title, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Project{}, ErrNotFound
		}
		return Project{}, fmt.Errorf("update project: %w", err)
	}
	return p, nil
}

func (r *Repository) Delete(ctx context.Context, userID, id string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM projects WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return fmt.Errorf("delete project: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) Touch(ctx context.Context, userID, id string) error {
	_, err := r.pool.Exec(ctx, `UPDATE projects SET updated_at = now() WHERE id = $1 AND user_id = $2`, id, userID)
	return err
}
