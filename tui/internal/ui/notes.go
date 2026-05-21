package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/Rancy21/flowtask/internal/model"
	"github.com/Rancy21/flowtask/internal/repository"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ── Messages ──────────────────────────────────────────────────────────────────

type notesLoadedMsg struct{ notes []noteWithTask }

type notesErrMsg struct{ err error }

// noteWithTask pairs a note with its parent task title for display.
type noteWithTask struct {
	Note      model.Note
	TaskTitle string
}

// ── Key bindings ──────────────────────────────────────────────────────────────

type notesKeyMap struct {
	Up     key.Binding
	Down   key.Binding
	Delete key.Binding
	View   key.Binding
}

var notesKeys = notesKeyMap{
	Up:     key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
	Down:   key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
	Delete: key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "delete")),
	View:   key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "view task")),
}

// ── Model ─────────────────────────────────────────────────────────────────────

type NotesModel struct {
	taskRepo *repository.TaskRepo
	noteRepo *repository.NoteRepo
	notes    []noteWithTask
	cursor   int
	width    int
	height   int
	loading  bool
	err      error

	confirmDelete string
}

func NewNotesModel(taskRepo *repository.TaskRepo, noteRepo *repository.NoteRepo) NotesModel {
	return NotesModel{
		taskRepo: taskRepo,
		noteRepo: noteRepo,
		loading:  true,
	}
}

func (m *NotesModel) SetSize(w, h int) {
	m.width = w
	m.height = h
}

// ── Init ──────────────────────────────────────────────────────────────────────

func (m NotesModel) Init() tea.Cmd {
	return m.load()
}

func (m NotesModel) load() tea.Cmd {
	return func() tea.Msg {
		notes, err := m.noteRepo.GetAll()
		if err != nil {
			return notesErrMsg{err}
		}

		// Enrich each note with its parent task title.
		enriched := make([]noteWithTask, 0, len(notes))
		for _, n := range notes {
			nwt := noteWithTask{Note: n}
			task, err := m.taskRepo.GetByID(n.TaskID)
			if err == nil && task != nil {
				nwt.TaskTitle = task.Title
			} else {
				nwt.TaskTitle = "(deleted task)"
			}
			enriched = append(enriched, nwt)
		}
		return notesLoadedMsg{notes: enriched}
	}
}

// ── Update ────────────────────────────────────────────────────────────────────

func (m NotesModel) Update(msg tea.Msg) (NotesModel, tea.Cmd) {
	switch msg := msg.(type) {

	case RefreshMsg:
		m.loading = true
		return m, m.load()

	case notesLoadedMsg:
		m.loading = false
		m.err = nil
		m.notes = msg.notes
		if m.cursor >= len(m.notes) {
			m.cursor = max(0, len(m.notes)-1)
		}
		return m, nil

	case notesErrMsg:
		m.loading = false
		m.err = msg.err
		return m, nil

	case tea.KeyMsg:
		if m.confirmDelete != "" {
			return m.handleConfirm(msg)
		}
		return m.handleKey(msg)
	}

	return m, nil
}

func (m NotesModel) handleKey(msg tea.KeyMsg) (NotesModel, tea.Cmd) {
	if len(m.notes) == 0 {
		return m, nil
	}

	switch {
	case key.Matches(msg, notesKeys.Up):
		if m.cursor > 0 {
			m.cursor--
		}

	case key.Matches(msg, notesKeys.Down):
		if m.cursor < len(m.notes)-1 {
			m.cursor++
		}

	case key.Matches(msg, notesKeys.View):
		nwt := m.notes[m.cursor]
		return m, func() tea.Msg {
			return OpenEditorMsg{TaskID: nwt.Note.TaskID}
		}

	case key.Matches(msg, notesKeys.Delete):
		m.confirmDelete = m.notes[m.cursor].Note.ID
	}

	return m, nil
}

func (m NotesModel) handleConfirm(msg tea.KeyMsg) (NotesModel, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		id := m.confirmDelete
		m.confirmDelete = ""
		return m, func() tea.Msg {
			if err := m.noteRepo.Delete(id); err != nil {
				return notesErrMsg{err}
			}
			return RefreshMsg{}
		}
	default:
		m.confirmDelete = ""
	}
	return m, nil
}

// ── View ──────────────────────────────────────────────────────────────────────

func (m NotesModel) View() string {
	if m.loading {
		return StyleEmpty.Render("Loading...")
	}
	if m.err != nil {
		return StyleEmpty.Render(fmt.Sprintf("Error: %s", m.err))
	}

	var b strings.Builder

	b.WriteString(m.renderHeader())
	b.WriteString("\n")

	if m.confirmDelete != "" {
		b.WriteString(
			lipgloss.NewStyle().
				Foreground(colorDanger).
				Bold(true).
				Padding(0, 1).
				Render("Delete this note? y to confirm, any other key to cancel"),
		)
		b.WriteString("\n")
	}

	if len(m.notes) == 0 {
		b.WriteString(StyleEmpty.Render("No notes yet. Edit a task to add one."))
		return b.String()
	}

	for i, nwt := range m.notes {
		b.WriteString(m.renderNote(i, nwt))
		b.WriteString("\n")
	}

	return b.String()
}

func (m NotesModel) renderHeader() string {
	count := fmt.Sprintf("%d note", len(m.notes))
	if len(m.notes) != 1 {
		count += "s"
	}

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(colorText).
		Render("Notes")

	meta := lipgloss.NewStyle().
		Foreground(colorSubtext).
		Render(count)

	sep := lipgloss.NewStyle().
		Foreground(colorBorder).
		Render(strings.Repeat("─", m.width-2))

	return lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.JoinHorizontal(lipgloss.Bottom, title, "  ", meta),
		sep,
	)
}

func (m NotesModel) renderNote(i int, nwt noteWithTask) string {
	selected := i == m.cursor

	cursor := "  "
	if selected {
		cursor = lipgloss.NewStyle().Foreground(colorPrimary).Render("▶ ")
	}

	content := StyleNoteContent.Render(nwt.Note.Content)

	// Time ago
	ts := timeAgo(nwt.Note.CreatedAt)

	meta := StyleNoteMeta.Render(fmt.Sprintf("%s  ·  %s", ts, nwt.TaskTitle))

	return lipgloss.JoinVertical(lipgloss.Left,
		fmt.Sprintf("%s%s", cursor, content),
		meta,
	)
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func timeAgo(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		m := int(d.Minutes())
		if m == 1 {
			return "1m ago"
		}
		return fmt.Sprintf("%dm ago", m)
	case d < 24*time.Hour:
		h := int(d.Hours())
		if h == 1 {
			return "1h ago"
		}
		return fmt.Sprintf("%dh ago", h)
	default:
		days := int(d.Hours() / 24)
		if days == 1 {
			return "1d ago"
		}
		if days < 7 {
			return fmt.Sprintf("%dd ago", days)
		}
		return t.Format("Jan 2")
	}
}
