package model

import "time"

type Note struct {
	ID        string
	TaskID    string
	Content   string
	CreatedAt time.Time
}
