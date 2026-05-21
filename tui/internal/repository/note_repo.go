package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/Rancy21/flowtask/internal/model"
	"github.com/google/uuid"
)

// NoteRepo handles all note persistence.
type NoteRepo struct {
	db *sql.DB
}

func NewNoteRepo(db *sql.DB) *NoteRepo {
	return &NoteRepo{db: db}
}

// GetAll returns all notes, newest first, with their linked task title.
func (r *NoteRepo) GetAll() ([]model.Note, error) {
	rows, err := r.db.Query(`
		SELECT n.id, n.task_id, n.content, n.created_at
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
		var createdAt string
		if err := rows.Scan(&n.ID, &n.TaskID, &n.Content, &createdAt); err != nil {
			return nil, err
		}
		t, err := time.Parse(time.RFC3339, createdAt)
		if err != nil {
			return nil, err
		}
		n.CreatedAt = t
		notes = append(notes, n)
	}
	return notes, rows.Err()
}

// GetByTask returns all notes for a specific task.
func (r *NoteRepo) GetByTask(taskID string) ([]model.Note, error) {
	rows, err := r.db.Query(`
		SELECT id, task_id, content, created_at
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
		var createdAt string
		if err := rows.Scan(&n.ID, &n.TaskID, &n.Content, &createdAt); err != nil {
			return nil, err
		}
		t, err := time.Parse(time.RFC3339, createdAt)
		if err != nil {
			return nil, err
		}
		n.CreatedAt = t
		notes = append(notes, n)
	}
	return notes, rows.Err()
}

// Create inserts a new note linked to a task.
func (r *NoteRepo) Create(taskID, content string) (*model.Note, error) {
	n := &model.Note{
		ID:        uuid.New().String(),
		TaskID:    taskID,
		Content:   content,
		CreatedAt: time.Now(),
	}
	_, err := r.db.Exec(`
		INSERT INTO notes (id, task_id, content, created_at)
		VALUES (?, ?, ?, ?)
	`, n.ID, n.TaskID, n.Content, n.CreatedAt.Format(time.RFC3339))
	if err != nil {
		return nil, fmt.Errorf("create note: %w", err)
	}
	return n, nil
}

// Delete removes a note by ID.
func (r *NoteRepo) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM notes WHERE id=?`, id)
	return err
}
