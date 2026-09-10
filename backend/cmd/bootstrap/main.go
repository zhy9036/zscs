package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zscaler/migration-platform/backend/internal/auth"
	"github.com/zscaler/migration-platform/backend/internal/message"
)

const (
	pgURL    = "postgres://zscaler:zscaler@localhost:5432/postgres?sslmode=disable"
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
	} else if err == pgx.ErrNoRows {
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
	} else {
		log.Fatalf("lookup user: %v", err)
	}

	fmt.Println("\nLogin credentials:")
	fmt.Printf("  username: %s\n  password: %s\n", demoUser, demoPass)

	// 4. Seed a demo project with a sample conversion chat.
	var projectID string
	err = pool.QueryRow(ctx,
		`INSERT INTO projects (user_id, title)
		 VALUES ($1, 'Zscaler → FortiSASE Migration Demo')
		 ON CONFLICT DO NOTHING
		 RETURNING id`,
		userID,
	).Scan(&projectID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			fmt.Println("demo project already exists, skipping messages")
			_ = time.Now
			_ = os.Args
			return
		}
		log.Fatalf("create demo project: %v", err)
	}
	fmt.Printf("seeded demo project (id=%s)\n", projectID)

	demoMessages := []struct {
		role    string
		content string
	}{
		{message.RoleUser, "Convert this Zscaler Internet Access firewall rule to FortiSASE:\n\nZscaler ZIA Firewall Rule:\n- Name: Allow-Web-DNS-LAN\n- Source IP: 10.0.0.0/24\n- Destination: Any\n- Services: HTTP, HTTPS, DNS\n- Action: Allow\n- Log: Enabled (all sessions)\n- URL Filtering: Corporate profile applied\n- DNS Control: Forwarded to Zscaler DNS Resolver"},
		{message.RoleAssistant, "Here's the equivalent FortiSASE configuration:\n\n**1. Onboarding / Connectivity:**\n- Enroll the 10.0.0.0/24 site via FortiClient or an IPsec tunnel to the nearest FortiSASE PoP.\n- Traffic steering is handled by the FortiSASE data center, replacing Zscaler ZCC/GRE tunneling.\n\n**2. Firewall Policy (FortiSASE REST API — POST /api/v2/cmdb/firewall/policy):**\n```json\n{\n  \"name\": \"Allow-Web-DNS-LAN\",\n  \"srcintf\": [{\"name\": \"ssl-root\"}],\n  \"dstintf\": [{\"name\": \"virtual-wan\"}],\n  \"srcaddr\": [{\"name\": \"LAN_10.0.0.0_24\"}],\n  \"dstaddr\": [{\"name\": \"all\"}],\n  \"action\": \"accept\",\n  \"schedule\": \"always\",\n  \"service\": [\n    {\"name\": \"HTTP\"},\n    {\"name\": \"HTTPS\"},\n    {\"name\": \"DNS\"}\n  ],\n  \"logtraffic\": \"all\",\n  \"utm-status\": \"enable\",\n  \"webfilter-profile\": \"default\",\n  \"dnsfilter-profile\": \"default\",\n  \"ips-sensor\": \"default\"\n}\n```\n\n**3. Web Filtering:**\n- The Zscaler `Corporate` URL category profile maps to a FortiSASE web filter profile applied via the `webfilter-profile` field in the policy JSON.\n\n**4. DNS Control:**\n- Zscaler DNS Resolver inspection maps to the FortiSASE DNS Filter profile (`dnsfilter-profile` field).\n\nNotes:\n- Interface/address/service fields are arrays of `{\"name\": \"...\"}` objects, not plain strings — this is the most common API mistake.\n- Zscaler's cloud-native firewall rule (source IP / service / action) maps directly to a FortiSASE policy, but you gain UTM profiles (IPS, web filter, DNS filter) on the same policy via `utm-status: enable`.\n- Zscaler's `Log: Enabled` maps to `logtraffic: all`.\n- Zscaler's `Destination: Any` becomes `dstaddr: [{\"name\": \"all\"}]` — tighten this with explicit address objects if you want stricter egress control."},
		{message.RoleUser, "How do I migrate the Zscaler URL filtering categories and custom block lists?"},
		{message.RoleAssistant, "Zscaler and FortiSASE use different URL category taxonomies, so there's no 1:1 mapping. Approach:\n\n1. Export your Zscaler custom URL categories and block/allow lists.\n2. In FortiSASE, recreate them under **Security Profiles → Web Filter**:\n   - Map each Zscaler category to the closest FortiGuard category.\n   - Recreate custom block lists as FortiSASE custom URL categories.\n3. Attach the web filter profile to the firewall policy via `set webfilter-profile`.\n\nFortiGuard categories are generally more granular than Zscaler's, so expect to split some Zscaler categories into multiple FortiGuard categories. Review the mapping carefully — a single Zscaler category often maps to 2-3 FortiGuard categories."},
	}

	for _, m := range demoMessages {
		_, err = pool.Exec(ctx,
			`INSERT INTO messages (project_id, role, content) VALUES ($1, $2, $3)`,
			projectID, m.role, m.content,
		)
		if err != nil {
			log.Fatalf("seed demo message (%s): %v", m.role, err)
		}
	}
	fmt.Printf("seeded %d demo messages\n", len(demoMessages))

	_ = time.Now
	_ = os.Args
}
