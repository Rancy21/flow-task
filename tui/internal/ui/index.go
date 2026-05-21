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

type inboxLoadedMsg struct{ items []model.InboxItem }
type inboxErrMsg struct{ err error }

// ── Key bindings ──────────────────────────────────────────────────────────────

type inboxKeyMap struct {
	Up     key.Binding
	Down   key.Binding
	New    key.Binding
	Edit   key.Binding
	Delete key.Binding
}

var inboxKeys = inboxKeyMap{
	Up:     key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
	Down:   key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
	New:    key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "capture")),
	Edit:   key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "edit")),
	Delete: key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "delete")),
}

// ── Model ─────────────────────────────────────────────────────────────────────

type InboxModel struct {
	repo    *repository.InboxRepo
	items   []model.InboxItem
	cursor  int
	width   int
	height  int
	loading bool
	err     error

	confirmDelete string
}

func NewInboxModel(repo *repository.InboxRepo) InboxModel {
	return InboxModel{
		repo:    repo,
		loading: true,
	}
}

func (m *InboxModel) SetSize(w, h int) {
	m.width = w
	m.height = h
}

// ── Init ──────────────────────────────────────────────────────────────────────

func (m InboxModel) Init() tea.Cmd {
	return m.load()
}

func (m InboxModel) load() tea.Cmd {
	return func() tea.Msg {
		items, err := m.repo.GetAll()
		if err != nil {
			return inboxErrMsg{err}
		}
		return inboxLoadedMsg{items}
	}
}

// ── Update ────────────────────────────────────────────────────────────────────

func (m InboxModel) Update(msg tea.Msg) (InboxModel, tea.Cmd) {
	switch msg := msg.(type) {

	case RefreshMsg:
		m.loading = true
		return m, m.load()

	case inboxLoadedMsg:
		m.loading = false
		m.err = nil
		m.items = msg.items
		if m.cursor >= len(m.items) {
			m.cursor = max(0, len(m.items)-1)
		}
		return m, nil

	case inboxErrMsg:
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

func (m InboxModel) handleKey(msg tea.KeyMsg) (InboxModel, tea.Cmd) {
	switch {
	case key.Matches(msg, inboxKeys.New):
		return m, func() tea.Msg { return OpenInboxEditorMsg{} }

	case key.Matches(msg, inboxKeys.Up):
		if m.cursor > 0 {
			m.cursor--
		}

	case key.Matches(msg, inboxKeys.Down):
		if m.cursor < len(m.items)-1 {
			m.cursor++
		}

	case key.Matches(msg, inboxKeys.Edit):
		if len(m.items) == 0 {
			return m, nil
		}
		id := m.items[m.cursor].ID
		return m, func() tea.Msg { return OpenInboxEditorMsg{ItemID: id} }

	case key.Matches(msg, inboxKeys.Delete):
		if len(m.items) == 0 {
			return m, nil
		}
		m.confirmDelete = m.items[m.cursor].ID
	}

	return m, nil
}

func (m InboxModel) handleConfirm(msg tea.KeyMsg) (InboxModel, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		id := m.confirmDelete
		m.confirmDelete = ""
		return m, func() tea.Msg {
			if err := m.repo.Delete(id); err != nil {
				return inboxErrMsg{err}
			}
			return RefreshMsg{}
		}
	default:
		m.confirmDelete = ""
	}
	return m, nil
}

// ── View ──────────────────────────────────────────────────────────────────────

func (m InboxModel) View() string {
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
				Render("Delete this item? y to confirm, any other key to cancel"),
		)
		b.WriteString("\n")
	}

	if len(m.items) == 0 {
		b.WriteString(StyleEmpty.Render("Inbox is empty."))
		b.WriteString("\n")
		b.WriteString(
			lipgloss.NewStyle().
				Foreground(colorSubtext).
				Italic(true).
				PaddingLeft(2).
				Render("Press n to capture something."),
		)
		return b.String()
	}

	for i, item := range m.items {
		b.WriteString(m.renderItem(i, item))
		b.WriteString("\n")
	}

	return b.String()
}

func (m InboxModel) renderHeader() string {
	count := fmt.Sprintf("%d item", len(m.items))
	if len(m.items) != 1 {
		count += "s"
	}

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(colorText).
		Render("Inbox")

	sub := lipgloss.NewStyle().
		Foreground(colorSubtext).
		Italic(true).
		Render("Capture zone — dump anything on your mind")

	meta := lipgloss.NewStyle().
		Foreground(colorSubtext).
		Render(count)

	sep := lipgloss.NewStyle().
		Foreground(colorBorder).
		Render(strings.Repeat("─", m.width-2))

	return lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.JoinHorizontal(lipgloss.Bottom, title, "  ", meta),
		sub,
		sep,
	)
}

func (m InboxModel) renderItem(i int, item model.InboxItem) string {
	selected := i == m.cursor

	cursor := "  "
	if selected {
		cursor = lipgloss.NewStyle().Foreground(colorPrimary).Render("▶ ")
	}

	var titleStyle lipgloss.Style
	if selected {
		titleStyle = StyleTaskSelected
	} else {
		titleStyle = StyleTaskNormal
	}

	title := titleStyle.Render(item.Title)
	row := fmt.Sprintf("%s%s", cursor, title)

	if item.Description != "" && selected {
		desc := StyleTaskDescription.Render(item.Description)
		return lipgloss.JoinVertical(lipgloss.Left, row, desc)
	}

	return row
}
