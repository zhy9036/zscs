package message

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/zscaler/migration-platform/backend/internal/agent"
	"github.com/zscaler/migration-platform/backend/internal/project"
)

type Service struct {
	repo     *Repository
	projects *project.Repository
	agent    agent.Service
}

func NewService(repo *Repository, projects *project.Repository, agent agent.Service) *Service {
	return &Service{repo: repo, projects: projects, agent: agent}
}

type SendResult struct {
	UserMessage      Message
	AssistantMessage *Message
	Status           string
}

func (s *Service) List(ctx context.Context, userID, projectID string) ([]Message, error) {
	if err := s.ensureOwned(ctx, userID, projectID); err != nil {
		return nil, err
	}
	return s.repo.ListByProject(ctx, projectID)
}

func (s *Service) Send(ctx context.Context, userID, projectID, content string) (SendResult, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return SendResult{}, errors.New("message content must not be empty")
	}

	if err := s.ensureOwned(ctx, userID, projectID); err != nil {
		return SendResult{}, err
	}

	userMsg, err := s.repo.Create(ctx, Message{
		ProjectID: projectID,
		Role:      RoleUser,
		Content:   content,
	})
	if err != nil {
		return SendResult{}, err
	}

	events, err := s.agent.Run(ctx, agent.Request{
		UserID:    userID,
		ProjectID: projectID,
		Message:   content,
	})
	if err != nil {
		return SendResult{UserMessage: userMsg, Status: "agent_error"}, nil
	}

	var assistantText strings.Builder
	for ev := range events {
		if ev.Type == agent.EventMessage {
			if s, ok := ev.Data.(map[string]string); ok && s["content"] != "" {
				assistantText.WriteString(s["content"])
			}
		}
	}

	result := SendResult{UserMessage: userMsg, Status: "accepted"}
	if assistantText.Len() > 0 {
		asst, err := s.repo.Create(ctx, Message{
			ProjectID: projectID,
			Role:      RoleAssistant,
			Content:   assistantText.String(),
		})
		if err != nil {
			return result, err
		}
		result.AssistantMessage = &asst
	}

	_ = s.projects.Touch(ctx, userID, projectID)
	return result, nil
}

func (s *Service) ensureOwned(ctx context.Context, userID, projectID string) error {
	_, err := s.projects.GetByID(ctx, userID, projectID)
	if err != nil {
		if errors.Is(err, project.ErrNotFound) {
			return project.ErrNotFound
		}
		return fmt.Errorf("verify project ownership: %w", err)
	}
	return nil
}
