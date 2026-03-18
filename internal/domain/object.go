package domain

import "time"

// Object represents the core entity for our storage service.
type Object struct {
	ID          string            `json:"id"`
	Key         string            `json:"key"`
	Size        int64             `json:"size"`
	ContentType string            `json:"content_type"`
	Metadata    map[string]string `json:"metadata"`
	CreatedAt   time.Time         `json:"created_at"`
}
