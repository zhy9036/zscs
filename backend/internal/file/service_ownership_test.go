package file

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"
)

type fakeRepo struct {
	file    File
	err     error
	deleted bool
}

func (f *fakeRepo) Create(_ context.Context, fl File) (File, error)         { return fl, nil }
func (f *fakeRepo) GetByID(_ context.Context, _, _ string) (File, error)     { return f.file, f.err }
func (f *fakeRepo) ListByProject(_ context.Context, _ string) ([]File, error) { return nil, nil }
func (f *fakeRepo) Delete(_ context.Context, _, _ string) error {
	f.deleted = true
	return nil
}

type fakeStorage struct{}

func (fakeStorage) Put(_ context.Context, _ string, r io.Reader) error  { _, _ = io.Copy(io.Discard, r); return nil }
func (fakeStorage) Get(_ context.Context, _ string) (io.ReadCloser, error) {
	return io.NopCloser(bytes.NewReader(nil)), nil
}
func (fakeStorage) Delete(_ context.Context, _ string) error { return nil }

type fakeOwnership struct {
	owned bool
}

func (f fakeOwnership) EnsureOwned(_ context.Context, _, _ string) error {
	if !f.owned {
		return errors.New("not owned")
	}
	return nil
}

func TestGet_RejectsNonOwner(t *testing.T) {
	svc := NewService(fakeStorage{}, &fakeRepo{file: File{ID: "f1", ProjectID: "p1", StorageKey: "x"}}, 1000, fakeOwnership{owned: false})

	_, _, err := svc.Get(context.Background(), "userA", "p1", "f1")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound for non-owner, got %v", err)
	}
}

func TestGet_AllowsOwner(t *testing.T) {
	svc := NewService(fakeStorage{}, &fakeRepo{file: File{ID: "f1", ProjectID: "p1", StorageKey: "x"}}, 1000, fakeOwnership{owned: true})

	_, _, err := svc.Get(context.Background(), "userA", "p1", "f1")
	if err != nil {
		t.Fatalf("expected no error for owner, got %v", err)
	}
}

func TestDelete_RejectsNonOwner(t *testing.T) {
	svc := NewService(fakeStorage{}, &fakeRepo{file: File{ID: "f1", ProjectID: "p1", StorageKey: "x"}}, 1000, fakeOwnership{owned: false})

	if err := svc.Delete(context.Background(), "userA", "p1", "f1"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound for non-owner, got %v", err)
	}
}
