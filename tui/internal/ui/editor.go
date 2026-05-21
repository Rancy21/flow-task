package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/Rancy21/flowtask/internal/model"
	"github.com/Rancy21/flowtask/internal/repository"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ── Messages ──────────────────────────────────────────────────────────────────

// EditorDoneMsg is emitted when the editor saves successfully.
type EditorDoneMsg struct{}

// EditorCancelMsg is emitted when the editor is dismissed without saving.
type EditorCancelMsg struct{}

// taskLoadedMsg carries an existing task for editing.
type taskLoadedMsg struct{ task *model.Task }

// editorNotesLoadedMsg carries notes for the task being edited.
type editorNotesLoadedMsg struct{ notes []model.Note }

// editorErrMsg carries a save/load error.
type editorErrMsg struct{ err error }

// ── Field indices ─────────────────────────────────────────────────────────────

const (
	fieldTitle editorField = iota
	fieldDescription
	fieldPriority
	fieldDate
	fieldNotes
	fieldActions
)

type editorField int

// ── Key bindings ──────────────────────────────────────────────────────────────

type editorKeyMap struct {
	Tab        key.Binding
	ShiftTab   key.Binding
	Save       key.Binding
	Cancel     key.Binding
	CycleRight key.Binding
	CycleLeft  key.Binding
	AddNote    key.Binding
}

var editorKeys = editorKeyMap{
	Tab:        key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next field")),
	ShiftTab:   key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("shift+tab", "prev field")),
	Save:       key.NewBinding(key.WithKeys("ctrl+s"), key.WithHelp("ctrl+s", "save")),
	Cancel:     key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel")),
	CycleRight: key.NewBinding(key.WithKeys("right", "l"), key.WithHelp("→/l", "next option")),
	CycleLeft:  key.NewBinding(key.WithKeys("left", "h"), key.WithHelp("←/h", "prev option")),
	AddNote:    key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "add note")),
}

// ── Model ─────────────────────────────────────────────────────────────────────

type EditorModel struct {
	taskRepo *repository.TaskRepo
	noteRepo *repository.NoteRepo

	// Editing state
	taskID string // empty = new task
	mode   string // "new" or "edit"

	// Inputs
	titleInput textinput.Model
	descInput  textarea.Model
	priority   model.Priority
	dateType   string // "none", "today", "tomorrow", "pick"
	dateInput  textinput.Model
	dateValue  *time.Time

	// Notes
	notes     []model.Note
	noteInput textinput.Model

	// Navigation
	activeField editorField
	width       int
	height      int

	err    error
	saving bool
}

func NewEditorModel(taskRepo *repository.TaskRepo, noteRepo *repository.NoteRepo) EditorModel {
	ti := textinput.New()
	ti.Placeholder = "What needs to be done?"
	ti.CharLimit = 200
	ti.Focus()

	di := textarea.New()
	di.Placeholder = "Details (optional)"
	di.CharLimit = 2000
	di.SetHeight(4)

	diDate := textinput.New()
	diDate.Placeholder = "YYYY-MM-DD"
	diDate.CharLimit = 10

	ni := textinput.New()
	ni.Placeholder = "Add a reflection note..."
	ni.CharLimit = 500

	return EditorModel{
		taskRepo:    taskRepo,
		noteRepo:    noteRepo,
		mode:        "new",
		titleInput:  ti,
		descInput:   di,
		priority:    model.P3,
		dateType:    "none",
		dateInput:   diDate,
		noteInput:   ni,
		activeField: fieldTitle,
	}
}

func (m *EditorModel) SetSize(w, h int) {
	m.width = w
	m.height = h
	m.titleInput.Width = max(20, w-12)
	m.descInput.SetWidth(max(20, w-12))
	m.dateInput.Width = 12
	m.noteInput.Width = max(20, w-12)
}

// LoadTask returns a command that fetches a task and its notes for editing.
func (m *EditorModel) LoadTask(taskID string) tea.Cmd {
	m.taskID = taskID
	m.mode = "edit"
	return tea.Batch(
		func() tea.Msg {
			task, err := m.taskRepo.GetByID(taskID)
			if err != nil {
				return editorErrMsg{err}
			}
			return taskLoadedMsg{task}
		},
		func() tea.Msg {
			notes, err := m.noteRepo.GetByTask(taskID)
			if err != nil {
				return editorErrMsg{err}
			}
			return editorNotesLoadedMsg{notes}
		},
	)
}

// ── Init ──────────────────────────────────────────────────────────────────────

func (m EditorModel) Init() tea.Cmd {
	return textinput.Blink
}

// ── Update ────────────────────────────────────────────────────────────────────

func (m EditorModel) Update(msg tea.Msg) (EditorModel, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	case taskLoadedMsg:
		t := msg.task
		if t == nil {
			m.err = fmt.Errorf("task not found")
			return m, nil
		}
		m.titleInput.SetValue(t.Title)
		m.descInput.SetValue(t.Description)
		m.priority = t.Priority
		if t.ScheduledDate != nil {
			m.dateType = "pick"
			m.dateInput.SetValue(t.ScheduledDate.Format("2006-01-02"))
			m.dateValue = t.ScheduledDate
		} else {
			m.dateType = "none"
		}
		return m, nil

	case editorNotesLoadedMsg:
		m.notes = msg.notes
		return m, nil

	case editorErrMsg:
		m.err = msg.err
		m.saving = false
		return m, nil

	case tea.KeyMsg:
		// Global save/cancel regardless of active field.
		if key.Matches(msg, editorKeys.Cancel) {
			return m, func() tea.Msg { return EditorCancelMsg{} }
		}
		if key.Matches(msg, editorKeys.Save) {
			return m, m.save()
		}

		// Tab navigation.
		if key.Matches(msg, editorKeys.Tab) {
			m.nextField()
			return m, nil
		}
		if key.Matches(msg, editorKeys.ShiftTab) {
			m.prevField()
			return m, nil
		}

		// Route to active field.
		switch m.activeField {

		case fieldTitle:
			var cmd tea.Cmd
			m.titleInput, cmd = m.titleInput.Update(msg)
			cmds = append(cmds, cmd)

		case fieldDescription:
			var cmd tea.Cmd
			m.descInput, cmd = m.descInput.Update(msg)
			cmds = append(cmds, cmd)

		case fieldPriority:
			if key.Matches(msg, editorKeys.CycleRight) {
				m.cyclePriority(1)
			}
			if key.Matches(msg, editorKeys.CycleLeft) {
				m.cyclePriority(-1)
			}

		case fieldDate:
			if key.Matches(msg, editorKeys.CycleRight) || key.Matches(msg, editorKeys.CycleLeft) {
				if key.Matches(msg, editorKeys.CycleRight) {
					m.cycleDateType(1)
				} else {
					m.cycleDateType(-1)
				}
			}
			if m.dateType == "pick" {
				var cmd tea.Cmd
				m.dateInput, cmd = m.dateInput.Update(msg)
				cmds = append(cmds, cmd)
			}

		case fieldNotes:
			if key.Matches(msg, editorKeys.AddNote) {
				content := strings.TrimSpace(m.noteInput.Value())
				if content != "" && m.taskID != "" {
					m.noteInput.SetValue("")
					return m, func() tea.Msg {
						_, err := m.noteRepo.Create(m.taskID, content)
						if err != nil {
							return editorErrMsg{err}
						}
						// Reload notes after creation.
						notes, err := m.noteRepo.GetByTask(m.taskID)
						if err != nil {
							return editorErrMsg{err}
						}
						return editorNotesLoadedMsg{notes}
					}
				}
			} else {
				var cmd tea.Cmd
				m.noteInput, cmd = m.noteInput.Update(msg)
				cmds = append(cmds, cmd)
			}
		}

		// Enter on actions = save.
		if m.activeField == fieldActions && msg.String() == "enter" {
			return m, m.save()
		}
	}

	return m, tea.Batch(cmds...)
}

// ── View ──────────────────────────────────────────────────────────────────────

func (m EditorModel) View() string {
	formWidth := min(60, m.width-4)

	headerLabel := "CREATE TASK"
	if m.mode == "edit" {
		headerLabel = "EDIT TASK"
	}
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(colorPrimary).
		Padding(0, 1).
		MarginBottom(1).
		Render(headerLabel)

	var b strings.Builder
	b.WriteString(header)
	b.WriteString("\n")

	if m.err != nil {
		b.WriteString(
			lipgloss.NewStyle().
				Foreground(colorDanger).
				Render(fmt.Sprintf("Error: %s", m.err)),
		)
		b.WriteString("\n\n")
	}
	if m.saving {
		b.WriteString(
			lipgloss.NewStyle().
				Foreground(colorMuted).
				Render("Saving..."),
		)
		b.WriteString("\n\n")
	}

	// ── Title ──────────────────────────────────────────────────────────────
	b.WriteString(m.renderField("Title", m.titleInput.View(), fieldTitle))
	b.WriteString("\n")

	// ── Description ────────────────────────────────────────────────────────
	b.WriteString(m.renderField("Description", m.descInput.View(), fieldDescription))
	b.WriteString("\n")

	// ── Priority ───────────────────────────────────────────────────────────
	pv := m.renderPriorityPicker()
	b.WriteString(m.renderField("Priority", pv, fieldPriority))
	b.WriteString("\n")

	// ── Scheduled date ─────────────────────────────────────────────────────
	dv := m.renderDatePicker()
	b.WriteString(m.renderField("Schedule", dv, fieldDate))
	b.WriteString("\n")

	// ── Notes ──────────────────────────────────────────────────────────────
	b.WriteString(m.renderField("Notes", m.renderNotes(), fieldNotes))
	b.WriteString("\n")

	// ── Actions ────────────────────────────────────────────────────────────
	av := m.renderActions()
	b.WriteString(m.renderField("", av, fieldActions))
	b.WriteString("\n")

	// Footer hints
	footer := lipgloss.NewStyle().
		Foreground(colorSubtext).
		MarginTop(1).
		Render("ctrl+s save  ·  esc cancel  ·  tab next field  ·  ← → change values")

	content := lipgloss.NewStyle().
		Width(formWidth).
		Render(lipgloss.JoinVertical(lipgloss.Left, b.String(), footer))

	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		StyleBox.Render(content),
	)
}

func (m EditorModel) renderField(label, content string, f editorField) string {
	var labelStyle func(...string) string
	if m.activeField == f {
		labelStyle = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Bold(true).
			Width(12).
			Render
	} else {
		labelStyle = lipgloss.NewStyle().
			Foreground(colorSubtext).
			Width(12).
			Render
	}

	if label == "" {
		return fmt.Sprintf("  %s", content)
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, labelStyle(label+":"), "  ", content)
}

func (m EditorModel) renderPriorityPicker() string {
	opts := []model.Priority{model.P3, model.P2, model.P1}
	segments := make([]string, len(opts))
	for i, p := range opts {
		if p == m.priority {
			segments[i] = PriorityStyle(string(p)).Render(string(p))
		} else {
			segments[i] = lipgloss.NewStyle().
				Foreground(colorMuted).
				Faint(true).
				Render(string(p))
		}
	}
	return strings.Join(segments, "  ")
}

func (m EditorModel) renderDatePicker() string {
	types := []struct {
		key   string
		label string
	}{
		{"none", "No date"},
		{"today", "Today"},
		{"tomorrow", "Tomorrow"},
		{"pick", "Pick..."},
	}

	activeStyle := lipgloss.NewStyle().
		Foreground(colorPrimary).
		Bold(true)

	inactiveStyle := lipgloss.NewStyle().
		Foreground(colorMuted)

	segments := make([]string, len(types))
	for i, t := range types {
		if t.key == m.dateType {
			segments[i] = activeStyle.Render(t.label)
		} else {
			segments[i] = inactiveStyle.Render(t.label)
		}
	}

	result := strings.Join(segments, "  ")

	if m.dateType == "pick" {
		result += "\n" + m.dateInput.View()
	}
	return result
}

func (m EditorModel) renderNotes() string {
	var parts []string

	// Existing notes.
	if len(m.notes) > 0 {
		noteStyle := lipgloss.NewStyle().
			Foreground(colorSubtext).
			Italic(true)

		for _, n := range m.notes {
			ts := n.CreatedAt.Format("Jan 2 15:04")
			line := fmt.Sprintf("  %s  %s", noteStyle.Render(ts), n.Content)
			parts = append(parts, line)
		}
	}

	// New note input (only when editing an existing task).
	if m.mode == "edit" {
		if m.activeField == fieldNotes {
			parts = append(parts, m.noteInput.View())
		} else {
			placeholder := lipgloss.NewStyle().
				Foreground(colorSubtext).
				Faint(true).
				Render("enter to add a note...")
			parts = append(parts, placeholder)
		}
	}

	if len(parts) == 0 {
		return lipgloss.NewStyle().
			Foreground(colorSubtext).
			Faint(true).
			Render("(notes can be added after creating the task)")
	}
	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

func (m EditorModel) renderActions() string {
	saveStyle := lipgloss.NewStyle().
		Foreground(colorSuccess).
		Bold(true)

	cancelStyle := lipgloss.NewStyle().
		Foreground(colorSubtext)

	if m.activeField == fieldActions {
		saveStyle = saveStyle.
			Background(colorPrimary).
			Foreground(lipgloss.Color("#1A1A2E")).
			Padding(0, 1)
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Center,
		saveStyle.Render("[ Save ]"),
		"  ",
		cancelStyle.Render("esc to cancel"),
	)
}

// ── Navigation ────────────────────────────────────────────────────────────────

func (m *EditorModel) nextField() {
	m.activeField = (m.activeField + 1) % 6
	m.updateFocus()
}

func (m *EditorModel) prevField() {
	if m.activeField == 0 {
		m.activeField = 5
	} else {
		m.activeField--
	}
	m.updateFocus()
}

func (m *EditorModel) updateFocus() {
	m.titleInput.Blur()
	m.descInput.Blur()
	m.noteInput.Blur()

	switch m.activeField {
	case fieldTitle:
		m.titleInput.Focus()
	case fieldDescription:
		m.descInput.Focus()
	case fieldNotes:
		m.noteInput.Focus()
	}
}

// ── Value cycling ─────────────────────────────────────────────────────────────

func (m *EditorModel) cyclePriority(dir int) {
	order := []model.Priority{model.P1, model.P2, model.P3}
	idx := -1
	for i, p := range order {
		if p == m.priority {
			idx = i
			break
		}
	}
	idx = (idx + dir + len(order)) % len(order)
	m.priority = order[idx]
}

func (m *EditorModel) cycleDateType(dir int) {
	types := []string{"none", "today", "tomorrow", "pick"}
	idx := -1
	for i, t := range types {
		if t == m.dateType {
			idx = i
			break
		}
	}
	idx = (idx + dir + len(types)) % len(types)
	m.dateType = types[idx]
}

// ── Save ──────────────────────────────────────────────────────────────────────

func (m EditorModel) save() tea.Cmd {
	title := strings.TrimSpace(m.titleInput.Value())
	if title == "" {
		m.err = fmt.Errorf("title is required")
		return nil
	}

	task := &model.Task{
		Title:       title,
		Description: strings.TrimSpace(m.descInput.Value()),
		Priority:    m.priority,
	}

	// Resolve scheduled date.
	switch m.dateType {
	case "today":
		now := time.Now()
		d := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
		task.ScheduledDate = &d
	case "tomorrow":
		now := time.Now().AddDate(0, 0, 1)
		d := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
		task.ScheduledDate = &d
	case "pick":
		val := strings.TrimSpace(m.dateInput.Value())
		if val != "" {
			parsed, err := time.Parse("2006-01-02", val)
			if err != nil {
				m.err = fmt.Errorf("invalid date format (use YYYY-MM-DD)")
				return nil
			}
			task.ScheduledDate = &parsed
		}
	}

	m.saving = true
	m.err = nil

	if m.mode == "edit" {
		task.ID = m.taskID
		return func() tea.Msg {
			if err := m.taskRepo.Update(task); err != nil {
				return editorErrMsg{err}
			}
			return EditorDoneMsg{}
		}
	}

	return func() tea.Msg {
		_, err := m.taskRepo.Create(task)
		if err != nil {
			return editorErrMsg{err}
		}
		return EditorDoneMsg{}
	}
}
