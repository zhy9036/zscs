package file

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type LocalFileStorage struct {
	root string
}

func NewLocalFileStorage(root string) (*LocalFileStorage, error) {
	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, fmt.Errorf("create storage dir: %w", err)
	}
	return &LocalFileStorage{root: root}, nil
}

func (s *LocalFileStorage) Put(_ context.Context, key string, r io.Reader) error {
	full := s.fullPath(key)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return fmt.Errorf("create dir: %w", err)
	}
	f, err := os.Create(full)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer f.Close()
	if _, err := io.Copy(f, r); err != nil {
		return fmt.Errorf("write file: %w", err)
	}
	return nil
}

func (s *LocalFileStorage) Get(_ context.Context, key string) (io.ReadCloser, error) {
	f, err := os.Open(s.fullPath(key))
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	return f, nil
}

func (s *LocalFileStorage) Delete(_ context.Context, key string) error {
	err := os.Remove(s.fullPath(key))
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete file: %w", err)
	}
	return nil
}

func (s *LocalFileStorage) fullPath(key string) string {
	return filepath.Join(s.root, key)
}

// ShardedKey builds a storage key from a UUID to avoid filesystem collisions
// and to prevent path traversal from user-supplied filenames.
// e.g. "9a3c..." -> "9a/9a3c..."
func ShardedKey(id string) string {
	if len(id) < 2 {
		return id
	}
	return id[:2] + "/" + id
}
