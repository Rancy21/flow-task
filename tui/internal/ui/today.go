package ui

import (
	"fmt"
	"strings"

	"github.com/Rancy21/flowtask/internal/model"
	"github.com/Rancy21/flowtask/internal/repository"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ── Messages ──────────────────────────────────────────────────────────────────

// todayLoadedMsg carries tasks fetched from the DB.
type todayLoadedMsg struct{ tasks []model.Task }

// todayErrMsg carries a load error.
type todayErrMsg struct{ err error }

// ── Key bindings ──────────────────────────────────────────────────────────────

type todayKeyMap struct {
	Up     key.Binding
	Down   key.Binding
	Done   key.Binding
	Edit   key.Binding
	Delete key.Binding
}

var todayKeys = todayKeyMap{
	Up:     key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
	Down:   key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
	Done:   key.NewBinding(key.WithKeys(" "), key.WithHelp("space", "mark done")),
	Edit:   key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "edit")),
	Delete: key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "delete")),
}

// ── Model ─────────────────────────────────────────────────────────────────────

type TodayModel struct {
	taskRepo *repository.TaskRepo
	tasks    []model.Task
	cursor   int
	width    int
	height   int
	loading  bool
	err      error

	// confirmDelete holds the ID of the task pending deletion confirmation.
	confirmDelete string
}

func NewTodayModel(taskRepo *repository.TaskRepo) TodayModel {
	return TodayModel{
		taskRepo: taskRepo,
		loading:  true,
	}
}

func (m *TodayModel) SetSize(w, h int) {
	m.width = w
	m.height = h
}

// ── Init ──────────────────────────────────────────────────────────────────────

func (m TodayModel) Init() tea.Cmd {
	return m.load()
}

func (m TodayModel) load() tea.Cmd {
	return func() tea.Msg {
		tasks, err := m.taskRepo.GetToday()
		if err != nil {
			return todayErrMsg{err}
		}
		return todayLoadedMsg{tasks}
	}
}

// ── Update ────────────────────────────────────────────────────────────────────

func (m TodayModel) Update(msg tea.Msg) (TodayModel, tea.Cmd) {
	switch msg := msg.(type) {

	case RefreshMsg:
		m.loading = true
		return m, m.load()

	case todayLoadedMsg:
		m.loading = false
		m.err = nil
		m.tasks = msg.tasks
		// Clamp cursor in case the list shrank.
		if m.cursor >= len(m.tasks) {
			m.cursor = max(0, len(m.tasks)-1)
		}
		return m, nil

	case todayErrMsg:
		m.loading = false
		m.err = msg.err
		return m, nil

	case tea.KeyMsg:
		// Confirmation prompt intercepts all keys.
		if m.confirmDelete != "" {
			return m.handleConfirm(msg)
		}
		return m.handleKey(msg)
	}

	return m, nil
}

func (m TodayModel) handleKey(msg tea.KeyMsg) (TodayModel, tea.Cmd) {
	if len(m.tasks) == 0 {
		return m, nil
	}

	switch {
	case key.Matches(msg, todayKeys.Up):
		if m.cursor > 0 {
			m.cursor--
		}

	case key.Matches(msg, todayKeys.Down):
		if m.cursor < len(m.tasks)-1 {
			m.cursor++
		}

	case key.Matches(msg, todayKeys.Done):
		task := m.tasks[m.cursor]
		return m, func() tea.Msg {
			if err := m.taskRepo.MarkDone(task.ID); err != nil {
				return todayErrMsg{err}
			}
			return RefreshMsg{}
		}

	case key.Matches(msg, todayKeys.Edit):
		task := m.tasks[m.cursor]
		return m, func() tea.Msg {
			return OpenEditorMsg{TaskID: task.ID}
		}

	case key.Matches(msg, todayKeys.Delete):
		m.confirmDelete = m.tasks[m.cursor].ID
	}

	return m, nil
}

func (m TodayModel) handleConfirm(msg tea.KeyMsg) (TodayModel, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		id := m.confirmDelete
		m.confirmDelete = ""
		return m, func() tea.Msg {
			if err := m.taskRepo.Delete(id); err != nil {
				return todayErrMsg{err}
			}
			return RefreshMsg{}
		}
	default:
		// Any other key cancels.
		m.confirmDelete = ""
	}
	return m, nil
}

// ── View ──────────────────────────────────────────────────────────────────────

func (m TodayModel) View() string {
	if m.loading {
		return StyleEmpty.Render("Loading...")
	}
	if m.err != nil {
		return StyleEmpty.Render(fmt.Sprintf("Error: %s", m.err))
	}

	var b strings.Builder

	// ── Header ────────────────────────────────────────────────────────────────
	b.WriteString(m.renderHeader())
	b.WriteString("\n")

	// ── Confirm delete overlay ────────────────────────────────────────────────
	if m.confirmDelete != "" {
		b.WriteString(
			lipgloss.NewStyle().
				Foreground(colorDanger).
				Bold(true).
				Padding(0, 1).
				Render("Delete this task? y to confirm, any other key to cancel"),
		)
		b.WriteString("\n")
	}

	// ── Empty state ───────────────────────────────────────────────────────────
	if len(m.tasks) == 0 {
		b.WriteString(StyleEmpty.Render("No tasks for today."))
		return b.String()
	}

	// ── Task list ─────────────────────────────────────────────────────────────
	for i, task := range m.tasks {
		b.WriteString(m.renderTask(i, task))
		b.WriteString("\n")
	}

	return b.String()
}

func (m TodayModel) renderHeader() string {
	count := fmt.Sprintf("%d task", len(m.tasks))
	if len(m.tasks) != 1 {
		count += "s"
	}

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(colorText).
		Render("Today")

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

func (m TodayModel) renderTask(i int, task model.Task) string {
	selected := i == m.cursor
	badge := PriorityStyle(string(task.Priority)).Render(string(task.Priority))

	var titleStyle lipgloss.Style
	if selected {
		titleStyle = StyleTaskSelected
	} else {
		titleStyle = StyleTaskNormal
	}

	cursor := "  "
	if selected {
		cursor = lipgloss.NewStyle().Foreground(colorPrimary).Render("▶ ")
	}

	title := titleStyle.Render(task.Title)
	row := fmt.Sprintf("%s%s %s", cursor, badge, title)

	if task.Description != "" && selected {
		desc := StyleTaskDescription.Render(task.Description)
		return lipgloss.JoinVertical(lipgloss.Left, row, desc)
	}

	return row
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
