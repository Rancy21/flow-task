package ui

import (
	"fmt"
	"strings"

	"github.com/Rancy21/flowtask/internal/repository"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ── Messages ──────────────────────────────────────────────────────────────────

type InboxEditorDoneMsg struct{}
type InboxEditorCancelMsg struct{}
type inboxItemLoadedMsg struct{ title, description string }

// inboxEditorErrMsg carries a save/load error.
type inboxEditorErrMsg struct{ err error }

// ── Field indices ─────────────────────────────────────────────────────────────

const (
	ifieldTitle inboxEditField = iota
	ifieldDescription
	ifieldActions
)

type inboxEditField int

// ── Key bindings ──────────────────────────────────────────────────────────────

type inboxEditKeyMap struct {
	Tab      key.Binding
	ShiftTab key.Binding
	Save     key.Binding
	Cancel   key.Binding
}

var inboxEditKeys = inboxEditKeyMap{
	Tab:      key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next field")),
	ShiftTab: key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("shift+tab", "prev field")),
	Save:     key.NewBinding(key.WithKeys("ctrl+s"), key.WithHelp("ctrl+s", "save")),
	Cancel:   key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel")),
}

// ── Model ─────────────────────────────────────────────────────────────────────

type InboxEditorModel struct {
	repo       *repository.InboxRepo
	itemID     string // empty = new item
	mode       string // "new" or "edit"
	titleInput textinput.Model
	descInput  textarea.Model
	activeField inboxEditField
	width      int
	height     int
	err        error
	saving     bool
}

func NewInboxEditorModel(repo *repository.InboxRepo) InboxEditorModel {
	ti := textinput.New()
	ti.Placeholder = "What's on your mind?"
	ti.CharLimit = 200
	ti.Focus()

	di := textarea.New()
	di.Placeholder = "Details (optional)"
	di.CharLimit = 2000
	di.SetHeight(4)

	return InboxEditorModel{
		repo:        repo,
		mode:        "new",
		titleInput:  ti,
		descInput:   di,
		activeField: ifieldTitle,
	}
}

func (m *InboxEditorModel) SetSize(w, h int) {
	m.width = w
	m.height = h
	m.titleInput.Width = max(20, w-12)
	m.descInput.SetWidth(max(20, w-12))
}

// LoadItem returns a command that fetches an inbox item for editing.
func (m *InboxEditorModel) LoadItem(itemID string) tea.Cmd {
	m.itemID = itemID
	m.mode = "edit"
	return func() tea.Msg {
		item, err := m.repo.GetByID(itemID)
		if err != nil {
			return inboxEditorErrMsg{err}
		}
		return inboxItemLoadedMsg{
			title:       item.Title,
			description: item.Description,
		}
	}
}

// ── Init ──────────────────────────────────────────────────────────────────────

func (m InboxEditorModel) Init() tea.Cmd {
	return textinput.Blink
}

// ── Update ────────────────────────────────────────────────────────────────────

func (m InboxEditorModel) Update(msg tea.Msg) (InboxEditorModel, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	case inboxItemLoadedMsg:
		m.titleInput.SetValue(msg.title)
		m.descInput.SetValue(msg.description)
		return m, nil

	case inboxEditorErrMsg:
		m.err = msg.err
		m.saving = false
		return m, nil

	case tea.KeyMsg:
		if key.Matches(msg, inboxEditKeys.Cancel) {
			return m, func() tea.Msg { return InboxEditorCancelMsg{} }
		}
		if key.Matches(msg, inboxEditKeys.Save) {
			return m, m.save()
		}

		if key.Matches(msg, inboxEditKeys.Tab) {
			m.nextField()
			return m, nil
		}
		if key.Matches(msg, inboxEditKeys.ShiftTab) {
			m.prevField()
			return m, nil
		}

		switch m.activeField {

		case ifieldTitle:
			var cmd tea.Cmd
			m.titleInput, cmd = m.titleInput.Update(msg)
			cmds = append(cmds, cmd)

		case ifieldDescription:
			var cmd tea.Cmd
			m.descInput, cmd = m.descInput.Update(msg)
			cmds = append(cmds, cmd)
		}

		// Enter on actions = save.
		if m.activeField == ifieldActions && msg.String() == "enter" {
			return m, m.save()
		}
	}

	return m, tea.Batch(cmds...)
}

// ── View ──────────────────────────────────────────────────────────────────────

func (m InboxEditorModel) View() string {
	formWidth := min(60, m.width-4)

	headerLabel := "INBOX CAPTURE"
	if m.mode == "edit" {
		headerLabel = "EDIT ITEM"
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

	b.WriteString(m.renderField("Title", m.titleInput.View(), ifieldTitle))
	b.WriteString("\n")
	b.WriteString(m.renderField("Description", m.descInput.View(), ifieldDescription))
	b.WriteString("\n")
	b.WriteString(m.renderField("", m.renderActions(), ifieldActions))
	b.WriteString("\n")

	footer := lipgloss.NewStyle().
		Foreground(colorSubtext).
		MarginTop(1).
		Render("ctrl+s save  ·  esc cancel  ·  tab next field")

	content := lipgloss.NewStyle().
		Width(formWidth).
		Render(lipgloss.JoinVertical(lipgloss.Left, b.String(), footer))

	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		StyleBox.Render(content),
	)
}

func (m InboxEditorModel) renderField(label, content string, f inboxEditField) string {
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

func (m InboxEditorModel) renderActions() string {
	saveStyle := lipgloss.NewStyle().
		Foreground(colorSuccess).
		Bold(true)

	cancelStyle := lipgloss.NewStyle().
		Foreground(colorSubtext)

	if m.activeField == ifieldActions {
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

func (m *InboxEditorModel) nextField() {
	m.activeField = (m.activeField + 1) % 3
	m.updateFocus()
}

func (m *InboxEditorModel) prevField() {
	if m.activeField == 0 {
		m.activeField = 2
	} else {
		m.activeField--
	}
	m.updateFocus()
}

func (m *InboxEditorModel) updateFocus() {
	m.titleInput.Blur()
	m.descInput.Blur()
	switch m.activeField {
	case ifieldTitle:
		m.titleInput.Focus()
	case ifieldDescription:
		m.descInput.Focus()
	}
}

// ── Save ──────────────────────────────────────────────────────────────────────

func (m InboxEditorModel) save() tea.Cmd {
	title := strings.TrimSpace(m.titleInput.Value())
	if title == "" {
		m.err = fmt.Errorf("title is required")
		return nil
	}

	m.saving = true
	m.err = nil

	desc := strings.TrimSpace(m.descInput.Value())

	if m.mode == "edit" {
		return func() tea.Msg {
			item, err := m.repo.GetByID(m.itemID)
			if err != nil {
				return inboxEditorErrMsg{err}
			}
			item.Title = title
			item.Description = desc
			if err := m.repo.Update(item); err != nil {
				return inboxEditorErrMsg{err}
			}
			return InboxEditorDoneMsg{}
		}
	}

	return func() tea.Msg {
		_, err := m.repo.Create(title, desc)
		if err != nil {
			return inboxEditorErrMsg{err}
		}
		return InboxEditorDoneMsg{}
	}
}
