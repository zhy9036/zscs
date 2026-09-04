# Zscaler Migration Agent — Frontend & Backend Design

## 1. Overview

### 1.1 Purpose

Build the frontend and backend foundation for a ChatGPT-like Zscaler migration application.

The system will provide:

1. Username/password authentication.
2. ChatGPT-style project management.
3. A left sidebar containing the user's projects.
4. A project creation flow that starts by uploading/dragging a Zscaler configuration file into a designated drop area.
5. A chat interface for interacting with the migration system.
6. Backend APIs that abstract authentication, projects, messages, configuration files, and future agent execution.
7. A clean extension point for the future migration-agent implementation.

The actual migration agent design, LLM orchestration, tools, prompts, reasoning, and migration logic are **out of scope for this phase**.

The backend should be implemented in **Go**.

---

# 2. Goals

## 2.1 MVP Goals

### Frontend

* Login page.
* Authentication state management.
* ChatGPT-like application layout.
* Left project sidebar.
* Create-new-project flow.
* Zscaler configuration drag-and-drop upload.
* Project page.
* Chat message display.
* Message input.
* Loading/streaming-ready UI.
* Project rename/delete support.
* Error handling.
* Logout.

### Backend

* Username/password authentication.
* Secure password storage.
* Session/token management.
* User management.
* Project CRUD.
* Configuration file upload.
* Message persistence.
* APIs designed for future agent integration.
* Database persistence.
* Authorization so users can only access their own projects/files.
* Health/readiness endpoints.

---

# 3. Non-Goals

The following are explicitly out of scope for this phase:

* Migration agent implementation.
* LLM provider integration.
* Prompt engineering.
* Agent planning/reasoning.
* Agent tools.
* Zscaler configuration parsing.
* Zscaler configuration conversion.
* Migration validation.
* Automatic migration execution.
* RAG/vector database.
* Agent memory implementation.
* Multi-agent orchestration.
* Enterprise SSO/SAML/OIDC.
* RBAC beyond basic user ownership.
* Production-grade object storage integration unless required by deployment.

The architecture should, however, make these features easy to add later.

---

# 4. High-Level Architecture

```text
                         Browser
                            |
                            | HTTPS
                            v
                  +---------------------+
                  |      Frontend       |
                  |                     |
                  | React / TypeScript  |
                  +----------+----------+
                             |
                             | REST API
                             v
                  +---------------------+
                  |      Go Backend     |
                  |                     |
                  |  Auth API           |
                  |  Project API   |
                  |  Message API        |
                  |  File API           |
                  |  Agent API           |
                  +----------+----------+
                             |
              +--------------+--------------+
              |                             |
              v                             v
      +---------------+             +---------------+
      |   PostgreSQL  |             | File Storage  |
      |               |             |               |
      | Users         |             | Zscaler cfg   |
      | Projects |             |               |
      | Messages      |             | Local / S3    |
      | Files         |             |               |
      +---------------+             +---------------+

                             |
                             | Future
                             v
                  +---------------------+
                  |   Agent Runtime     |
                  |                     |
                  | Migration Agent     |
                  | LLM / Tools / etc.  |
                  +---------------------+
```

The backend should act as a **stable API boundary** between the frontend and the future agent runtime.

---

# 5. Recommended Technology Stack

## Frontend

* React
* TypeScript
* Vite
* React Router
* TanStack Query
* Tailwind CSS
* A component library such as shadcn/ui
* Native browser drag-and-drop API or a small dropzone library

## Backend

* Go
* Go 1.24+
* HTTP server using `net/http`
* Chi router
* PostgreSQL
* `pgx` PostgreSQL driver
* SQL migrations
* JWT access tokens
* bcrypt or Argon2id password hashing
* Structured logging

Recommended backend dependency structure:

```text
cmd/
internal/
pkg/
migrations/
```

Avoid introducing a large Go web framework unless there is a strong project-wide reason.

---

# 6. Authentication

## 6.1 Login

The initial authentication mechanism is:

```text
username + password
```

Users submit:

```json
{
  "username": "user1",
  "password": "password"
}
```

The backend validates the credentials and returns an access token.

Example:

```json
{
  "access_token": "<token>",
  "token_type": "Bearer",
  "expires_in": 3600,
  "user": {
    "id": "uuid",
    "username": "user1"
  }
}
```

The frontend sends:

```http
Authorization: Bearer <access_token>
```

for authenticated requests.

---

# 7. Authentication Security

Passwords must never be stored in plaintext.

Recommended:

```text
Argon2id
```

or:

```text
bcrypt
```

Store only the password hash.

The backend must:

* validate password length
* hash passwords before storage
* never return password hashes
* never log passwords
* return generic authentication errors

For example:

```text
Invalid username or password
```

rather than:

```text
Username does not exist
```

to avoid account enumeration.

---

# 8. Token Design

Use JWT for the initial implementation.

JWT claims:

```json
{
  "sub": "user-uuid",
  "username": "user1",
  "iat": 1750000000,
  "exp": 1750003600
}
```

The backend authentication middleware should:

1. Extract the Authorization header.
2. Validate the Bearer token.
3. Validate signature.
4. Validate expiration.
5. Extract user ID.
6. Put the authenticated user ID into the request context.

Example:

```go
type AuthContext struct {
    UserID string
}
```

All project/file/message APIs should derive ownership from this authenticated user ID rather than trusting a user ID supplied by the frontend.

---

# 9. Database Model

Use PostgreSQL.

## 9.1 Users

```sql
users
-----
id
username
password_hash
created_at
updated_at
```

Constraints:

```text
id          UUID PRIMARY KEY
username    UNIQUE NOT NULL
```

---

## 9.2 Projects

```sql
projects
-------------
id
user_id
title
created_at
updated_at
```

Relationship:

```text
User 1 ---- N Projects
```

A project belongs to exactly one user.

---

## 9.3 Configuration Files

```sql
project_files
-------------------
id
project_id
filename
content_type
size
storage_key
created_at
```

A project may have one or more uploaded configuration files.

For MVP, normally the first configuration file starts the project.

---

## 9.4 Messages

```sql
messages
--------
id
project_id
role
content
created_at
```

Role:

```text
user
assistant
system
```

The model should not assume that an assistant message necessarily came from the future agent. The backend should allow future metadata to identify the source.

Potential future fields:

```text
agent_run_id
metadata
```

These do not have to be implemented in the initial MVP unless useful.

---

# 10. Entity Relationships

```text
User
 |
 +---- Project
          |
          +---- ProjectFile
          |
          +---- Message
```

Ownership hierarchy:

```text
User
  |
  +-- Project
        |
        +-- File
        |
        +-- Message
```

Every API operation must verify this ownership chain.

Example:

```text
GET /api/projects/{projectID}
```

must not return a project belonging to another user.

---

# 11. REST API

API prefix:

```text
/api/v1
```

---

# 12. Authentication API

## POST /api/v1/auth/login

Request:

```json
{
  "username": "user1",
  "password": "password"
}
```

Response:

```json
{
  "access_token": "...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "user": {
    "id": "uuid",
    "username": "user1"
  }
}
```

---

## GET /api/v1/auth/me

Requires authentication.

Response:

```json
{
  "id": "uuid",
  "username": "user1"
}
```

Frontend uses this endpoint to restore authentication state after page refresh.

---

# 13. Project API

## GET /api/v1/projects

Return projects belonging to the authenticated user.

Response:

```json
{
  "items": [
    {
      "id": "project-uuid",
      "title": "Zscaler migration - branch office",
      "created_at": "2026-08-27T17:00:00Z",
      "updated_at": "2026-08-27T17:30:00Z"
    }
  ]
}
```

Sort by:

```text
updated_at DESC
```

This produces ChatGPT-style recent projects at the top.

---

## POST /api/v1/projects

This endpoint should normally be called after a configuration file is selected.

Prefer a multipart request:

```http
POST /api/v1/projects
Content-Type: multipart/form-data
```

Fields:

```text
file=<zscaler-config>
title=<optional>
```

The backend:

1. Authenticates the user.
2. Validates the file.
3. Creates the project.
4. Stores the configuration file.
5. Creates an initial project state.
6. Returns the project.

Response:

```json
{
  "id": "project-uuid",
  "title": "zscaler-config.conf",
  "created_at": "2026-08-27T17:00:00Z",
  "updated_at": "2026-08-27T17:00:00Z",
  "files": [
    {
      "id": "file-uuid",
      "filename": "zscaler-config.conf",
      "content_type": "text/plain",
      "size": 12345
    }
  ]
}
```

---

## GET /api/v1/projects/{projectID}

Return:

* project metadata
* uploaded files
* messages

Example:

```json
{
  "id": "project-uuid",
  "title": "Zscaler migration",
  "created_at": "...",
  "updated_at": "...",
  "files": [],
  "messages": []
}
```

---

## PATCH /api/v1/projects/{projectID}

Used for renaming.

Request:

```json
{
  "title": "HQ Zscaler Migration"
}
```

---

## DELETE /api/v1/projects/{projectID}

Delete the project and associated resources.

For MVP, hard deletion is acceptable.

Future implementation may use soft deletion.

---

# 14. Message API

## GET /api/v1/projects/{projectID}/messages

Return messages ordered by creation time.

Response:

```json
{
  "items": [
    {
      "id": "message-1",
      "role": "user",
      "content": "Analyze this Zscaler configuration.",
      "created_at": "..."
    },
    {
      "id": "message-2",
      "role": "assistant",
      "content": "I will analyze the configuration.",
      "created_at": "..."
    }
  ]
}
```

---

## POST /api/v1/projects/{projectID}/messages

Request:

```json
{
  "content": "What migration issues do you see?"
}
```

For the MVP, this endpoint can persist the user message and return a placeholder response.

Example:

```json
{
  "user_message": {
    "id": "message-1",
    "role": "user",
    "content": "What migration issues do you see?",
    "created_at": "..."
  },
  "assistant_message": null,
  "status": "accepted"
}
```

The actual agent execution should be implemented later behind a service interface.

---

# 15. Future Agent Integration

The backend should not directly couple HTTP handlers to agent implementation.

Define an abstraction such as:

```go
type AgentService interface {
    Run(ctx context.Context, request AgentRequest) (<-chan AgentEvent, error)
}
```

Example request:

```go
type AgentRequest struct {
    UserID         string
    ProjectID string
    Message        string
}
```

Example future events:

```go
type AgentEvent struct {
    Type string
    Data any
}
```

Potential event types:

```text
message
thinking
tool_start
tool_result
file
error
done
```

The HTTP layer should only know about `AgentService`.

It should not know:

* which LLM is being used
* which tools exist
* how the agent reasons
* how Zscaler migration works
* which model provider is used

This allows the agent team to implement the agent independently.

---

# 16. Streaming Architecture

The UI should be designed to support streaming responses even if the first MVP does not implement streaming.

Recommended future API:

```text
POST /api/v1/projects/{id}/messages/stream
```

Use Server-Sent Events (SSE).

Example:

```text
event: message
data: {"content":"I"}

event: message
data: {"content":" will"}

event: message
data: {"content":" analyze"}

event: done
data: {}
```

Frontend behavior:

```text
User sends message
       |
       v
Create user message
       |
       v
Open SSE connection
       |
       v
Receive agent events
       |
       v
Update assistant message incrementally
```

WebSocket is not necessary for the initial implementation unless there is a requirement for bidirectional real-time events.

SSE is simpler and maps well to agent-generated output.

---

# 17. Configuration Upload UX

The main "New Project" flow is intentionally different from a normal ChatGPT project.

## Initial Screen

Display a large drop zone:

```text
+------------------------------------------------+
|                                                |
|          Start a Zscaler Migration             |
|                                                |
|       Drag and drop your Zscaler config        |
|                    here                        |
|                                                |
|                  or                            |
|                                                |
|              [ Browse Files ]                 |
|                                                |
|     Supported configuration files              |
|                                                |
+------------------------------------------------+
```

The user can:

* drag a file into the area
* click Browse Files
* select a configuration file

---

# 18. Upload Flow

```text
User
 |
 | Drag config
 v
Frontend Drop Zone
 |
 | Validate basic file properties
 v
POST /api/v1/projects
 |
 v
Backend
 |
 +-- authenticate user
 |
 +-- validate file
 |
 +-- create project
 |
 +-- store file
 |
 v
Return project ID
 |
 v
Frontend navigates to:
 /chat/{projectID}
```

---

# 19. File Validation

Frontend validation is for UX only.

Backend validation is authoritative.

At minimum validate:

* file exists
* filename exists
* size limit
* content type if available
* allowed extension if an extension policy is defined

Example initial limit:

```text
Maximum file size: 50 MB
```

Make this configurable:

```text
MAX_UPLOAD_SIZE
```

Do not hard-code this limit throughout the codebase.

---

# 20. File Storage Abstraction

Do not make the application depend directly on local filesystem APIs.

Define:

```go
type FileStorage interface {
    Put(ctx context.Context, key string, r io.Reader) error
    Get(ctx context.Context, key string) (io.ReadCloser, error)
    Delete(ctx context.Context, key string) error
}
```

MVP implementation:

```text
LocalFileStorage
```

Future implementation:

```text
S3FileStorage
```

This allows deployment to switch from local storage to object storage without changing project logic.

---

# 21. Frontend Application Structure

Recommended structure:

```text
src/
├── app/
│   ├── App.tsx
│   ├── routes.tsx
│   └── providers.tsx
│
├── components/
│   ├── layout/
│   │   ├── AppLayout.tsx
│   │   ├── Sidebar.tsx
│   │   └── Header.tsx
│   │
│   ├── chat/
│   │   ├── ChatWindow.tsx
│   │   ├── MessageList.tsx
│   │   ├── MessageBubble.tsx
│   │   └── MessageInput.tsx
│   │
│   ├── upload/
│   │   └── ConfigDropZone.tsx
│   │
│   └── common/
│
├── pages/
│   ├── LoginPage.tsx
│   ├── HomePage.tsx
│   ├── NewProjectPage.tsx
│   └── ProjectPage.tsx
│
├── api/
│   ├── client.ts
│   ├── auth.ts
│   ├── projects.ts
│   └── messages.ts
│
├── hooks/
│   ├── useAuth.ts
│   ├── useProjects.ts
│   └── useMessages.ts
│
├── stores/
│   └── authStore.ts
│
├── types/
│   ├── auth.ts
│   ├── project.ts
│   └── message.ts
│
└── main.tsx
```

---

# 22. Frontend Routes

```text
/login

/
/new

/chat/:projectId
```

Behavior:

```text
Unauthenticated
      |
      v
    /login

Authenticated
      |
      +--> /
      |
      +--> /new
      |
      +--> /chat/:projectId
```

---

# 23. Main Layout

The application should visually resemble ChatGPT.

```text
+----------------+------------------------------------------+
|                |                                          |
| Projects  |                                          |
|                |          Project Header             |
| + New Chat     |                                          |
|                |------------------------------------------|
| Project 1 |                                          |
| Project 2 |          Message History                 |
| Project 3 |                                          |
|                |                                          |
|                |                                          |
|                |------------------------------------------|
|                |       Message Input / Send               |
|                |                                          |
| User           |                                          |
| Logout         |                                          |
+----------------+------------------------------------------+
```

Sidebar width:

```text
~260px
```

Desktop behavior:

* fixed left sidebar
* scrollable project list
* main content fills remaining viewport

Mobile behavior:

* sidebar becomes a drawer

---

# 24. Project Sidebar

Sidebar contains:

```text
[ + New project ]

Recent
-------------------------
Zscaler HQ Migration
Branch Office Migration
Test Configuration
-------------------------

User
username

[ Logout ]
```

Each project item should support:

* click → open project
* rename
* delete

The sidebar should update after:

* creating a project
* renaming
* deleting
* receiving a new message

---

# 25. Project Title

When a new configuration is uploaded, generate an initial title from the filename.

Example:

```text
zscaler_hq.conf
```

becomes:

```text
zscaler_hq
```

Do not depend on the future agent to generate titles.

Later, the agent may optionally generate a better title.

---

# 26. Chat Interface

The project page contains:

```text
Header
  |
  +-- project title
  |
MessageList
  |
  +-- User message
  +-- Assistant message
  +-- User message
  +-- Assistant message
  |
MessageInput
```

The UI should clearly distinguish:

```text
user
assistant
system
```

Assistant messages should support Markdown rendering.

---

# 27. Message Input

Requirements:

* multiline input
* Enter → send
* Shift+Enter → newline
* Send button
* disabled while request is being submitted
* display upload/agent errors
* automatically scroll to latest message

Future support can add:

* attachments
* agent/tool status
* stop generation
* retry
* regenerate response

These do not need to be implemented initially.

---

# 28. Empty Project State

Immediately after uploading a configuration:

```text
+------------------------------------------------+
|                                                |
|          Zscaler configuration loaded          |
|                                                |
|       Ask a question to get started.           |
|                                                |
|  "Analyze this configuration"                  |
|  "What migration issues should I know about?"  |
|  "Summarize the configuration"                 |
|                                                |
+------------------------------------------------+
```

Suggested prompts are frontend-only initially.

---

# 29. Backend Package Structure

Recommended:

```text
backend/
├── cmd/
│   └── server/
│       └── main.go
│
├── internal/
│   ├── auth/
│   │   ├── service.go
│   │   ├── handler.go
│   │   └── middleware.go
│   │
│   ├── project/
│   │   ├── model.go
│   │   ├── repository.go
│   │   ├── service.go
│   │   └── handler.go
│   │
│   ├── message/
│   │   ├── model.go
│   │   ├── repository.go
│   │   ├── service.go
│   │   └── handler.go
│   │
│   ├── file/
│   │   ├── storage.go
│   │   ├── service.go
│   │   └── handler.go
│   │
│   ├── agent/
│   │   ├── service.go
│   │   └── mock.go
│   │
│   ├── user/
│   │   ├── model.go
│   │   └── repository.go
│   │
│   ├── database/
│   │   └── database.go
│   │
│   └── server/
│       ├── server.go
│       └── routes.go
│
├── migrations/
│
├── go.mod
└── README.md
```

---

# 30. Backend Layering

Use this dependency direction:

```text
HTTP Handler
     |
     v
Service
     |
     v
Repository / Storage
     |
     v
Database / Filesystem
```

For example:

```text
ProjectHandler
       |
       v
ProjectService
       |
       +---- ProjectRepository
       |
       +---- FileStorage
       |
       +---- AgentService (future)
```

Handlers should contain minimal business logic.

---

# 31. Repository Interfaces

Example:

```go
type ProjectRepository interface {
    Create(ctx context.Context, project *Project) error
    GetByID(ctx context.Context, userID, id string) (*Project, error)
    List(ctx context.Context, userID string) ([]Project, error)
    Update(ctx context.Context, userID, id string, update ProjectUpdate) error
    Delete(ctx context.Context, userID, id string) error
}
```

The `userID` should be part of repository methods where ownership matters.

This makes accidental cross-user access harder.

---

# 32. Error Model

All APIs should return consistent errors.

Example:

```json
{
  "error": {
    "code": "project_not_found",
    "message": "Project not found"
  }
}
```

Suggested codes:

```text
invalid_request
unauthorized
forbidden
not_found
conflict
file_too_large
unsupported_file_type
internal_error
```

Do not expose internal database errors to clients.

---

# 33. HTTP Status Codes

Use conventional HTTP status codes.

```text
200 OK
201 Created
204 No Content

400 Bad Request
401 Unauthorized
403 Forbidden
404 Not Found
409 Conflict
413 Payload Too Large

500 Internal Server Error
```

---

# 34. Configuration

Backend configuration should come from environment variables.

Example:

```text
SERVER_PORT=8080

DATABASE_URL=postgres://...

JWT_SECRET=...

JWT_EXPIRATION=1h

MAX_UPLOAD_SIZE=52428800

FILE_STORAGE_PATH=/data/uploads
```

Never hard-code secrets.

---

# 35. CORS

During development, allow the frontend development server.

Example:

```text
http://localhost:5173
```

Production CORS should be explicitly configured.

Do not use:

```text
Access-Control-Allow-Origin: *
```

when authentication credentials are involved.

---

# 36. Database Migrations

Use versioned SQL migrations.

Example:

```text
migrations/
├── 001_create_users.up.sql
├── 001_create_users.down.sql
├── 002_create_projects.up.sql
├── 002_create_projects.down.sql
├── 003_create_files.up.sql
├── 003_create_files.down.sql
├── 004_create_messages.up.sql
└── 004_create_messages.down.sql
```

Migrations should be executable independently of application startup.

---

# 37. API Client Design

Frontend API calls should be centralized.

Example:

```text
api/
├── client.ts
├── auth.ts
├── projects.ts
├── messages.ts
└── files.ts
```

Do not call `fetch()` directly from React components.

Example:

```ts
projectApi.create(file)
projectApi.list()
projectApi.get(id)
projectApi.rename(id, title)
projectApi.delete(id)

messageApi.list(projectId)
messageApi.send(projectId, content)
```

---

# 38. Authentication State

Frontend authentication state should contain:

```ts
type AuthState = {
    user: User | null
    accessToken: string | null
    authenticated: boolean
}
```

On application startup:

```text
Load token
   |
   v
GET /auth/me
   |
   +-- success --> authenticated
   |
   +-- 401 ------> logout
```

---

# 39. Token Storage

For the initial implementation, the frontend can use a token-based authentication mechanism.

Prefer an architecture that can later migrate to:

```text
HttpOnly Secure SameSite cookie
```

for stronger XSS resistance.

If localStorage is used for MVP, document it as a security tradeoff and keep authentication handling centralized so it can be replaced later.

---

# 40. Security Requirements

The backend must enforce:

### Authentication

Every protected API requires authentication.

### Authorization

Users can only access their own:

* projects
* messages
* files

### File Security

Uploaded files must not be executed.

Do not construct filesystem paths directly from user-supplied filenames.

Instead:

```text
database UUID -> storage key
```

Example:

```text
uploads/
  9a/9a3c...uuid...
```

Use generated IDs for storage keys.

### Password Security

Never log passwords.

### JWT

Keep JWT secret outside source code.

### Upload Limits

Enforce request/file size limits on the backend.

---

# 41. Observability

Backend should provide:

```text
GET /health
GET /ready
```

Example:

```json
{
  "status": "ok"
}
```

Use structured logs.

Every request should ideally have:

```text
request_id
user_id
project_id
```

where applicable.

Do not log:

* passwords
* JWT tokens
* complete Zscaler configurations
* sensitive configuration content

---

# 42. Testing

## Backend

Unit tests for:

* authentication
* password verification
* JWT validation
* project ownership
* project CRUD
* message CRUD
* file validation
* authorization
* agent service abstraction

Integration tests for:

```text
HTTP -> service -> PostgreSQL
```

At minimum test:

```text
User A cannot access User B's project.
User A cannot access User B's messages.
User A cannot access User B's uploaded files.
```

This is a critical security requirement.

---

# 43. Frontend Testing

Test:

* login success
* login failure
* unauthenticated redirect
* project creation
* drag-and-drop upload
* file upload failure
* project navigation
* rename
* delete
* sending a message
* loading states
* empty states

---

# 44. Agent Integration Contract

The future agent should integrate through a well-defined interface.

Initial interface:

```go
type AgentService interface {
    Run(
        ctx context.Context,
        request AgentRequest,
    ) (<-chan AgentEvent, error)
}
```

Example:

```go
type AgentRequest struct {
    UserID         string
    ProjectID string
    Message        string
}
```

Future agent implementation:

```text
AgentService
     |
     v
MigrationAgent
     |
     +-- LLM
     +-- Zscaler parser
     +-- Migration tools
     +-- Validation tools
     +-- RAG
```

The HTTP/API layer should remain unchanged.

---

# 45. Mock Agent

Implement a mock agent for the frontend/backend MVP.

Example behavior:

```text
User:
"Analyze this configuration"

Mock Agent:
"Agent integration is not implemented yet.
This is a placeholder response."
```

This allows the frontend team to implement the complete ChatGPT-like workflow before the real agent is available.

---

# 46. Recommended Development Sequence

## Phase 1 — Backend Foundation

1. Create Go project.
2. Create PostgreSQL connection.
3. Create migrations.
4. Implement user model/repository.
5. Implement password hashing.
6. Implement JWT authentication.
7. Implement authentication middleware.
8. Implement `/auth/login`.
9. Implement `/auth/me`.
10. Implement health endpoints.

## Phase 2 — Projects

1. Project model.
2. Project repository.
3. Project service.
4. Project CRUD APIs.
5. Ownership checks.
6. Tests.

## Phase 3 — File Upload

1. File storage interface.
2. Local filesystem implementation.
3. Multipart upload API.
4. File validation.
5. Project creation + file upload.
6. Tests.

## Phase 4 — Messages

1. Message model.
2. Message repository.
3. Message APIs.
4. Mock agent.
5. User → mock agent → assistant response flow.

## Phase 5 — Frontend

1. React application.
2. Routing.
3. Login page.
4. Authentication state.
5. Application layout.
6. Sidebar.
7. New project page.
8. Drag-and-drop configuration upload.
9. Project page.
10. Message UI.
11. Message input.
12. Rename/delete.
13. Error/loading states.

## Phase 6 — Agent Integration Readiness

1. Define `AgentService`.
2. Define `AgentRequest`.
3. Define `AgentEvent`.
4. Add streaming-compatible API design.
5. Replace mock agent with real agent implementation later.

---

# 47. MVP User Flow

Complete user journey:

```text
                  +-------------+
                  |    Login    |
                  +------+------+
                         |
                         v
                +----------------+
                | Chat Home Page |
                +-------+--------+
                        |
                        | New Project
                        v
              +----------------------+
              | Config Drop Zone     |
              |                      |
              | Drag Zscaler config  |
              +----------+-----------+
                         |
                         | Upload
                         v
              +----------------------+
              | Create Project  |
              |                      |
              | Store Config          |
              +----------+-----------+
                         |
                         v
              +----------------------+
              | Chat Project    |
              |                      |
              | Config loaded        |
              |                      |
              | User asks question   |
              +----------+-----------+
                         |
                         v
              +----------------------+
              | Mock Agent / Future  |
              | Agent Service        |
              +----------+-----------+
                         |
                         v
              +----------------------+
              | Assistant Response   |
              +----------------------+
```

---

# 48. Frontend UX Requirements

The application should feel like a specialized ChatGPT rather than a traditional enterprise form application.

Important UX principles:

* project navigation should be fast
* no unnecessary page reloads
* upload should be drag-and-drop first
* project should open immediately after upload
* message history should persist
* loading states should be obvious
* errors should be recoverable
* sidebar should remain persistent on desktop
* the UI should already accommodate streaming assistant responses

---

# 49. Suggested API Summary

```text
Authentication
-------------
POST   /api/v1/auth/login
GET    /api/v1/auth/me


Projects
-------------
GET    /api/v1/projects
POST   /api/v1/projects
GET    /api/v1/projects/{id}
PATCH  /api/v1/projects/{id}
DELETE /api/v1/projects/{id}


Messages
--------
GET    /api/v1/projects/{id}/messages
POST   /api/v1/projects/{id}/messages

Future:
POST   /api/v1/projects/{id}/messages/stream


Files
-----
GET    /api/v1/projects/{id}/files/{fileId}
DELETE /api/v1/projects/{id}/files/{fileId}


System
------
GET    /health
GET    /ready
```

The initial `POST /projects` endpoint can combine project creation and initial configuration upload to keep the frontend flow simple.

---

# 50. Definition of Done

The MVP is complete when a user can:

* [ ] Open the application.
* [ ] Log in with username/password.
* [ ] Remain authenticated across page navigation.
* [ ] See a ChatGPT-style project sidebar.
* [ ] Create a new project.
* [ ] Drag a Zscaler configuration into the upload area.
* [ ] Upload the configuration.
* [ ] Automatically navigate to the new project.
* [ ] See the uploaded configuration associated with the project.
* [ ] Send a chat message.
* [ ] Persist the message.
* [ ] Receive a mock assistant response.
* [ ] Refresh the browser and retain the project.
* [ ] Navigate between projects.
* [ ] Rename a project.
* [ ] Delete a project.
* [ ] Log out.
* [ ] Be prevented from accessing another user's projects/files/messages.

The backend is also considered ready for the agent team when the `AgentService` abstraction and API contract are implemented without requiring the frontend to know anything about the underlying agent implementation.

---

# 51. Coding Agent Instructions

When implementing this design:

1. Prefer simple, idiomatic Go over framework-heavy abstractions.
2. Keep HTTP handlers thin.
3. Keep business logic in services.
4. Keep database access in repositories.
5. Use interfaces at integration boundaries.
6. Do not implement migration-agent logic.
7. Implement a mock agent so the end-to-end chat workflow works.
8. Make the agent interface replaceable without changing HTTP handlers.
9. Do not trust user IDs supplied by clients.
10. Always derive the authenticated user from the authentication context.
11. Enforce ownership at the backend.
12. Do not store passwords in plaintext.
13. Do not log authentication tokens or uploaded configuration contents.
14. Keep configuration in environment variables.
15. Add unit tests for security-sensitive logic.
16. Add integration tests for project ownership.
17. Keep the API versioned under `/api/v1`.
18. Make file storage replaceable with S3/object storage later.
19. Design the message API to support SSE streaming later.
20. Keep frontend API access centralized rather than calling HTTP APIs directly from UI components.

The implementation should prioritize a clean MVP that can be extended by the agent team rather than prematurely implementing agent-specific functionality.
