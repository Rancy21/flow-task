package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/Rancy21/flowtask/internal/model"
	"github.com/google/uuid"
)

type InboxRepo struct {
	db *sql.DB
}

func NewInboxRepo(db *sql.DB) *InboxRepo {
	return &InboxRepo{db: db}
}

func (r *InboxRepo) GetAll() ([]model.InboxItem, error) {
	rows, err := r.db.Query(`
		SELECT id, title, description, created_at, updated_at
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
		var createdAt, updatedAt string
		if err := rows.Scan(&it.ID, &it.Title, &desc, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		it.Description = desc.String
		t, err := time.Parse(time.RFC3339, createdAt)
		if err != nil {
			return nil, err
		}
		it.CreatedAt = t
		it.UpdatedAt = updatedAt
		items = append(items, it)
	}
	return items, rows.Err()
}

func (r *InboxRepo) GetByID(id string) (*model.InboxItem, error) {
	var it model.InboxItem
	var desc sql.NullString
	var createdAt, updatedAt string
	err := r.db.QueryRow(`
		SELECT id, title, description, created_at, updated_at
		FROM inbox WHERE id = ?
	`, id).Scan(&it.ID, &it.Title, &desc, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}
	it.Description = desc.String
	t, err := time.Parse(time.RFC3339, createdAt)
	if err != nil {
		return nil, err
	}
	it.CreatedAt = t
	it.UpdatedAt = updatedAt
	return &it, nil
}

func (r *InboxRepo) Create(title, description string) (*model.InboxItem, error) {
	now := time.Now()
	it := &model.InboxItem{
		ID:          uuid.New().String(),
		Title:       title,
		Description: description,
		CreatedAt:   now,
		UpdatedAt:   now.Format(time.RFC3339),
	}
	_, err := r.db.Exec(`
		INSERT INTO inbox (id, title, description, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`, it.ID, it.Title, nullInboxString(it.Description), it.CreatedAt.Format(time.RFC3339), it.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create inbox item: %w", err)
	}
	return it, nil
}

func (r *InboxRepo) Update(it *model.InboxItem) error {
	now := time.Now().Format(time.RFC3339)
	_, err := r.db.Exec(`
		UPDATE inbox SET title=?, description=?, updated_at=? WHERE id=?
	`, it.Title, nullInboxString(it.Description), now, it.ID)
	return err
}

func (r *InboxRepo) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM inbox WHERE id=?`, id)
	return err
}

// Upsert inserts an inbox item or replaces if already exists (used during sync pull).
func (r *InboxRepo) Upsert(it *model.InboxItem) error {
	_, err := r.db.Exec(`
		INSERT OR REPLACE INTO inbox (id, title, description, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`, it.ID, it.Title, nullInboxString(it.Description), it.CreatedAt.Format(time.RFC3339), it.UpdatedAt)
	return err
}

func nullInboxString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}
