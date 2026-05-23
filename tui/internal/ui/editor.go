package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/Rancy21/flowtask/internal/model"
	"github.com/Rancy21/flowtask/internal/repository"
	"github.com/Rancy21/flowtask/internal/sync"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ── Messages ──────────────────────────────────────────────────────────────────

type EditorDoneMsg struct{}
type EditorCancelMsg struct{}
type taskLoadedMsg struct{ task *model.Task }
type editorNotesLoadedMsg struct{ notes []model.Note }
type editorErrMsg struct{ err error }

// ── Field indices ─────────────────────────────────────────────────────────────

type editorField int

const (
	fieldTitle editorField = iota
	fieldDescription
	fieldPriority
	fieldDate
	fieldNotes
	fieldActions
	fieldCount = 6
)

// ── Key bindings ──────────────────────────────────────────────────────────────

type editorKeyMap struct {
	NextField  key.Binding
	PrevField  key.Binding
	Save       key.Binding
	Cancel     key.Binding
	CycleRight key.Binding
	CycleLeft  key.Binding
	AddNote    key.Binding
}

var editorKeys = editorKeyMap{
	NextField:  key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next field")),
	PrevField:  key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("shift+tab", "prev field")),
	Save:       key.NewBinding(key.WithKeys("ctrl+s"), key.WithHelp("ctrl+s", "save")),
	Cancel:     key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel")),
	CycleRight: key.NewBinding(key.WithKeys("right", "l"), key.WithHelp("→/l", "next")),
	CycleLeft:  key.NewBinding(key.WithKeys("left", "h"), key.WithHelp("←/h", "prev")),
	AddNote:    key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "add note")),
}

// ── Styles (editor-local) ─────────────────────────────────────────────────────

var (
	editorContainer = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder).
			Padding(1, 3)

	editorHeader = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPrimary).
			MarginBottom(1)

	fieldLabel = lipgloss.NewStyle().
			Foreground(colorSubtext).
			MarginBottom(0)

	fieldLabelActive = lipgloss.NewStyle().
				Foreground(colorPrimary).
				Bold(true).
				MarginBottom(0)

	fieldBox = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(colorBorder).
			Padding(0, 1)

	fieldBoxActive = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(colorPrimary).
			Padding(0, 1)

	priorityActive = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true).
			Padding(0, 2)

	priorityInactive = lipgloss.NewStyle().
				Foreground(colorSubtext).
				Padding(0, 2)

	dateOptionActive = lipgloss.NewStyle().
				Foreground(colorPrimary).
				Bold(true).
				Underline(true)

	dateOptionInactive = lipgloss.NewStyle().
				Foreground(colorSubtext)

	editorFooter = lipgloss.NewStyle().
			Foreground(colorSubtext).
			MarginTop(1)

	saveBtn = lipgloss.NewStyle().
		Foreground(colorSuccess).
		Bold(true).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorBorder).
		Padding(0, 2)

	saveBtnActive = lipgloss.NewStyle().
			Foreground(colorDark).
			Bold(true).
			Background(colorPrimary).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorPrimary).
			Padding(0, 2)

	divider = lipgloss.NewStyle().
		Foreground(colorBorder)
)

// ── Model ─────────────────────────────────────────────────────────────────────

type EditorModel struct {
	taskRepo *repository.TaskRepo
	noteRepo *repository.NoteRepo
	sync     *sync.Client

	taskID string
	mode   string // "new" | "edit"

	titleInput textinput.Model
	descInput  textarea.Model
	priority   model.Priority
	dateType   string // "none" | "today" | "tomorrow" | "pick"
	dateInput  textinput.Model
	dateValue  *time.Time

	notes     []model.Note
	noteInput textinput.Model

	activeField editorField
	width       int
	height      int
	innerWidth  int // usable width inside the box

	err    error
	saving bool
}

func NewEditorModel(taskRepo *repository.TaskRepo, noteRepo *repository.NoteRepo, sync *sync.Client) EditorModel {
	ti := textinput.New()
	ti.Placeholder = "What needs to be done?"
	ti.CharLimit = 200
	ti.Focus()

	di := textarea.New()
	di.Placeholder = "Add details..."
	di.CharLimit = 2000
	di.SetHeight(5)
	di.ShowLineNumbers = true

	dd := textinput.New()
	dd.Placeholder = "YYYY-MM-DD"
	dd.CharLimit = 10

	ni := textinput.New()
	ni.Placeholder = "Reflection note..."
	ni.CharLimit = 500

	return EditorModel{
		taskRepo:    taskRepo,
		noteRepo:    noteRepo,
		sync:        sync,
		mode:        "new",
		titleInput:  ti,
		descInput:   di,
		priority:    model.P3,
		dateType:    "none",
		dateInput:   dd,
		noteInput:   ni,
		activeField: fieldTitle,
	}
}

func (m *EditorModel) SetSize(w, h int) {
	m.width = w
	m.height = h
	// Box uses Padding(1,3) + border = 2+6+2 = 10 horizontal overhead
	// Cap form at 72 chars for readability
	formW := min(72, w-6)
	m.innerWidth = formW - 8 // subtract box padding + border
	m.titleInput.Width = m.innerWidth
	m.descInput.SetWidth(m.innerWidth)
	m.dateInput.Width = 14
	m.noteInput.Width = m.innerWidth
}

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
		if msg.task == nil {
			m.err = fmt.Errorf("task not found")
			return m, nil
		}
		t := msg.task
		m.titleInput.SetValue(t.Title)
		m.descInput.SetValue(t.Description)
		m.priority = t.Priority
		if t.ScheduledDate != nil {
			m.dateType = "pick"
			m.dateInput.SetValue(t.ScheduledDate.Format("2006-01-02"))
			m.dateValue = t.ScheduledDate
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
		if key.Matches(msg, editorKeys.Cancel) {
			return m, func() tea.Msg { return EditorCancelMsg{} }
		}
		if key.Matches(msg, editorKeys.Save) {
			return m, m.save()
		}

		// Tab only navigates when NOT in the description field.
		if key.Matches(msg, editorKeys.NextField) && m.activeField != fieldDescription {
			m.nextField()
			return m, nil
		}
		if key.Matches(msg, editorKeys.PrevField) && m.activeField != fieldDescription {
			m.prevField()
			return m, nil
		}

		switch m.activeField {
		case fieldTitle:
			var cmd tea.Cmd
			m.titleInput, cmd = m.titleInput.Update(msg)
			cmds = append(cmds, cmd)

		case fieldDescription:
			// ctrl+tab exits the description field
			if msg.String() == "ctrl+down" || msg.String() == "ctrl+j" {
				m.nextField()
				return m, nil
			}
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
			if key.Matches(msg, editorKeys.CycleRight) {
				m.cycleDateType(1)
			}
			if key.Matches(msg, editorKeys.CycleLeft) {
				m.cycleDateType(-1)
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
						n, err := m.noteRepo.Create(m.taskID, content)
						if err != nil {
							return editorErrMsg{err}
						}
						_ = m.sync.PushNote(n)
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

		case fieldActions:
			if msg.String() == "enter" {
				return m, m.save()
			}
		}
	}

	return m, tea.Batch(cmds...)
}

// ── View ──────────────────────────────────────────────────────────────────────

func (m EditorModel) View() string {
	formW := min(72, m.width-6)
	innerW := formW - 8

	var sections []string

	// ── Header ────────────────────────────────────────────────────────────────
	label := "NEW TASK"
	if m.mode == "edit" {
		label = "EDIT TASK"
	}
	sections = append(sections, editorHeader.Render(label))
	sections = append(sections, divider.Render(strings.Repeat("─", innerW)))

	// ── Error ─────────────────────────────────────────────────────────────────
	if m.err != nil {
		sections = append(sections,
			lipgloss.NewStyle().Foreground(colorDanger).Render("✖  "+m.err.Error()),
		)
	}

	// ── Title ─────────────────────────────────────────────────────────────────
	sections = append(sections, m.renderTextSection(
		"TITLE", m.titleInput.View(), fieldTitle, false,
	))

	// ── Description ───────────────────────────────────────────────────────────
	sections = append(sections, m.renderTextSection(
		"DESCRIPTION  (ctrl+j to continue)", m.descInput.View(), fieldDescription, true,
	))

	// ── Priority ──────────────────────────────────────────────────────────────
	sections = append(sections, m.renderInlineSection(
		"PRIORITY", m.renderPriority(), fieldPriority,
	))

	// ── Schedule ──────────────────────────────────────────────────────────────
	sections = append(sections, m.renderInlineSection(
		"SCHEDULE", m.renderDate(), fieldDate,
	))

	// ── Notes ─────────────────────────────────────────────────────────────────
	if m.mode == "edit" {
		sections = append(sections, m.renderTextSection(
			"NOTES", m.renderNotes(innerW), fieldNotes, false,
		))
	}

	// ── Actions ───────────────────────────────────────────────────────────────
	sections = append(sections, "")
	sections = append(sections, m.renderActions())

	// ── Footer ────────────────────────────────────────────────────────────────
	sections = append(sections, divider.Render(strings.Repeat("─", innerW)))
	sections = append(sections, editorFooter.Render(
		"ctrl+s save  ·  esc cancel  ·  tab/shift+tab navigate  ·  ←→ change values",
	))

	form := lipgloss.NewStyle().Width(formW).Render(
		lipgloss.JoinVertical(lipgloss.Left, sections...),
	)

	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		editorContainer.Render(form),
	)
}

// renderTextSection renders a label above a full-width input/textarea.
func (m EditorModel) renderTextSection(label, content string, f editorField, isTextarea bool) string {
	active := m.activeField == f

	lStyle := fieldLabel
	bStyle := fieldBox
	if active {
		lStyle = fieldLabelActive
		bStyle = fieldBoxActive
	}

	_ = isTextarea // reserved for future per-type tweaks
	return lipgloss.JoinVertical(lipgloss.Left,
		lStyle.Render(label),
		bStyle.Render(content),
	)
}

// renderInlineSection renders a label on the left, value on the right.
func (m EditorModel) renderInlineSection(label, content string, f editorField) string {
	active := m.activeField == f

	lStyle := fieldLabel.Copy().Width(12)
	if active {
		lStyle = fieldLabelActive.Copy().Width(12)
	}

	return lipgloss.JoinHorizontal(lipgloss.Top,
		lStyle.Render(label),
		"  ",
		content,
	)
}

func (m EditorModel) renderPriority() string {
	type opt struct {
		p   model.Priority
		bg  lipgloss.Color
	}
	opts := []opt{
		{model.P1, colorDanger},
		{model.P2, colorWarning},
		{model.P3, colorMuted},
	}

	parts := make([]string, len(opts))
	for i, o := range opts {
		if o.p == m.priority {
			parts[i] = priorityActive.Copy().Background(o.bg).Render(string(o.p))
		} else {
			parts[i] = priorityInactive.Render(string(o.p))
		}
	}
	return lipgloss.JoinHorizontal(lipgloss.Center, parts...)
}

func (m EditorModel) renderDate() string {
	types := []struct{ key, label string }{
		{"none", "No date"},
		{"today", "Today"},
		{"tomorrow", "Tomorrow"},
		{"pick", "Pick..."},
	}

	parts := make([]string, len(types))
	for i, t := range types {
		if t.key == m.dateType {
			parts[i] = dateOptionActive.Render(t.label)
		} else {
			parts[i] = dateOptionInactive.Render(t.label)
		}
	}

	row := lipgloss.JoinHorizontal(lipgloss.Center, strings.Join(parts, "  "))
	if m.dateType == "pick" {
		return lipgloss.JoinVertical(lipgloss.Left, row, m.dateInput.View())
	}
	return row
}

func (m EditorModel) renderNotes(innerW int) string {
	var lines []string

	for _, n := range m.notes {
		ts := lipgloss.NewStyle().Foreground(colorSubtext).Italic(true).Render(n.CreatedAt.Format("Jan 2 15:04"))
		lines = append(lines, fmt.Sprintf("%s  %s", ts, n.Content))
	}

	if m.activeField == fieldNotes {
		lines = append(lines, m.noteInput.View())
	} else {
		lines = append(lines,
			lipgloss.NewStyle().Foreground(colorSubtext).Faint(true).Render("↵ to add a note"),
		)
	}

	return strings.Join(lines, "\n")
}

func (m EditorModel) renderActions() string {
	var s lipgloss.Style
	if m.activeField == fieldActions {
		s = saveBtnActive
	} else {
		s = saveBtn
	}
	return lipgloss.JoinHorizontal(lipgloss.Center,
		s.Render("Save"),
		"   ",
		lipgloss.NewStyle().Foreground(colorSubtext).Render("esc to cancel"),
	)
}

// ── Navigation ────────────────────────────────────────────────────────────────

func (m *EditorModel) nextField() {
	m.activeField = (m.activeField + 1) % fieldCount
	m.updateFocus()
}

func (m *EditorModel) prevField() {
	if m.activeField == 0 {
		m.activeField = fieldCount - 1
	} else {
		m.activeField--
	}
	m.updateFocus()
}

func (m *EditorModel) updateFocus() {
	m.titleInput.Blur()
	m.descInput.Blur()
	m.dateInput.Blur()
	m.noteInput.Blur()
	switch m.activeField {
	case fieldTitle:
		m.titleInput.Focus()
	case fieldDescription:
		m.descInput.Focus()
	case fieldDate:
		m.dateInput.Focus()
	case fieldNotes:
		m.noteInput.Focus()
	}
}

// ── Cycling ───────────────────────────────────────────────────────────────────

func (m *EditorModel) cyclePriority(dir int) {
	order := []model.Priority{model.P1, model.P2, model.P3}
	for i, p := range order {
		if p == m.priority {
			m.priority = order[(i+dir+len(order))%len(order)]
			return
		}
	}
}

func (m *EditorModel) cycleDateType(dir int) {
	types := []string{"none", "today", "tomorrow", "pick"}
	for i, t := range types {
		if t == m.dateType {
			m.dateType = types[(i+dir+len(types))%len(types)]
			return
		}
	}
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
				m.err = fmt.Errorf("invalid date — use YYYY-MM-DD")
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
			_ = m.sync.PushTask(task)
			return EditorDoneMsg{}
		}
	}

	return func() tea.Msg {
		_, err := m.taskRepo.Create(task)
		if err != nil {
			return editorErrMsg{err}
		}
		_ = m.sync.PushTask(task)
		return EditorDoneMsg{}
	}
}
