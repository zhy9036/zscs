# Zscaler Migration Platform — Architecture Summary

## 1. What we're building

A ChatGPT-style web application that lets customers upload a Zscaler configuration and chat with a migration agent to analyze and convert it. This phase delivers the **frontend + backend foundation** with a clean plug-in point for the real migration agent, which is implemented later by the agent team.

**Scope of this phase:** auth, projects, file upload, chat UI, message persistence, and a mock agent so the full workflow is demoable end-to-end.
**Out of scope:** the real LLM agent, config parsing/conversion, RAG, SSO, multi-tenant RBAC.

---

## 2. High-level architecture

```text
        Browser (React + TS + Vite)
                  |  REST /api/v1
                  v
        Go Backend (Chi router)
        ┌─────────────────────┐
        │ Auth  Project       │
        │ File  Message      │
        │ AgentService iface │
        └──────────┬──────────┘
                   │
        ┌──────────┴──────────┐
        v                     v
   PostgreSQL            File Storage
  (users, projects,     (local now,
   files, messages)      S3-ready later)

                   │ future
                   v
        Agent Runtime (LLM + tools + RAG)
```

The backend is a **stable API boundary** — the frontend never knows which LLM or tools back the agent.

---

## 3. Tech stack

| Layer    | Choice                                   |
|----------|------------------------------------------|
| Frontend | React, TypeScript, Vite, TanStack Query, Tailwind, shadcn/ui |
| Backend  | Go 1.25, Chi, pgx, JWT, zerolog          |
| Database | PostgreSQL 16                            |
| Storage  | Local filesystem (abstracted behind `FileStorage` interface, S3-swappable) |

---

## 4. Core data model

```text
User 1──N Project 1──N ProjectFile
                    └──N Message (role: user | assistant | system)
```

Ownership is enforced server-side on every request — a user can never read or mutate another user's projects, files, or messages.

---

## 5. API surface (versioned under `/api/v1`)

| Method | Endpoint                              | Purpose                       |
|--------|---------------------------------------|-------------------------------|
| POST   | /auth/login                           | Username/password → JWT       |
| GET    | /auth/me                              | Restore session on refresh    |
| GET    | /projects                             | List user's projects          |
| POST   | /projects                             | Create project + upload config (multipart) |
| GET/PATCH/DELETE | /projects/{id}             | Read / rename / delete        |
| GET    | /projects/{id}/messages               | Message history               |
| POST   | /projects/{id}/messages               | Send message → mock reply     |
| GET/DELETE | /projects/{id}/files/{fileId}     | Download / remove config file |
| GET    | /health, /ready                       | Liveness / readiness          |

Future: `POST /projects/{id}/messages/stream` (SSE) for streaming agent responses.

---

## 6. Security

- Passwords hashed with bcrypt; never logged or returned.
- JWT access tokens; user identity taken from token, never from client-supplied IDs.
- Generic "invalid username or password" errors to prevent account enumeration.
- Per-user ownership checks on every project/file/message operation.
- Uploaded files stored under generated UUID keys (no user-controlled paths), with size limits enforced server-side.
- Secrets and config via environment variables only.

---

## 7. Agent integration contract

The agent plugs in behind a single interface, so the HTTP layer is untouched when the real agent ships:

```go
type AgentService interface {
    Run(ctx context.Context, req AgentRequest) (<-chan AgentEvent, error)
}
```

Event types (`message`, `thinking`, `tool_start`, `tool_result`, `file`, `error`, `done`) are already designed for streaming. Today this is backed by a `Mock` that returns a placeholder reply — enough to exercise the full chat loop and demo the product.

---

## 8. Current status

- ✅ Auth (login, JWT, `/me`)
- ✅ Project CRUD with ownership enforcement
- ✅ Config file upload + download/delete
- ✅ Message persistence + mock agent round-trip
- ✅ Frontend: login, sidebar, new-project drop zone, chat UI
- ⏳ Streaming (SSE) endpoint scaffolded, not yet implemented
- ⏳ Real migration agent — next phase

A demo account (`demo` / `demo123`) is seeded locally; the full login → upload → chat flow is runnable today.

---

## 9. Roadmap

1. **Now** — richer mock agent for convincing POC demos.
2. **Next phase** — implement `AgentService` against the real LLM + Zscaler parser + migration tools; enable SSE streaming.
3. **Later** — S3 file storage, SSO, soft-delete, multi-user RBAC.

The architecture is deliberately shaped so that step 2 replaces one component without touching the frontend or the API contracts.
