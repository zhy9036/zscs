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

	// 4. Seed a demo project with a sample conversion chat.
	var projectID string
	err = pool.QueryRow(ctx,
		`INSERT INTO projects (user_id, title)
		 VALUES ($1, 'FortiGate → Zscaler Internet Access Demo')
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
		{message.RoleUser, "Convert this FortiGate policy to Zscaler Internet Access:\n\nconfig firewall policy\n edit 10\n  set srcintf \"port1\"\n  set dstintf \"port2\"\n  set srcaddr \"LAN_10.0.0.0/24\"\n  set dstaddr \"all\"\n  set action accept\n  set schedule \"always\"\n  set service \"HTTP\" \"HTTPS\" \"DNS\"\n  set logtraffic all\n next\nend"},
		{message.RoleAssistant, "Here's the equivalent Zscaler Internet Access configuration:\n\n**1. Source IP Anchoring / Client Connector:**\n- Enroll the 10.0.0.0/24 subnet via Zscaler Client Connector (ZCC) or GRE/IPsec tunnel for the FortiGate LAN.\n\n**2. Firewall Rules (Zscaler Firewall):**\n- Rule name: `FG-Policy-10-Allow-Web-DNS`\n- Source: `10.0.0.0/24` (LAN_10.0.0.0/24)\n- Destination: `Any`\n- Services: `HTTP (80)`, `HTTPS (443)`, `DNS (53)`\n- Action: `Allow`\n- Log: `Enabled` (all sessions)\n\n**3. URL Filtering Policy:**\n- Apply the default `Corporate` URL category profile to the same source group.\n\n**4. DNS Control:**\n- Ensure DNS queries from 10.0.0.0/24 are forwarded to Zscaler DNS Resolver for inspection.\n\nNotes:\n- FortiGate `srcintf`/`dstintf` concepts don't apply in ZIA — traffic is steered by tunnel/ZCC enrollment, not interfaces.\n- `set logtraffic all` maps to enabling session logging on the firewall rule.\n- Replace `dstaddr all` with explicit destination groups if you want stricter egress control."},
		{message.RoleUser, "What about the NAT rules I had on the FortiGate?"},
		{message.RoleAssistant, "Zscaler ZIA is a cloud-delivered secure web gateway — it inspects traffic steered to it, it doesn't perform source NAT for outbound internet access. Your existing edge router or firewall handling NAT for the LAN remains responsible for that.\n\nIf the FortiGate was doing U-turn NAT or hairpinning, keep that on the on-premises device. ZIA only inspects the steered flows; it doesn't translate addresses."},
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
