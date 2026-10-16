package storage

import (
	"app/internal/models"
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultsRoundTripAndCorruptFile(t *testing.T) {
	s := &FileStore{Path: filepath.Join(t.TempDir(), "go-clock", "data.json")}
	d, err := s.Load()
	if err != nil || d.Settings != models.DefaultSettings() {
		t.Fatal(d, err)
	}
	d.Settings.FocusesPerLongBreak = 3
	d.Stats = models.Stats{CompletedFocuses: 8, FocusesInCycle: 2, FocusSeconds: 12000}
	if err = s.Save(d); err != nil {
		t.Fatal(err)
	}
	loaded, err := s.Load()
	if err != nil || loaded != d {
		t.Fatal(loaded, err)
	}
	if err = os.WriteFile(s.Path, []byte("broken"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = s.Load(); err == nil {
		t.Fatal("corrupt file silently discarded")
	}
}
