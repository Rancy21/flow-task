package model

import "time"

// InboxItem is a lightweight capture item — not a task.
// It lives in the Inbox tab as a brain-dump zone.
// The user manually creates a task from an inbox item when ready.
type InboxItem struct {
	ID          string
	Title       string
	Description string
	CreatedAt   time.Time
}
