package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/Rancy21/flowtask/internal/model"
	"github.com/Rancy21/flowtask/internal/repository"
	"github.com/Rancy21/flowtask/internal/sync"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ── Messages ──────────────────────────────────────────────────────────────────

type weekLoadedMsg struct{ groups []dayGroup }

type weekErrMsg struct{ err error }

// ── Key bindings ──────────────────────────────────────────────────────────────

type weekKeyMap struct {
	Up     key.Binding
	Down   key.Binding
	Done   key.Binding
	Edit   key.Binding
	Delete key.Binding
}

var weekKeys = weekKeyMap{
	Up:     key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
	Down:   key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
	Done:   key.NewBinding(key.WithKeys("space", " "), key.WithHelp("space", "mark done")),
	Edit:   key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "edit")),
	Delete: key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "delete")),
}

// ── Day group ─────────────────────────────────────────────────────────────────

type dayGroup struct {
	Label string
	Today bool
	Tasks []model.Task
}

// ── Model ─────────────────────────────────────────────────────────────────────

type WeekModel struct {
	taskRepo *repository.TaskRepo
	sync     *sync.Client
	groups   []dayGroup
	cursor   int // flat index across all groups
	width    int
	height   int
	loading  bool
	err      error

	confirmDelete string
}

func NewWeekModel(taskRepo *repository.TaskRepo, sync *sync.Client) WeekModel {
	return WeekModel{
		taskRepo: taskRepo,
		sync:     sync,
		loading:  true,
	}
}

func (m *WeekModel) SetSize(w, h int) {
	m.width = w
	m.height = h
}

// totalTasks counts all tasks across groups.
func (m WeekModel) totalTasks() int {
	n := 0
	for _, g := range m.groups {
		n += len(g.Tasks)
	}
	return n
}

// taskAt returns the task at the flat cursor position and its group index.
func (m WeekModel) taskAt(idx int) (int, *model.Task) {
	remaining := idx
	for gi, g := range m.groups {
		if remaining < len(g.Tasks) {
			return gi, &g.Tasks[remaining]
		}
		remaining -= len(g.Tasks)
	}
	return -1, nil
}

// ── Init ──────────────────────────────────────────────────────────────────────

func (m WeekModel) Init() tea.Cmd {
	return m.load()
}

func (m WeekModel) load() tea.Cmd {
	return func() tea.Msg {
		tasks, err := m.taskRepo.GetWeek()
		if err != nil {
			return weekErrMsg{err}
		}
		groups := buildDayGroups(tasks)
		return weekLoadedMsg{groups}
	}
}

// ── Update ────────────────────────────────────────────────────────────────────

func (m WeekModel) Update(msg tea.Msg) (WeekModel, tea.Cmd) {
	switch msg := msg.(type) {

	case RefreshMsg:
		m.loading = true
		return m, m.load()

	case weekLoadedMsg:
		m.loading = false
		m.err = nil
		m.groups = msg.groups
		total := m.totalTasks()
		if m.cursor >= total {
			m.cursor = max(0, total-1)
		}
		return m, nil

	case weekErrMsg:
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

func (m WeekModel) handleKey(msg tea.KeyMsg) (WeekModel, tea.Cmd) {
	total := m.totalTasks()
	if total == 0 {
		return m, nil
	}

	switch {
	case key.Matches(msg, weekKeys.Up):
		if m.cursor > 0 {
			m.cursor--
		}

	case key.Matches(msg, weekKeys.Down):
		if m.cursor < total-1 {
			m.cursor++
		}

	case key.Matches(msg, weekKeys.Done):
		_, task := m.taskAt(m.cursor)
		if task == nil {
			return m, nil
		}
		return m, func() tea.Msg {
			if err := m.taskRepo.MarkDone(task.ID); err != nil {
				return weekErrMsg{err}
			}
			if updated, err := m.taskRepo.GetByID(task.ID); err == nil {
				m.sync.PushTask(updated)
			}
			return RefreshMsg{}
		}

	case key.Matches(msg, weekKeys.Edit):
		_, task := m.taskAt(m.cursor)
		if task == nil {
			return m, nil
		}
		return m, func() tea.Msg {
			return OpenEditorMsg{TaskID: task.ID}
		}

	case key.Matches(msg, weekKeys.Delete):
		_, task := m.taskAt(m.cursor)
		if task == nil {
			return m, nil
		}
		m.confirmDelete = task.ID
	}

	return m, nil
}

func (m WeekModel) handleConfirm(msg tea.KeyMsg) (WeekModel, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		id := m.confirmDelete
		m.confirmDelete = ""
		return m, func() tea.Msg {
			if err := m.taskRepo.Delete(id); err != nil {
				return weekErrMsg{err}
			}
			m.sync.DeleteTask(id)
			return RefreshMsg{}
		}
	default:
		m.confirmDelete = ""
	}
	return m, nil
}

// ── View ──────────────────────────────────────────────────────────────────────

func (m WeekModel) View() string {
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
				Render("Delete this task? y to confirm, any other key to cancel"),
		)
		b.WriteString("\n")
	}

	if m.totalTasks() == 0 {
		b.WriteString(StyleEmpty.Render("No tasks this week."))
		return b.String()
	}

	flatIdx := 0
	for gi, group := range m.groups {
		// Skip day groups with no tasks.
		if len(group.Tasks) == 0 {
			continue
		}
		b.WriteString(m.renderDayHeader(group))
		b.WriteString("\n")
		for ti, task := range group.Tasks {
			b.WriteString(m.renderTask(gi, ti, flatIdx, task))
			b.WriteString("\n")
			flatIdx++
		}
	}

	return b.String()
}

func (m WeekModel) renderHeader() string {
	now := time.Now()
	monday, _ := weekBoundDates(now)

	count := fmt.Sprintf("%d task", m.totalTasks())
	if m.totalTasks() != 1 {
		count += "s"
	}

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(colorText).
		Render(fmt.Sprintf("Week of %s", monday.Format("Jan 2")))

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

// weekBoundDates returns the Monday and Sunday time.Time values for the
// current week. Duplicates the repository helper so the UI stays layer-clean.
func weekBoundDates(t time.Time) (time.Time, time.Time) {
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7 // Sunday → 7
	}
	monday := t.AddDate(0, 0, -(weekday - 1))
	sunday := monday.AddDate(0, 0, 6)
	return monday, sunday
}

// buildDayGroups takes a flat task list (already sorted by date, then priority)
// from the repo and groups by day (Mon–Sun).
func buildDayGroups(tasks []model.Task) []dayGroup {
	now := time.Now()
	_, sunday := weekBoundDates(now)

	// Pre-build the 7 days Mon→Sun.
	type daySlot struct {
		label string
		date  time.Time
		dateS string // YYYY-MM-DD for comparison
	}
	slots := make([]daySlot, 7)
	monday := sunday.AddDate(0, 0, -6)
	for i := 0; i < 7; i++ {
		d := monday.AddDate(0, 0, i)
		slots[i] = daySlot{
			label: d.Format("Mon Jan 2"),
			date:  d,
			dateS: d.Format("2006-01-02"),
		}
	}

	// Index tasks by date.
	byDate := make(map[string][]model.Task)
	for _, t := range tasks {
		if t.ScheduledDate == nil {
			continue
		}
		ds := t.ScheduledDate.Format("2006-01-02")
		byDate[ds] = append(byDate[ds], t)
	}

	groups := make([]dayGroup, 7)
	for i, s := range slots {
		groups[i] = dayGroup{
			Label: s.label,
			Today: s.dateS == now.Format("2006-01-02"),
			Tasks: byDate[s.dateS],
		}
	}
	return groups
}

func (m WeekModel) renderDayHeader(group dayGroup) string {
	style := StyleSectionHeader
	if group.Today {
		style = StyleSectionHeaderToday
	}
	return style.Render(group.Label)
}

func (m WeekModel) renderTask(gi, ti, flatIdx int, task model.Task) string {
	selected := flatIdx == m.cursor
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
