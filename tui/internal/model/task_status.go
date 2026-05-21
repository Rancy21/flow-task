package model

type TaskStatus string

const (
	StatusInbox     TaskStatus = "INBOX"
	StatusScheduled TaskStatus = "SCHEDULED"
	StatusDone      TaskStatus = "DONE"
)
