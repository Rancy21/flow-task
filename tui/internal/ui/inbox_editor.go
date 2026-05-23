package ui

import (
	"fmt"
	"strings"

	"github.com/Rancy21/flowtask/internal/repository"
	"github.com/Rancy21/flowtask/internal/sync"
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
type inboxEditorErrMsg struct{ err error }

// ── Field indices ─────────────────────────────────────────────────────────────

type inboxEditField int

const (
	ifieldTitle inboxEditField = iota
	ifieldDescription
	ifieldActions
	ifieldCount = 3
)

// ── Key bindings ──────────────────────────────────────────────────────────────

type inboxEditKeyMap struct {
	NextField key.Binding
	PrevField key.Binding
	Save      key.Binding
	Cancel    key.Binding
}

var inboxEditKeys = inboxEditKeyMap{
	NextField: key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next field")),
	PrevField: key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("shift+tab", "prev field")),
	Save:      key.NewBinding(key.WithKeys("ctrl+s"), key.WithHelp("ctrl+s", "save")),
	Cancel:    key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel")),
}

// ── Model ─────────────────────────────────────────────────────────────────────

type InboxEditorModel struct {
	repo   *repository.InboxRepo
	sync   *sync.Client
	itemID string
	mode   string // "new" | "edit"

	titleInput  textinput.Model
	descInput   textarea.Model
	activeField inboxEditField

	width      int
	height     int
	innerWidth int

	err    error
	saving bool
}

func NewInboxEditorModel(repo *repository.InboxRepo, sync *sync.Client) InboxEditorModel {
	ti := textinput.New()
	ti.Placeholder = "What's on your mind?"
	ti.CharLimit = 200
	ti.Focus()

	di := textarea.New()
	di.Placeholder = "Add details..."
	di.CharLimit = 2000
	di.SetHeight(5)
	di.ShowLineNumbers = true

	return InboxEditorModel{
		repo:        repo,
		sync:        sync,
		mode:        "new",
		titleInput:  ti,
		descInput:   di,
		activeField: ifieldTitle,
	}
}

func (m *InboxEditorModel) SetSize(w, h int) {
	m.width = w
	m.height = h
	formW := min(72, w-6)
	m.innerWidth = formW - 8
	m.titleInput.Width = m.innerWidth
	m.descInput.SetWidth(m.innerWidth)
}

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

		// Tab only navigates when NOT in description.
		if key.Matches(msg, inboxEditKeys.NextField) && m.activeField != ifieldDescription {
			m.nextField()
			return m, nil
		}
		if key.Matches(msg, inboxEditKeys.PrevField) && m.activeField != ifieldDescription {
			m.prevField()
			return m, nil
		}

		switch m.activeField {
		case ifieldTitle:
			var cmd tea.Cmd
			m.titleInput, cmd = m.titleInput.Update(msg)
			cmds = append(cmds, cmd)

		case ifieldDescription:
			// ctrl+j exits description field downward
			if msg.String() == "ctrl+j" || msg.String() == "ctrl+down" {
				m.nextField()
				return m, nil
			}
			var cmd tea.Cmd
			m.descInput, cmd = m.descInput.Update(msg)
			cmds = append(cmds, cmd)

		case ifieldActions:
			if msg.String() == "enter" {
				return m, m.save()
			}
		}
	}

	return m, tea.Batch(cmds...)
}

// ── View ──────────────────────────────────────────────────────────────────────

func (m InboxEditorModel) View() string {
	formW := min(72, m.width-6)
	innerW := formW - 8

	var sections []string

	// ── Header ────────────────────────────────────────────────────────────────
	label := "CAPTURE"
	if m.mode == "edit" {
		label = "EDIT ITEM"
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
		"TITLE", m.titleInput.View(), ifieldTitle,
	))

	// ── Description ───────────────────────────────────────────────────────────
	sections = append(sections, m.renderTextSection(
		"DESCRIPTION  (ctrl+j to continue)", m.descInput.View(), ifieldDescription,
	))

	// ── Actions ───────────────────────────────────────────────────────────────
	sections = append(sections, "")
	sections = append(sections, m.renderActions())

	// ── Footer ────────────────────────────────────────────────────────────────
	sections = append(sections, divider.Render(strings.Repeat("─", innerW)))
	sections = append(sections, editorFooter.Render(
		"ctrl+s save  ·  esc cancel  ·  tab/shift+tab navigate",
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

func (m InboxEditorModel) renderTextSection(label, content string, f inboxEditField) string {
	active := m.activeField == f
	lStyle := fieldLabel
	bStyle := fieldBox
	if active {
		lStyle = fieldLabelActive
		bStyle = fieldBoxActive
	}
	return lipgloss.JoinVertical(lipgloss.Left,
		lStyle.Render(label),
		bStyle.Render(content),
	)
}

func (m InboxEditorModel) renderActions() string {
	var s lipgloss.Style
	if m.activeField == ifieldActions {
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

func (m *InboxEditorModel) nextField() {
	m.activeField = (m.activeField + 1) % ifieldCount
	m.updateFocus()
}

func (m *InboxEditorModel) prevField() {
	if m.activeField == 0 {
		m.activeField = ifieldCount - 1
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
			_ = m.sync.PushInboxItem(item)
			return InboxEditorDoneMsg{}
		}
	}

	return func() tea.Msg {
		item, err := m.repo.Create(title, desc)
		if err != nil {
			return inboxEditorErrMsg{err}
		}
		_ = m.sync.PushInboxItem(item)
		return InboxEditorDoneMsg{}
	}
}