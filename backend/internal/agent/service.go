package agent

import "context"

// Event types emitted by an AgentService.
const (
	EventMessage    = "message"
	EventThinking  = "thinking"
	EventToolStart  = "tool_start"
	EventToolResult = "tool_result"
	EventFile       = "file"
	EventError      = "error"
	EventDone       = "done"
)

type Request struct {
	UserID    string
	ProjectID string
	Message   string
}

type Event struct {
	Type string `json:"type"`
	Data any    `json:"data,omitempty"`
}

// Service is the integration contract for the future migration agent.
// The HTTP layer depends only on this interface.
type Service interface {
	Run(ctx context.Context, req Request) (<-chan Event, error)
}
