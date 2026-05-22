package main

import (
	"fmt"
	"os"
	"time"

	"github.com/Rancy21/flowtask/internal/db"
	"github.com/Rancy21/flowtask/internal/repository"
	"github.com/Rancy21/flowtask/internal/sync"
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

	// Pull latest from Supabase on startup (best effort).
	syncClient := sync.New(taskRepo, noteRepo, inboxRepo)
	if err := syncClient.PullAll(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: initial sync pull failed: %v\n", err)
	}

	// Periodic background pull every 30 seconds.
	stopCh := make(chan struct{})
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				_ = syncClient.PullAll()
			case <-stopCh:
				return
			}
		}
	}()

	app := ui.NewApp(taskRepo, noteRepo, inboxRepo, syncClient)

	p := tea.NewProgram(app, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running app: %v\n", err)
		os.Exit(1)
	}

	close(stopCh)
}
