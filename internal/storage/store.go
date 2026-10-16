package storage

import (
	"app/internal/models"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type Store interface {
	Load() (models.Data, error)
	Save(models.Data) error
}
type FileStore struct{ Path string }

func New() (*FileStore, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}
	return &FileStore{filepath.Join(dir, "go-clock", "data.json")}, nil
}
func (s *FileStore) Load() (models.Data, error) {
	d := models.Data{Settings: models.DefaultSettings()}
	b, err := os.ReadFile(s.Path)
	if errors.Is(err, os.ErrNotExist) {
		return d, nil
	}
	if err != nil {
		return d, err
	}
	if err = json.Unmarshal(b, &d); err != nil {
		return d, fmt.Errorf("dados inválidos: %w", err)
	}
	if err = d.Settings.Validate(); err != nil {
		return d, err
	}
	if d.Stats.CompletedFocuses < 0 || d.Stats.FocusesInCycle < 0 || d.Stats.FocusSeconds < 0 {
		return d, fmt.Errorf("métricas inválidas")
	}
	return d, nil
}
func (s *FileStore) Save(d models.Data) error {
	if err := d.Settings.Validate(); err != nil {
		return err
	}
	b, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return err
	}
	dir := filepath.Dir(s.Path)
	if err = os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	f, err := os.CreateTemp(dir, ".data-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if _, err = f.Write(b); err != nil {
		f.Close()
		return err
	}
	if err = f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	return os.Rename(f.Name(), s.Path)
}
