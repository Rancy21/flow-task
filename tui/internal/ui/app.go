package ui

import (
	"fmt"
	"strings"

	"github.com/Rancy21/flowtask/internal/repository"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ── Tabs ──────────────────────────────────────────────────────────────────────

type Tab int

const (
	TabToday Tab = iota
	TabWeek
	TabInbox
	TabNotes
)

var tabLabels = []string{"󰃰  Today", "󰅐  Week", "󰍉  Inbox", "󰎞  Notes"}
var tabLabelsFallback = []string{"Today", "Week", "Inbox", "Notes"}

// ── Messages ──────────────────────────────────────────────────────────────────

// RefreshMsg tells child models to reload their data from the DB.
type RefreshMsg struct{}

// OpenEditorMsg opens the task editor (empty TaskID = new task).
type OpenEditorMsg struct {
	TaskID string
}

// OpenInboxEditorMsg opens the inbox editor (empty ItemID = new item).
type OpenInboxEditorMsg struct {
	ItemID string
}

// ── Key bindings ──────────────────────────────────────────────────────────────

type keyMap struct {
	TabNext key.Binding
	TabPrev key.Binding
	Tab1    key.Binding
	Tab2    key.Binding
	Tab3    key.Binding
	Tab4    key.Binding
	New     key.Binding
	Quit    key.Binding
}

var keys = keyMap{
	TabNext: key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next tab")),
	TabPrev: key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("shift+tab", "prev tab")),
	Tab1:    key.NewBinding(key.WithKeys("1"), key.WithHelp("1", "today")),
	Tab2:    key.NewBinding(key.WithKeys("2"), key.WithHelp("2", "week")),
	Tab3:    key.NewBinding(key.WithKeys("3"), key.WithHelp("3", "inbox")),
	Tab4:    key.NewBinding(key.WithKeys("4"), key.WithHelp("4", "notes")),
	New:     key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "new")),
	Quit:    key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
}

// ── App model ─────────────────────────────────────────────────────────────────

// App is the root Bubble Tea model.
type App struct {
	taskRepo  *repository.TaskRepo
	noteRepo  *repository.NoteRepo
	inboxRepo *repository.InboxRepo

	activeTab Tab
	today     TodayModel
	week      WeekModel
	inbox     InboxModel
	notes     NotesModel

	// Task editor
	editor     EditorModel
	editorOpen bool

	// Inbox editor
	inboxEditor     InboxEditorModel
	inboxEditorOpen bool

	width  int
	height int
	err    error
}

func NewApp(
	taskRepo *repository.TaskRepo,
	noteRepo *repository.NoteRepo,
	inboxRepo *repository.InboxRepo,
) App {
	return App{
		taskRepo:  taskRepo,
		noteRepo:  noteRepo,
		inboxRepo: inboxRepo,
		today:     NewTodayModel(taskRepo),
		week:      NewWeekModel(taskRepo),
		inbox:     NewInboxModel(inboxRepo),
		notes:     NewNotesModel(taskRepo, noteRepo),
		editor:      NewEditorModel(taskRepo, noteRepo),
		inboxEditor: NewInboxEditorModel(inboxRepo),
	}
}

// ── Init ──────────────────────────────────────────────────────────────────────

func (a App) Init() tea.Cmd {
	return tea.Batch(
		a.today.Init(),
		a.week.Init(),
		a.inbox.Init(),
		a.notes.Init(),
	)
}

// ── Update ────────────────────────────────────────────────────────────────────

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.today.SetSize(msg.Width, a.contentHeight())
		a.week.SetSize(msg.Width, a.contentHeight())
		a.inbox.SetSize(msg.Width, a.contentHeight())
		a.notes.SetSize(msg.Width, a.contentHeight())
		a.editor.SetSize(msg.Width, a.contentHeight())
		a.inboxEditor.SetSize(msg.Width, a.contentHeight())

	case tea.KeyMsg:
		// Inbox editor intercepts all keys when open.
		if a.inboxEditorOpen {
			var cmd tea.Cmd
			a.inboxEditor, cmd = a.inboxEditor.Update(msg)
			cmds = append(cmds, cmd)
			return a, tea.Batch(cmds...)
		}

		// Task editor intercepts all keys when open.
		if a.editorOpen {
			var cmd tea.Cmd
			a.editor, cmd = a.editor.Update(msg)
			cmds = append(cmds, cmd)
			return a, tea.Batch(cmds...)
		}

		switch {
		case key.Matches(msg, keys.Quit):
			return a, tea.Quit
		case key.Matches(msg, keys.Tab1):
			a.activeTab = TabToday
		case key.Matches(msg, keys.Tab2):
			a.activeTab = TabWeek
		case key.Matches(msg, keys.Tab3):
			a.activeTab = TabInbox
		case key.Matches(msg, keys.Tab4):
			a.activeTab = TabNotes
		case key.Matches(msg, keys.TabNext):
			a.activeTab = (a.activeTab + 1) % 4
		case key.Matches(msg, keys.TabPrev):
			if a.activeTab == 0 {
				a.activeTab = 3
			} else {
				a.activeTab--
			}
		case key.Matches(msg, keys.New):
			// In inbox tab, n opens the inbox capture editor.
			if a.activeTab == TabInbox {
				a.inboxEditorOpen = true
				a.inboxEditor = NewInboxEditorModel(a.inboxRepo)
				a.inboxEditor.SetSize(a.width, a.contentHeight())
				cmds = append(cmds, a.inboxEditor.Init())
			} else {
				a.editorOpen = true
				a.editor = NewEditorModel(a.taskRepo, a.noteRepo)
				a.editor.SetSize(a.width, a.contentHeight())
				cmds = append(cmds, a.editor.Init())
			}
		}

	// ── Task editor messages ─────────────────────────────────────────────────
	case OpenEditorMsg:
		a.editorOpen = true
		a.editor = NewEditorModel(a.taskRepo, a.noteRepo)
		if msg.TaskID != "" {
			cmds = append(cmds, a.editor.LoadTask(msg.TaskID))
		}
		a.editor.SetSize(a.width, a.contentHeight())

	case EditorDoneMsg:
		a.editorOpen = false
		var cmd tea.Cmd
		a.today, cmd = a.today.Update(RefreshMsg{})
		cmds = append(cmds, cmd)
		a.week, cmd = a.week.Update(RefreshMsg{})
		cmds = append(cmds, cmd)
		a.inbox, cmd = a.inbox.Update(RefreshMsg{})
		cmds = append(cmds, cmd)
		a.notes, cmd = a.notes.Update(RefreshMsg{})
		cmds = append(cmds, cmd)

	case EditorCancelMsg:
		a.editorOpen = false

	// ── Inbox editor messages ────────────────────────────────────────────────
	case OpenInboxEditorMsg:
		a.inboxEditorOpen = true
		a.inboxEditor = NewInboxEditorModel(a.inboxRepo)
		if msg.ItemID != "" {
			cmds = append(cmds, a.inboxEditor.LoadItem(msg.ItemID))
		}
		a.inboxEditor.SetSize(a.width, a.contentHeight())

	case InboxEditorDoneMsg:
		a.inboxEditorOpen = false
		var cmd tea.Cmd
		a.inbox, cmd = a.inbox.Update(RefreshMsg{})
		cmds = append(cmds, cmd)

	case InboxEditorCancelMsg:
		a.inboxEditorOpen = false
	}

	// Route remaining messages.
	if a.inboxEditorOpen {
		var cmd tea.Cmd
		a.inboxEditor, cmd = a.inboxEditor.Update(msg)
		cmds = append(cmds, cmd)
	} else if a.editorOpen {
		var cmd tea.Cmd
		a.editor, cmd = a.editor.Update(msg)
		cmds = append(cmds, cmd)
	} else {
		if _, isKey := msg.(tea.KeyMsg); isKey {
			var cmd tea.Cmd
			switch a.activeTab {
			case TabToday:
				a.today, cmd = a.today.Update(msg)
			case TabWeek:
				a.week, cmd = a.week.Update(msg)
			case TabInbox:
				a.inbox, cmd = a.inbox.Update(msg)
			case TabNotes:
				a.notes, cmd = a.notes.Update(msg)
			}
			cmds = append(cmds, cmd)
		} else {
			var cmd tea.Cmd
			a.today, cmd = a.today.Update(msg)
			cmds = append(cmds, cmd)
			a.week, cmd = a.week.Update(msg)
			cmds = append(cmds, cmd)
			a.inbox, cmd = a.inbox.Update(msg)
			cmds = append(cmds, cmd)
			a.notes, cmd = a.notes.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	return a, tea.Batch(cmds...)
}

// ── View ──────────────────────────────────────────────────────────────────────

func (a App) View() string {
	if a.inboxEditorOpen {
		return a.inboxEditor.View()
	}
	if a.editorOpen {
		return a.editor.View()
	}

	header := a.renderHeader()
	content := a.renderContent()
	statusBar := a.renderStatusBar()

	return lipgloss.JoinVertical(lipgloss.Left, header, content, statusBar)
}

func (a App) renderHeader() string {
	title := StyleAppTitle.Render("FlowTask")

	tabs := make([]string, len(tabLabels))
	for i, label := range tabLabels {
		if Tab(i) == a.activeTab {
			tabs[i] = StyleTabActive.Render(label)
		} else {
			tabs[i] = StyleTabInactive.Render(label)
		}
	}

	tabBar := lipgloss.JoinHorizontal(lipgloss.Bottom, tabs...)
	gap := a.width - lipgloss.Width(title) - lipgloss.Width(tabBar) - 2
	if gap < 0 {
		gap = 0
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Bottom,
		title,
		strings.Repeat(" ", gap),
		tabBar,
	)
}

func (a App) renderContent() string {
	var view string
	switch a.activeTab {
	case TabToday:
		view = a.today.View()
	case TabWeek:
		view = a.week.View()
	case TabInbox:
		view = a.inbox.View()
	case TabNotes:
		view = a.notes.View()
	}

	// Pad content to always fill the available height so the status bar
	// stays anchored and doesn't jump between tabs with different lengths.
	ch := a.contentHeight()
	lines := strings.Count(view, "\n") + 1
	for i := lines; i < ch; i++ {
		view += "\n"
	}
	return view
}

func (a App) renderStatusBar() string {
	var hints []string
	switch a.activeTab {
	case TabToday, TabWeek:
		hints = []string{
			fmt.Sprintf("%s new", StyleStatusKey.Render("n")),
			fmt.Sprintf("%s done", StyleStatusKey.Render("space")),
			fmt.Sprintf("%s edit", StyleStatusKey.Render("e")),
			fmt.Sprintf("%s delete", StyleStatusKey.Render("d")),
		}
	case TabInbox:
		hints = []string{
			fmt.Sprintf("%s capture", StyleStatusKey.Render("n")),
			fmt.Sprintf("%s edit", StyleStatusKey.Render("e")),
			fmt.Sprintf("%s delete", StyleStatusKey.Render("d")),
		}
	case TabNotes:
		hints = []string{
			fmt.Sprintf("%s new task", StyleStatusKey.Render("n")),
			fmt.Sprintf("%s view task", StyleStatusKey.Render("enter")),
			fmt.Sprintf("%s delete", StyleStatusKey.Render("d")),
		}
	}
	hints = append(hints,
		fmt.Sprintf("%s tabs", StyleStatusKey.Render("1-4")),
		fmt.Sprintf("%s quit", StyleStatusKey.Render("q")),
	)
	bar := strings.Join(hints, "  ")
	return StyleStatusBar.Width(a.width).Render(bar)
}

func (a App) contentHeight() int {
	// header (1) + status bar (1) + padding
	return a.height - 3
}
