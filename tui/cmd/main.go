package main

import (
	"fmt"
	"os"

	"github.com/Rancy21/flowtask/internal/db"
	"github.com/Rancy21/flowtask/internal/repository"
	"github.com/Rancy21/flowtask/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	database, err := db.Open()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening database: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	taskRepo := repository.NewTaskRepo(database)
	noteRepo := repository.NewNoteRepo(database)
	inboxRepo := repository.NewInboxRepo(database)

	app := ui.NewApp(taskRepo, noteRepo, inboxRepo)

	p := tea.NewProgram(app, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running app: %v\n", err)
		os.Exit(1)
	}
}
