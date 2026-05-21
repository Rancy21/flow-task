package ui

import "github.com/charmbracelet/lipgloss"

// ── Palette ───────────────────────────────────────────────────────────────────

var (
	colorPrimary = lipgloss.Color("#4A90D9") // steel blue
	colorDark    = lipgloss.Color("#1A1A2E") // deep navy
	colorMuted   = lipgloss.Color("#6B7280") // gray
	colorSuccess = lipgloss.Color("#22C55E") // green
	colorWarning = lipgloss.Color("#F59E0B") // amber
	colorDanger  = lipgloss.Color("#EF4444") // red
	colorSurface = lipgloss.Color("#1E1E2E") // dark surface
	colorBorder  = lipgloss.Color("#313244") // subtle border
	colorText    = lipgloss.Color("#CDD6F4") // light text
	colorSubtext = lipgloss.Color("#6C7086") // dim text
)

// ── Priority badge styles ─────────────────────────────────────────────────────

var (
	StyleP1 = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(colorDanger).
		Padding(0, 1)

	StyleP2 = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(colorWarning).
		Padding(0, 1)

	StyleP3 = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(colorMuted).
		Padding(0, 1)
)

// ── Layout styles ─────────────────────────────────────────────────────────────

var (
	// App chrome
	StyleAppTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPrimary).
			Padding(0, 1)

	// Tab bar
	StyleTabActive = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPrimary).
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(colorPrimary).
			Padding(0, 2)

	StyleTabInactive = lipgloss.NewStyle().
				Foreground(colorSubtext).
				Padding(0, 2)

	// Section header (e.g. day label in week view)
	StyleSectionHeader = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorText).
				Background(colorSurface).
				Padding(0, 1).
				MarginTop(1)

	StyleSectionHeaderToday = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorPrimary).
				Background(colorSurface).
				Padding(0, 1).
				MarginTop(1)

	// Task item
	StyleTaskSelected = lipgloss.NewStyle().
				Foreground(colorPrimary).
				Bold(true)

	StyleTaskNormal = lipgloss.NewStyle().
			Foreground(colorText)

	StyleTaskDone = lipgloss.NewStyle().
			Foreground(colorSubtext).
			Strikethrough(true)

	StyleTaskDescription = lipgloss.NewStyle().
				Foreground(colorSubtext).
				Italic(true).
				PaddingLeft(4)

	// Status bar at the bottom
	StyleStatusBar = lipgloss.NewStyle().
			Foreground(colorSubtext).
			Background(colorSurface).
			Padding(0, 1)

	StyleStatusKey = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Bold(true)

	// Empty state
	StyleEmpty = lipgloss.NewStyle().
			Foreground(colorSubtext).
			Italic(true).
			Padding(2, 4)

	// Border box (editor, dialogs)
	StyleBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder).
			Padding(1, 2)

	StyleBoxFocused = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorPrimary).
			Padding(1, 2)

	// Note item
	StyleNoteContent = lipgloss.NewStyle().
				Foreground(colorText).
				PaddingLeft(2)

	StyleNoteMeta = lipgloss.NewStyle().
			Foreground(colorSubtext).
			Italic(true).
			PaddingLeft(2)
)

// PriorityStyle returns the badge style for a given priority string.
func PriorityStyle(p string) lipgloss.Style {
	switch p {
	case "P1":
		return StyleP1
	case "P2":
		return StyleP2
	default:
		return StyleP3
	}
}
