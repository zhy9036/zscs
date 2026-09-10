package file

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

var (
	ErrTooLarge     = errors.New("file too large")
	ErrUnsupported  = errors.New("unsupported file type")
	ErrEmpty        = errors.New("file is empty")
)

var allowedExtensions = map[string]bool{
	".conf": true,
	".txt":  true,
	".xml":  true,
	".json": true,
	".csv":  true,
}

var allowedContentTypes = map[string]bool{
	"text/plain":                true,
	"text/xml":                  true,
	"application/json":          true,
	"application/xml":           true,
	"application/octet-stream":  true,
	"text/csv":                  true,
}

type ProjectOwnershipChecker interface {
	EnsureOwned(ctx context.Context, userID, projectID string) error
}

type Repo interface {
	Create(ctx context.Context, f File) (File, error)
	GetByID(ctx context.Context, projectID, id string) (File, error)
	ListByProject(ctx context.Context, projectID string) ([]File, error)
	Delete(ctx context.Context, projectID, id string) error
}

type Service struct {
	storage       FileStorage
	repo          Repo
	maxUploadSize int64
	ownership     ProjectOwnershipChecker
}

func NewService(storage FileStorage, repo Repo, maxUploadSize int64, ownership ProjectOwnershipChecker) *Service {
	return &Service{
		storage:       storage,
		repo:          repo,
		maxUploadSize: maxUploadSize,
		ownership:     ownership,
	}
}

// UploadInput carries the metadata for a file being uploaded.
type UploadInput struct {
	ProjectID   string
	Filename    string
	ContentType string
	Reader      io.Reader
	Size        int64
}

func (s *Service) Upload(ctx context.Context, in UploadInput) (File, error) {
	if err := validate(in.Filename, in.ContentType, in.Size); err != nil {
		return File{}, err
	}

	fileID := uuid.NewString()
	storageKey := ShardedKey(fileID)

	if err := s.storage.Put(ctx, storageKey, in.Reader); err != nil {
		return File{}, fmt.Errorf("store file: %w", err)
	}

	f, err := s.repo.Create(ctx, File{
		ProjectID:   in.ProjectID,
		Filename:    sanitizeFilename(in.Filename),
		ContentType: normalizeContentType(in.ContentType, in.Filename),
		Size:        in.Size,
		StorageKey:  storageKey,
	})
	if err != nil {
		_ = s.storage.Delete(ctx, storageKey)
		return File{}, err
	}

	return f, nil
}

func (s *Service) Get(ctx context.Context, userID, projectID, id string) (File, io.ReadCloser, error) {
	if err := s.ownership.EnsureOwned(ctx, userID, projectID); err != nil {
		return File{}, nil, ErrNotFound
	}
	f, err := s.repo.GetByID(ctx, projectID, id)
	if err != nil {
		return File{}, nil, err
	}
	rc, err := s.storage.Get(ctx, f.StorageKey)
	if err != nil {
		return File{}, nil, err
	}
	return f, rc, nil
}

func (s *Service) ListByProject(ctx context.Context, projectID string) ([]File, error) {
	return s.repo.ListByProject(ctx, projectID)
}

func (s *Service) Delete(ctx context.Context, userID, projectID, id string) error {
	if err := s.ownership.EnsureOwned(ctx, userID, projectID); err != nil {
		return ErrNotFound
	}
	f, err := s.repo.GetByID(ctx, projectID, id)
	if err != nil {
		return err
	}
	if err := s.storage.Delete(ctx, f.StorageKey); err != nil {
		return err
	}
	return s.repo.Delete(ctx, projectID, id)
}

func (s *Service) MaxUploadSize() int64 {
	return s.maxUploadSize
}

func validate(filename, contentType string, size int64) error {
	if size <= 0 {
		return ErrEmpty
	}
	if size > 0 && contentType == "" {
		// allow — will infer from extension
	}
	ext := strings.ToLower(filepath.Ext(filename))
	if !allowedExtensions[ext] {
		return ErrUnsupported
	}
	ct := strings.ToLower(strings.TrimSpace(contentType))
	if ct != "" && !allowedContentTypes[ct] {
		return ErrUnsupported
	}
	return nil
}

func sanitizeFilename(name string) string {
	name = filepath.Base(name)
	if name == "" || name == "." || name == "/" || name == "\\" {
		name = "upload"
	}
	return name
}

func normalizeContentType(ct, filename string) string {
	if ct != "" {
		return ct
	}
	if ext := filepath.Ext(filename); ext != "" {
		if m := mime.TypeByExtension(ext); m != "" {
			return m
		}
	}
	return "application/octet-stream"
}
