package file

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("file not found")

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, f File) (File, error) {
	const q = `
		INSERT INTO project_files (project_id, filename, content_type, size, storage_key)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, project_id, filename, content_type, size, storage_key, created_at
	`
	var out File
	err := r.pool.QueryRow(ctx, q, f.ProjectID, f.Filename, f.ContentType, f.Size, f.StorageKey).
		Scan(&out.ID, &out.ProjectID, &out.Filename, &out.ContentType, &out.Size, &out.StorageKey, &out.CreatedAt)
	if err != nil {
		return File{}, fmt.Errorf("create file: %w", err)
	}
	return out, nil
}

func (r *Repository) GetByID(ctx context.Context, projectID, id string) (File, error) {
	const q = `
		SELECT id, project_id, filename, content_type, size, storage_key, created_at
		FROM project_files WHERE id = $1 AND project_id = $2
	`
	var f File
	err := r.pool.QueryRow(ctx, q, id, projectID).
		Scan(&f.ID, &f.ProjectID, &f.Filename, &f.ContentType, &f.Size, &f.StorageKey, &f.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return File{}, ErrNotFound
		}
		return File{}, fmt.Errorf("get file: %w", err)
	}
	return f, nil
}

func (r *Repository) ListByProject(ctx context.Context, projectID string) ([]File, error) {
	const q = `
		SELECT id, project_id, filename, content_type, size, storage_key, created_at
		FROM project_files WHERE project_id = $1 ORDER BY created_at ASC
	`
	rows, err := r.pool.Query(ctx, q, projectID)
	if err != nil {
		return nil, fmt.Errorf("list files: %w", err)
	}
	defer rows.Close()

	var items []File
	for rows.Next() {
		var f File
		if err := rows.Scan(&f.ID, &f.ProjectID, &f.Filename, &f.ContentType, &f.Size, &f.StorageKey, &f.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan file: %w", err)
		}
		items = append(items, f)
	}
	return items, rows.Err()
}

func (r *Repository) Delete(ctx context.Context, projectID, id string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM project_files WHERE id = $1 AND project_id = $2`, id, projectID)
	if err != nil {
		return fmt.Errorf("delete file: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
