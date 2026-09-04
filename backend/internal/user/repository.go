package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("user not found")

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, username, passwordHash string) (User, error) {
	const q = `
		INSERT INTO users (username, password_hash)
		VALUES ($1, $2)
		RETURNING id, username, password_hash, created_at, updated_at
	`
	var u User
	err := r.pool.QueryRow(ctx, q, username, passwordHash).
		Scan(&u.ID, &u.Username, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return User{}, fmt.Errorf("create user: %w", err)
	}
	return u, nil
}

func (r *Repository) GetByUsername(ctx context.Context, username string) (User, error) {
	const q = `
		SELECT id, username, password_hash, created_at, updated_at
		FROM users WHERE username = $1
	`
	var u User
	err := r.pool.QueryRow(ctx, q, username).
		Scan(&u.ID, &u.Username, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, ErrNotFound
		}
		return User{}, fmt.Errorf("get user by username: %w", err)
	}
	return u, nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (User, error) {
	const q = `
		SELECT id, username, password_hash, created_at, updated_at
		FROM users WHERE id = $1
	`
	var u User
	err := r.pool.QueryRow(ctx, q, id).
		Scan(&u.ID, &u.Username, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, ErrNotFound
		}
		return User{}, fmt.Errorf("get user by id: %w", err)
	}
	return u, nil
}
