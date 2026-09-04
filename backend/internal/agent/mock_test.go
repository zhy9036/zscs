package agent

import (
	"context"
	"testing"
)

func TestMockEmitsMessageThenDone(t *testing.T) {
	m := NewMock()
	ch, err := m.Run(context.Background(), Request{
		UserID:    "u1",
		ProjectID: "p1",
		Message:   "analyze",
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	var events []Event
	for ev := range ch {
		events = append(events, ev)
	}

	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}
	if events[0].Type != EventMessage {
		t.Fatalf("first event = %s, want message", events[0].Type)
	}
	if events[1].Type != EventDone {
		t.Fatalf("second event = %s, want done", events[1].Type)
	}
}
