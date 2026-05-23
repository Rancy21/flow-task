package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/Rancy21/flowtask/internal/model"
	"github.com/google/uuid"
)

type NoteRepo struct {
	db *sql.DB
}

func NewNoteRepo(db *sql.DB) *NoteRepo {
	return &NoteRepo{db: db}
}

func (r *NoteRepo) GetAllActive() ([]model.Note, error) {
	rows, err := r.db.Query(`
		SELECT n.id, n.task_id, n.content, n.created_at, n.updated_at
		FROM notes n
		JOIN tasks t ON t.id = n.task_id
		WHERE t.status != 'DONE'
		ORDER BY n.created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notes []model.Note
	for rows.Next() {
		var n model.Note
		var createdAt, updatedAt string
		if err := rows.Scan(&n.ID, &n.TaskID, &n.Content, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		t, err := time.Parse(time.RFC3339, createdAt)
		if err != nil {
			return nil, err
		}
		n.CreatedAt = t
		n.UpdatedAt = updatedAt
		notes = append(notes, n)
	}
	return notes, rows.Err()
}

func (r *NoteRepo) GetAll() ([]model.Note, error) {
	rows, err := r.db.Query(`
		SELECT n.id, n.task_id, n.content, n.created_at, n.updated_at
		FROM notes n
		ORDER BY n.created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notes []model.Note
	for rows.Next() {
		var n model.Note
		var createdAt, updatedAt string
		if err := rows.Scan(&n.ID, &n.TaskID, &n.Content, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		t, err := time.Parse(time.RFC3339, createdAt)
		if err != nil {
			return nil, err
		}
		n.CreatedAt = t
		n.UpdatedAt = updatedAt
		notes = append(notes, n)
	}
	return notes, rows.Err()
}

func (r *NoteRepo) GetByTask(taskID string) ([]model.Note, error) {
	rows, err := r.db.Query(`
		SELECT id, task_id, content, created_at, updated_at
		FROM notes WHERE task_id = ?
		ORDER BY created_at DESC
	`, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notes []model.Note
	for rows.Next() {
		var n model.Note
		var createdAt, updatedAt string
		if err := rows.Scan(&n.ID, &n.TaskID, &n.Content, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		t, err := time.Parse(time.RFC3339, createdAt)
		if err != nil {
			return nil, err
		}
		n.CreatedAt = t
		n.UpdatedAt = updatedAt
		notes = append(notes, n)
	}
	return notes, rows.Err()
}

func (r *NoteRepo) Create(taskID, content string) (*model.Note, error) {
	now := time.Now()
	n := &model.Note{
		ID:        uuid.New().String(),
		TaskID:    taskID,
		Content:   content,
		CreatedAt: now,
		UpdatedAt: now.Format(time.RFC3339),
	}
	_, err := r.db.Exec(`
		INSERT INTO notes (id, task_id, content, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`, n.ID, n.TaskID, n.Content, n.CreatedAt.Format(time.RFC3339), n.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create note: %w", err)
	}
	return n, nil
}

func (r *NoteRepo) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM notes WHERE id=?`, id)
	return err
}

// Upsert inserts a note or replaces if already exists (used during sync pull).
func (r *NoteRepo) Upsert(n *model.Note) error {
	_, err := r.db.Exec(`
		INSERT OR REPLACE INTO notes (id, task_id, content, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`, n.ID, n.TaskID, n.Content, n.CreatedAt.Format(time.RFC3339), n.UpdatedAt)
	return err
}
