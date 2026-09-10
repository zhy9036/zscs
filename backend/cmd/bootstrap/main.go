package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zscaler/migration-platform/backend/internal/auth"
)

const (
	pgURL    = "postgres://postgres@127.0.0.1:5432/postgres?sslmode=disable"
	dbName   = "zscaler_migration"
	dbURL    = "postgres://zscaler:zscaler@localhost:5432/zscaler_migration?sslmode=disable"
	demoUser = "demo"
	demoPass = "demo123"
)

func main() {
	ctx := context.Background()

	// 1. Connect to the default postgres DB and create the app DB if missing.
	admin, err := pgxpool.New(ctx, pgURL)
	if err != nil {
		log.Fatalf("connect postgres: %v", err)
	}
	defer admin.Close()

	var exists bool
	err = admin.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)`, dbName,
	).Scan(&exists)
	if err != nil {
		log.Fatalf("check db existence: %v", err)
	}
	if !exists {
		_, err = admin.Exec(ctx, fmt.Sprintf(`CREATE DATABASE %s`, dbName))
		if err != nil {
			log.Fatalf("create database: %v", err)
		}
		fmt.Printf("created database %q\n", dbName)
	} else {
		fmt.Printf("database %q already exists\n", dbName)
	}
	admin.Close()

	// 2. Connect to the app DB and run migrations.
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("connect app db: %v", err)
	}
	defer pool.Close()

	migrations := []string{
		`CREATE EXTENSION IF NOT EXISTS "pgcrypto"`,
		`CREATE TABLE IF NOT EXISTS users (
			id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			username      TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
			updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
		)`,
		`CREATE TABLE IF NOT EXISTS projects (
			id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			title      TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_projects_user_updated ON projects (user_id, updated_at DESC)`,
		`CREATE TABLE IF NOT EXISTS project_files (
			id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			project_id   UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
			filename     TEXT NOT NULL,
			content_type TEXT NOT NULL,
			size         BIGINT NOT NULL,
			storage_key  TEXT NOT NULL,
			created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_project_files_project ON project_files (project_id)`,
		`CREATE TABLE IF NOT EXISTS messages (
			id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			project_id   UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
			role         TEXT NOT NULL CHECK (role IN ('user', 'assistant', 'system')),
			content      TEXT NOT NULL,
			agent_run_id UUID,
			metadata     JSONB,
			created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_messages_project_created ON messages (project_id, created_at)`,
	}
	for i, q := range migrations {
		if _, err := pool.Exec(ctx, q); err != nil {
			log.Fatalf("migration step %d failed: %v", i+1, err)
		}
	}
	fmt.Println("migrations applied")

	// 3. Seed the demo user (idempotent).
	var userID string
	err = pool.QueryRow(ctx, `SELECT id FROM users WHERE username = $1`, demoUser).Scan(&userID)
	if err == nil {
		fmt.Printf("user %q already exists (id=%s)\n", demoUser, userID)
		return
	}
	if err != pgx.ErrNoRows {
		log.Fatalf("lookup user: %v", err)
	}

	hash, err := auth.HashPassword(demoPass)
	if err != nil {
		log.Fatalf("hash password: %v", err)
	}

	err = pool.QueryRow(ctx,
		`INSERT INTO users (username, password_hash) VALUES ($1, $2) RETURNING id`,
		demoUser, hash,
	).Scan(&userID)
	if err != nil {
		log.Fatalf("create demo user: %v", err)
	}

	fmt.Printf("seeded user %q (id=%s)\n", demoUser, userID)
	fmt.Println("\nLogin credentials:")
	fmt.Printf("  username: %s\n  password: %s\n", demoUser, demoPass)

	_ = time.Now
	_ = os.Args
}
