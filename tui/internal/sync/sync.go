package sync

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Rancy21/flowtask/internal/model"
	"github.com/Rancy21/flowtask/internal/repository"
)

const (
	supabaseURL = "https://ykksgiyweklxbrfoomwa.supabase.co"
	supabaseKey = "sb_publishable_8JsM8svXjt1-yagX9M0n_w_jN7OWZBZ"
)

// Client syncs local SQLite with remote Supabase.
type Client struct {
	http   *http.Client
	taskDB *repository.TaskRepo
	noteDB *repository.NoteRepo
	inboxDB *repository.InboxRepo
}

func New(taskDB *repository.TaskRepo, noteDB *repository.NoteRepo, inboxDB *repository.InboxRepo) *Client {
	return &Client{
		http:    &http.Client{Timeout: 15 * time.Second},
		taskDB:  taskDB,
		noteDB:  noteDB,
		inboxDB: inboxDB,
	}
}

// ── Pull from Supabase ────────────────────────────────────────────────────────

// PullAll fetches all records from Supabase and upserts them locally.
func (c *Client) PullAll() error {
	if err := c.pullTasks(""); err != nil {
		return fmt.Errorf("pull tasks: %w", err)
	}
	if err := c.pullNotes(""); err != nil {
		return fmt.Errorf("pull notes: %w", err)
	}
	if err := c.pullInbox(""); err != nil {
		return fmt.Errorf("pull inbox: %w", err)
	}
	return nil
}

func (c *Client) pullTasks(since string) error {
	url := fmt.Sprintf("%s/rest/v1/tasks?select=*", supabaseURL)
	if since != "" {
		url += fmt.Sprintf("&updated_at=gte.%s", since)
	}
	url += "&order=updated_at.asc"

	var tasks []supabaseTask
	if err := c.get(url, &tasks); err != nil {
		return err
	}
	for _, st := range tasks {
		t := st.toModel()
		if err := c.taskDB.Upsert(t); err != nil {
			return fmt.Errorf("upsert task %s: %w", t.ID, err)
		}
	}
	return nil
}

func (c *Client) pullNotes(since string) error {
	url := fmt.Sprintf("%s/rest/v1/notes?select=*", supabaseURL)
	if since != "" {
		url += fmt.Sprintf("&updated_at=gte.%s", since)
	}
	url += "&order=updated_at.asc"

	var notes []supabaseNote
	if err := c.get(url, &notes); err != nil {
		return err
	}
	for _, sn := range notes {
		n := sn.toModel()
		if err := c.noteDB.Upsert(n); err != nil {
			return fmt.Errorf("upsert note %s: %w", n.ID, err)
		}
	}
	return nil
}

func (c *Client) pullInbox(since string) error {
	url := fmt.Sprintf("%s/rest/v1/inbox?select=*", supabaseURL)
	if since != "" {
		url += fmt.Sprintf("&updated_at=gte.%s", since)
	}
	url += "&order=updated_at.asc"

	var items []supabaseInboxItem
	if err := c.get(url, &items); err != nil {
		return err
	}
	for _, si := range items {
		it := si.toModel()
		if err := c.inboxDB.Upsert(it); err != nil {
			return fmt.Errorf("upsert inbox %s: %w", it.ID, err)
		}
	}
	return nil
}

// ── Push to Supabase ─────────────────────────────────────────────────────────

// PushTask creates or updates a task on Supabase.
func (c *Client) PushTask(t *model.Task) error {
	st := taskToSupabase(t)
	body, _ := json.Marshal(st)

	// Try PATCH first (update), fall back to POST (create)
	url := fmt.Sprintf("%s/rest/v1/tasks?id=eq.%s", supabaseURL, t.ID)
	resp, err := c.do("PATCH", url, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// If 404 (not found), create instead
	if resp.StatusCode == 404 || resp.StatusCode == 204 && c.isEmptyPatch(resp) {
		return c.postTask(t)
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("push task %s: HTTP %d", t.ID, resp.StatusCode)
	}
	return nil
}

func (c *Client) postTask(t *model.Task) error {
	st := taskToSupabase(t)
	body, _ := json.Marshal(st)
	url := fmt.Sprintf("%s/rest/v1/tasks", supabaseURL)

	req, _ := http.NewRequest("POST", url, bytes.NewReader(body))
	c.setHeaders(req)
	req.Header.Set("Prefer", "return=minimal")

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("POST task %s: HTTP %d", t.ID, resp.StatusCode)
	}
	return nil
}

// PushNote creates a note on Supabase.
func (c *Client) PushNote(n *model.Note) error {
	sn := noteToSupabase(n)
	body, _ := json.Marshal(sn)
	url := fmt.Sprintf("%s/rest/v1/notes?id=eq.%s", supabaseURL, n.ID)

	resp, err := c.do("PATCH", url, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 || (resp.StatusCode == 204 && c.isEmptyPatch(resp)) {
		return c.postNote(n)
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("push note %s: HTTP %d", n.ID, resp.StatusCode)
	}
	return nil
}

func (c *Client) postNote(n *model.Note) error {
	sn := noteToSupabase(n)
	body, _ := json.Marshal(sn)
	url := fmt.Sprintf("%s/rest/v1/notes", supabaseURL)

	req, _ := http.NewRequest("POST", url, bytes.NewReader(body))
	c.setHeaders(req)
	req.Header.Set("Prefer", "return=minimal")

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("POST note %s: HTTP %d", n.ID, resp.StatusCode)
	}
	return nil
}

// PushInboxItem creates or updates an inbox item on Supabase.
func (c *Client) PushInboxItem(it *model.InboxItem) error {
	si := inboxToSupabase(it)
	body, _ := json.Marshal(si)
	url := fmt.Sprintf("%s/rest/v1/inbox?id=eq.%s", supabaseURL, it.ID)

	resp, err := c.do("PATCH", url, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 || (resp.StatusCode == 204 && c.isEmptyPatch(resp)) {
		return c.postInboxItem(it)
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("push inbox %s: HTTP %d", it.ID, resp.StatusCode)
	}
	return nil
}

func (c *Client) postInboxItem(it *model.InboxItem) error {
	si := inboxToSupabase(it)
	body, _ := json.Marshal(si)
	url := fmt.Sprintf("%s/rest/v1/inbox", supabaseURL)

	req, _ := http.NewRequest("POST", url, bytes.NewReader(body))
	c.setHeaders(req)
	req.Header.Set("Prefer", "return=minimal")

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("POST inbox %s: HTTP %d", it.ID, resp.StatusCode)
	}
	return nil
}

// DeleteTask removes a task from Supabase.
func (c *Client) DeleteTask(id string) error {
	url := fmt.Sprintf("%s/rest/v1/tasks?id=eq.%s", supabaseURL, id)
	req, _ := http.NewRequest("DELETE", url, nil)
	c.setHeaders(req)
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 && resp.StatusCode != 404 {
		return fmt.Errorf("delete task %s: HTTP %d", id, resp.StatusCode)
	}
	return nil
}

// DeleteNote removes a note from Supabase.
func (c *Client) DeleteNote(id string) error {
	url := fmt.Sprintf("%s/rest/v1/notes?id=eq.%s", supabaseURL, id)
	req, _ := http.NewRequest("DELETE", url, nil)
	c.setHeaders(req)
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 && resp.StatusCode != 404 {
		return fmt.Errorf("delete note %s: HTTP %d", id, resp.StatusCode)
	}
	return nil
}

// DeleteInboxItem removes an inbox item from Supabase.
func (c *Client) DeleteInboxItem(id string) error {
	url := fmt.Sprintf("%s/rest/v1/inbox?id=eq.%s", supabaseURL, id)
	req, _ := http.NewRequest("DELETE", url, nil)
	c.setHeaders(req)
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 && resp.StatusCode != 404 {
		return fmt.Errorf("delete inbox %s: HTTP %d", id, resp.StatusCode)
	}
	return nil
}

// ── HTTP helpers ─────────────────────────────────────────────────────────────

func (c *Client) get(url string, v interface{}) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	c.setHeaders(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("GET %s: HTTP %d: %s", url, resp.StatusCode, string(body))
	}

	return json.NewDecoder(resp.Body).Decode(v)
}

func (c *Client) do(method, url string, body []byte) (*http.Response, error) {
	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)
	req.Header.Set("Prefer", "return=minimal")
	return c.http.Do(req)
}

func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("apikey", supabaseKey)
	req.Header.Set("Authorization", "Bearer "+supabaseKey)
	req.Header.Set("Content-Type", "application/json")
}

func (c *Client) isEmptyPatch(resp *http.Response) bool {
	body, _ := io.ReadAll(resp.Body)
	resp.Body = io.NopCloser(bytes.NewReader(body))
	return len(body) == 0 || string(body) == "[]" || bytes.Equal(body, []byte("[]"))
}

// ── Supabase types (JSON-tagged for REST API) ────────────────────────────────

type supabaseTask struct {
	ID            string  `json:"id"`
	Title         string  `json:"title"`
	Description   *string `json:"description"`
	Priority      string  `json:"priority"`
	Status        string  `json:"status"`
	ScheduledDate *string `json:"scheduled_date"`
	CreatedAt     string  `json:"created_at"`
	CompletedAt   *string `json:"completed_at"`
	UpdatedAt     string  `json:"updated_at"`
}

type supabaseNote struct {
	ID        string `json:"id"`
	TaskID    string `json:"task_id"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type supabaseInboxItem struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Description *string `json:"description"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

// ── Converters ───────────────────────────────────────────────────────────────

func (st supabaseTask) toModel() *model.Task {
	t := &model.Task{
		ID:        st.ID,
		Title:     st.Title,
		Priority:  model.Priority(st.Priority),
		Status:    model.TaskStatus(st.Status),
		UpdatedAt: st.UpdatedAt,
	}
	if st.Description != nil {
		t.Description = *st.Description
	}
	createdAt, _ := time.Parse(time.RFC3339, st.CreatedAt)
	t.CreatedAt = createdAt
	if st.ScheduledDate != nil {
		d, _ := time.Parse("2006-01-02", *st.ScheduledDate)
		t.ScheduledDate = &d
	}
	if st.CompletedAt != nil {
		ct, _ := time.Parse(time.RFC3339, *st.CompletedAt)
		t.CompletedAt = &ct
	}
	return t
}

func taskToSupabase(t *model.Task) supabaseTask {
	st := supabaseTask{
		ID:        t.ID,
		Title:     t.Title,
		Priority:  string(t.Priority),
		Status:    string(t.Status),
		UpdatedAt: t.UpdatedAt,
	}
	if t.Description != "" {
		st.Description = &t.Description
	}
	st.CreatedAt = t.CreatedAt.Format(time.RFC3339)
	if t.ScheduledDate != nil {
		s := t.ScheduledDate.Format("2006-01-02")
		st.ScheduledDate = &s
	}
	if t.CompletedAt != nil {
		s := t.CompletedAt.Format(time.RFC3339)
		st.CompletedAt = &s
	}
	return st
}

func (sn supabaseNote) toModel() *model.Note {
	n := &model.Note{
		ID:        sn.ID,
		TaskID:    sn.TaskID,
		Content:   sn.Content,
		UpdatedAt: sn.UpdatedAt,
	}
	createdAt, _ := time.Parse(time.RFC3339, sn.CreatedAt)
	n.CreatedAt = createdAt
	return n
}

func noteToSupabase(n *model.Note) supabaseNote {
	return supabaseNote{
		ID:        n.ID,
		TaskID:    n.TaskID,
		Content:   n.Content,
		CreatedAt: n.CreatedAt.Format(time.RFC3339),
		UpdatedAt: n.UpdatedAt,
	}
}

func (si supabaseInboxItem) toModel() *model.InboxItem {
	it := &model.InboxItem{
		ID:        si.ID,
		Title:     si.Title,
		UpdatedAt: si.UpdatedAt,
	}
	if si.Description != nil {
		it.Description = *si.Description
	}
	createdAt, _ := time.Parse(time.RFC3339, si.CreatedAt)
	it.CreatedAt = createdAt
	return it
}

func inboxToSupabase(it *model.InboxItem) supabaseInboxItem {
	si := supabaseInboxItem{
		ID:        it.ID,
		Title:     it.Title,
		UpdatedAt: it.UpdatedAt,
	}
	if it.Description != "" {
		si.Description = &it.Description
	}
	si.CreatedAt = it.CreatedAt.Format(time.RFC3339)
	return si
}
