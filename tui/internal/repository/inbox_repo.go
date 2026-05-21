package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/Rancy21/flowtask/internal/model"
	"github.com/google/uuid"
)

// InboxRepo handles all inbox item persistence.
type InboxRepo struct {
	db *sql.DB
}

func NewInboxRepo(db *sql.DB) *InboxRepo {
	return &InboxRepo{db: db}
}

// GetAll returns all inbox items, newest first.
func (r *InboxRepo) GetAll() ([]model.InboxItem, error) {
	rows, err := r.db.Query(`
		SELECT id, title, description, created_at
		FROM inbox
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []model.InboxItem
	for rows.Next() {
		var it model.InboxItem
		var desc sql.NullString
		var createdAt string
		if err := rows.Scan(&it.ID, &it.Title, &desc, &createdAt); err != nil {
			return nil, err
		}
		it.Description = desc.String
		t, err := time.Parse(time.RFC3339, createdAt)
		if err != nil {
			return nil, err
		}
		it.CreatedAt = t
		items = append(items, it)
	}
	return items, rows.Err()
}

// GetByID returns a single inbox item.
func (r *InboxRepo) GetByID(id string) (*model.InboxItem, error) {
	var it model.InboxItem
	var desc sql.NullString
	var createdAt string
	err := r.db.QueryRow(`
		SELECT id, title, description, created_at
		FROM inbox WHERE id = ?
	`, id).Scan(&it.ID, &it.Title, &desc, &createdAt)
	if err != nil {
		return nil, err
	}
	it.Description = desc.String
	t, err := time.Parse(time.RFC3339, createdAt)
	if err != nil {
		return nil, err
	}
	it.CreatedAt = t
	return &it, nil
}

// Create inserts a new inbox item.
func (r *InboxRepo) Create(title, description string) (*model.InboxItem, error) {
	it := &model.InboxItem{
		ID:          uuid.New().String(),
		Title:       title,
		Description: description,
		CreatedAt:   time.Now(),
	}
	_, err := r.db.Exec(`
		INSERT INTO inbox (id, title, description, created_at)
		VALUES (?, ?, ?, ?)
	`, it.ID, it.Title, nullInboxString(it.Description), it.CreatedAt.Format(time.RFC3339))
	if err != nil {
		return nil, fmt.Errorf("create inbox item: %w", err)
	}
	return it, nil
}

// Update saves changes to an existing inbox item.
func (r *InboxRepo) Update(it *model.InboxItem) error {
	_, err := r.db.Exec(`
		UPDATE inbox SET title=?, description=? WHERE id=?
	`, it.Title, nullInboxString(it.Description), it.ID)
	return err
}

// Delete removes an inbox item.
func (r *InboxRepo) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM inbox WHERE id=?`, id)
	return err
}

func nullInboxString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}
