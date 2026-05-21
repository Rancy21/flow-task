package model

import "time"

type Task struct {
	ID            string
	Title         string
	Description   string
	Priority      Priority
	Status        TaskStatus
	ScheduledDate *time.Time // nil = lives in Index
	CreatedAt     time.Time
	CompletedAt   *time.Time
}

func (t *Task) IsToday() bool {
	if t.ScheduledDate == nil {
		return false
	}
	now := time.Now()
	d := *t.ScheduledDate
	return d.Year() == now.Year() && d.Month() == now.Month() && d.Day() == now.Day()
}

func (t *Task) IsInbox() bool {
	return t.Status == StatusInbox || t.ScheduledDate == nil
}

func (t *Task) IsDone() bool {
	return t.Status == StatusDone
}
