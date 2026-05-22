package model

import "time"

type Note struct {
	ID        string
	TaskID    string
	Content   string
	CreatedAt time.Time
	UpdatedAt string // RFC 3339, used for sync conflict resolution
}
