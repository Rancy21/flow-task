package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/Rancy21/flowtask/internal/model"
	"github.com/google/uuid"
)

// TaskRepo handles all task persistence.
type TaskRepo struct {
	db *sql.DB
}

func NewTaskRepo(db *sql.DB) *TaskRepo {
	return &TaskRepo{db: db}
}

// ── Queries ───────────────────────────────────────────────────────────────────

func (r *TaskRepo) GetToday() ([]model.Task, error) {
	today := time.Now().Format("2006-01-02")
	rows, err := r.db.Query(`
		SELECT id, title, description, priority, status, scheduled_date, created_at, completed_at, updated_at
		FROM tasks
		WHERE scheduled_date = ? AND status != 'DONE'
		ORDER BY
			CASE priority WHEN 'P1' THEN 0 WHEN 'P2' THEN 1 ELSE 2 END,
			created_at ASC
	`, today)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTasks(rows)
}

func (r *TaskRepo) GetWeek() ([]model.Task, error) {
	monday, sunday := weekBounds(time.Now())
	rows, err := r.db.Query(`
		SELECT id, title, description, priority, status, scheduled_date, created_at, completed_at, updated_at
		FROM tasks
		WHERE scheduled_date BETWEEN ? AND ? AND status != 'DONE'
		ORDER BY
			scheduled_date ASC,
			CASE priority WHEN 'P1' THEN 0 WHEN 'P2' THEN 1 ELSE 2 END,
			created_at ASC
	`, monday, sunday)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTasks(rows)
}

func (r *TaskRepo) GetInbox() ([]model.Task, error) {
	rows, err := r.db.Query(`
		SELECT id, title, description, priority, status, scheduled_date, created_at, completed_at, updated_at
		FROM tasks
		WHERE status = 'INBOX' OR scheduled_date IS NULL
		ORDER BY
			CASE priority WHEN 'P1' THEN 0 WHEN 'P2' THEN 1 ELSE 2 END,
			created_at ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTasks(rows)
}

func (r *TaskRepo) GetByID(id string) (*model.Task, error) {
	row := r.db.QueryRow(`
		SELECT id, title, description, priority, status, scheduled_date, created_at, completed_at, updated_at
		FROM tasks WHERE id = ?
	`, id)
	return scanTask(row)
}

// ── Mutations ─────────────────────────────────────────────────────────────────

func (r *TaskRepo) Create(t *model.Task) (*model.Task, error) {
	t.ID = uuid.New().String()
	now := time.Now()
	t.CreatedAt = now
	t.UpdatedAt = now.Format(time.RFC3339)

	if t.ScheduledDate != nil {
		t.Status = model.StatusScheduled
	} else {
		t.Status = model.StatusInbox
	}

	_, err := r.db.Exec(`
		INSERT INTO tasks (id, title, description, priority, status, scheduled_date, created_at, completed_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		t.ID, t.Title, nullString(t.Description), string(t.Priority), string(t.Status),
		nullDate(t.ScheduledDate), t.CreatedAt.Format(time.RFC3339), nullTime(t.CompletedAt), t.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create task: %w", err)
	}
	return t, nil
}

func (r *TaskRepo) Update(t *model.Task) error {
	if t.ScheduledDate != nil && t.Status == model.StatusInbox {
		t.Status = model.StatusScheduled
	}
	now := time.Now().Format(time.RFC3339)
	_, err := r.db.Exec(`
		UPDATE tasks
		SET title=?, description=?, priority=?, status=?, scheduled_date=?, completed_at=?, updated_at=?
		WHERE id=?
	`, t.Title, nullString(t.Description), string(t.Priority), string(t.Status),
		nullDate(t.ScheduledDate), nullTime(t.CompletedAt), now, t.ID)
	return err
}

func (r *TaskRepo) MarkDone(id string) error {
	now := time.Now()
	_, err := r.db.Exec(`
		UPDATE tasks SET status='DONE', completed_at=?, updated_at=? WHERE id=?
	`, now.Format(time.RFC3339), now.Format(time.RFC3339), id)
	return err
}

func (r *TaskRepo) Schedule(id string, date time.Time) error {
	now := time.Now().Format(time.RFC3339)
	_, err := r.db.Exec(`
		UPDATE tasks SET status='SCHEDULED', scheduled_date=?, updated_at=? WHERE id=?
	`, date.Format("2006-01-02"), now, id)
	return err
}

func (r *TaskRepo) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM tasks WHERE id=?`, id)
	return err
}

// ── Sync ─────────────────────────────────────────────────────────────────────

// Upsert inserts a task or replaces if it already exists (used during sync pull).
func (r *TaskRepo) Upsert(t *model.Task) error {
	_, err := r.db.Exec(`
		INSERT OR REPLACE INTO tasks (id, title, description, priority, status, scheduled_date, created_at, completed_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, t.ID, t.Title, nullString(t.Description), string(t.Priority), string(t.Status),
		nullDate(t.ScheduledDate), t.CreatedAt.Format(time.RFC3339), nullTime(t.CompletedAt), t.UpdatedAt)
	return err
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func scanTasks(rows *sql.Rows) ([]model.Task, error) {
	var tasks []model.Task
	for rows.Next() {
		t, err := scanRowTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, *t)
	}
	return tasks, rows.Err()
}

func scanTask(row *sql.Row) (*model.Task, error) {
	var (
		t             model.Task
		description   sql.NullString
		scheduledDate sql.NullString
		completedAt   sql.NullString
		createdAt     string
		updatedAt     sql.NullString
		priority      string
		status        string
	)
	err := row.Scan(&t.ID, &t.Title, &description, &priority, &status,
		&scheduledDate, &createdAt, &completedAt, &updatedAt)
	if err != nil {
		return nil, err
	}
	return buildTask(&t, description, priority, status, scheduledDate, createdAt, completedAt, updatedAt)
}

func scanRowTask(rows *sql.Rows) (*model.Task, error) {
	var (
		t             model.Task
		description   sql.NullString
		scheduledDate sql.NullString
		completedAt   sql.NullString
		createdAt     string
		updatedAt     sql.NullString
		priority      string
		status        string
	)
	err := rows.Scan(&t.ID, &t.Title, &description, &priority, &status,
		&scheduledDate, &createdAt, &completedAt, &updatedAt)
	if err != nil {
		return nil, err
	}
	return buildTask(&t, description, priority, status, scheduledDate, createdAt, completedAt, updatedAt)
}

func buildTask(
	t *model.Task,
	description sql.NullString,
	priority, status string,
	scheduledDate sql.NullString,
	createdAt string,
	completedAt sql.NullString,
	updatedAt sql.NullString,
) (*model.Task, error) {
	t.Description = description.String
	t.Priority = model.Priority(priority)
	t.Status = model.TaskStatus(status)
	t.UpdatedAt = updatedAt.String

	if scheduledDate.Valid {
		d, err := time.Parse("2006-01-02", scheduledDate.String)
		if err != nil {
			return nil, err
		}
		t.ScheduledDate = &d
	}

	ca, err := time.Parse(time.RFC3339, createdAt)
	if err != nil {
		return nil, err
	}
	t.CreatedAt = ca

	if completedAt.Valid {
		ct, err := time.Parse(time.RFC3339, completedAt.String)
		if err != nil {
			return nil, err
		}
		t.CompletedAt = &ct
	}

	return t, nil
}

func weekBounds(t time.Time) (string, string) {
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	monday := t.AddDate(0, 0, -(weekday - 1))
	sunday := monday.AddDate(0, 0, 6)
	return monday.Format("2006-01-02"), sunday.Format("2006-01-02")
}

func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

func nullDate(t *time.Time) sql.NullString {
	if t == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: t.Format("2006-01-02"), Valid: true}
}

func nullTime(t *time.Time) sql.NullString {
	if t == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: t.Format(time.RFC3339), Valid: true}
}
