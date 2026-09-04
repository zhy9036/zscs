package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/zscaler/migration-platform/backend/internal/auth"
	"github.com/zscaler/migration-platform/backend/internal/database"
	"github.com/zscaler/migration-platform/backend/internal/user"
)

func main() {
	username := os.Getenv("SEED_USERNAME")
	password := os.Getenv("SEED_PASSWORD")
	dbURL := os.Getenv("DATABASE_URL")

	if username == "" || password == "" {
		log.Fatal("SEED_USERNAME and SEED_PASSWORD are required")
	}
	if dbURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	ctx := context.Background()
	pool, err := database.Connect(ctx, dbURL)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	defer pool.Close()

	repo := user.NewRepository(pool)
	if _, err := repo.GetByUsername(ctx, username); err == nil {
		fmt.Printf("user %q already exists\n", username)
		return
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		log.Fatalf("hash password: %v", err)
	}

	if _, err := repo.Create(ctx, username, hash); err != nil {
		log.Fatalf("create user: %v", err)
	}
	fmt.Printf("seeded user %q\n", username)
}
