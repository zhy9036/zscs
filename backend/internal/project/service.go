package project

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/zscaler/migration-platform/backend/internal/file"
)

type Service struct {
	repo   *Repository
	files  *file.Service
}

func NewService(repo *Repository, files *file.Service) *Service {
	return &Service{repo: repo, files: files}
}

type CreateInput struct {
	UserID   string
	Title    string
	Filename string
	ContentType string
	Size     int64
	Reader   io.Reader
}

type Detail struct {
	Project  Project    `json:"project"`
	Files    []file.File `json:"files"`
}

func (s *Service) Create(ctx context.Context, in CreateInput) (Project, file.File, error) {
	title := strings.TrimSpace(in.Title)
	if title == "" {
		title = titleFromFilename(in.Filename)
	}

	p, err := s.repo.Create(ctx, Project{UserID: in.UserID, Title: title})
	if err != nil {
		return Project{}, file.File{}, fmt.Errorf("create project: %w", err)
	}

	f, err := s.files.Upload(ctx, file.UploadInput{
		ProjectID:   p.ID,
		Filename:    in.Filename,
		ContentType: in.ContentType,
		Reader:      in.Reader,
		Size:        in.Size,
	})
	if err != nil {
		_ = s.repo.Delete(ctx, in.UserID, p.ID)
		return Project{}, file.File{}, fmt.Errorf("upload file: %w", err)
	}

	return p, f, nil
}

func (s *Service) Get(ctx context.Context, userID, id string) (Detail, error) {
	p, err := s.repo.GetByID(ctx, userID, id)
	if err != nil {
		return Detail{}, err
	}
	files, err := s.files.ListByProject(ctx, id)
	if err != nil {
		return Detail{}, err
	}
	return Detail{Project: p, Files: files}, nil
}

func (s *Service) List(ctx context.Context, userID string) ([]Project, error) {
	return s.repo.List(ctx, userID)
}

func (s *Service) Rename(ctx context.Context, userID, id, title string) (Project, error) {
	if strings.TrimSpace(title) == "" {
		return Project{}, errors.New("title must not be empty")
	}
	return s.repo.Update(ctx, userID, id, ProjectUpdate{Title: title})
}

func (s *Service) Delete(ctx context.Context, userID, id string) error {
	return s.repo.Delete(ctx, userID, id)
}

func (s *Service) MaxUploadSize() int64 {
	return s.files.MaxUploadSize()
}

func titleFromFilename(name string) string {
	base := filepath.Base(name)
	if dot := strings.LastIndex(base, "."); dot > 0 {
		base = base[:dot]
	}
	if base == "" {
		base = "Untitled"
	}
	return base
}
