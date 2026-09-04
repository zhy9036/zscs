package file

import (
	"time"
)

type File struct {
	ID          string    `json:"id"`
	ProjectID   string    `json:"project_id"`
	Filename    string    `json:"filename"`
	ContentType string    `json:"content_type"`
	Size        int64     `json:"size"`
	StorageKey  string    `json:"-"`
	CreatedAt   time.Time `json:"created_at"`
}
