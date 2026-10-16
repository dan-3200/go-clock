package internal

import (
	"app/internal/engine"
	"app/internal/notifications"
	"app/internal/storage"
	tea "github.com/charmbracelet/bubbletea"
)

func Run() error {
	store, err := storage.New()
	if err != nil {
		return err
	}
	data, err := store.Load()
	if err != nil {
		return err
	}
	_, err = tea.NewProgram(engine.New(data, store, notifications.Desktop{}), tea.WithAltScreen()).Run()
	return err
}
