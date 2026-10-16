package engine

import (
	"app/internal/models"
	"errors"
	tea "github.com/charmbracelet/bubbletea"
	"testing"
	"time"
)

type memoryStore struct {
	data models.Data
	err  error
}

func (s *memoryStore) Load() (models.Data, error) { return s.data, s.err }
func (s *memoryStore) Save(d models.Data) error {
	if s.err == nil {
		s.data = d
	}
	return s.err
}

type fakeNotifier struct{ calls int }

func (n *fakeNotifier) Notify(string, string) error {
	n.calls++
	return errors.New("desktop unavailable")
}
func setup() (App, *memoryStore, *fakeNotifier) {
	s := &memoryStore{}
	n := &fakeNotifier{}
	a := New(models.Data{Settings: models.DefaultSettings()}, s, n)
	a.sound = func() error { return nil }
	return a, s, n
}
func key(a App, k tea.KeyType) App { m, _ := a.Update(tea.KeyMsg{Type: k}); return m.(App) }
func complete(a App) (App, tea.Cmd) {
	a.Running = true
	a.Deadline = time.Now()
	m, c := a.Update(tickMsg{a.generation, a.Deadline})
	return m.(App), c
}
func TestCycleAndStats(t *testing.T) {
	a, s, _ := setup()
	for i := 1; i <= 4; i++ {
		a, _ = complete(a)
		want := models.ShortBreak
		if i == 4 {
			want = models.LongBreak
		}
		if a.Phase != want || a.Running || a.Paused || a.Data.Stats.CompletedFocuses != int64(i) {
			t.Fatalf("focus %d: %+v", i, a)
		}
		a, _ = complete(a)
		if a.Phase != models.Pomodoro {
			t.Fatal("break did not return to focus")
		}
	}
	if s.data.Stats.FocusSeconds != 6000 || s.data.Stats.FocusesInCycle != 0 {
		t.Fatalf("stats: %+v", s.data.Stats)
	}
	a.Data.Settings.FocusesPerLongBreak = 1
	a, _ = complete(a)
	if a.Phase != models.LongBreak {
		t.Fatal("configured cycle ignored")
	}
}
func TestPauseResetAndStaleTick(t *testing.T) {
	a, _, _ := setup()
	a = key(a, tea.KeyEnter)
	old := a.generation
	a = key(a, tea.KeyEnter)
	before := a.Remaining
	m, c := a.Update(tickMsg{old, time.Now().Add(time.Hour)})
	a = m.(App)
	if c != nil || a.Remaining != before || a.Running {
		t.Fatal("paused timer advanced")
	}
	a = key(a, tea.KeyEnter)
	m, c = a.Update(tickMsg{old, time.Now().Add(time.Hour)})
	a = m.(App)
	if c != nil || a.Data.Stats.CompletedFocuses != 0 {
		t.Fatal("stale tick completed focus")
	}
	m, _ = a.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	a = m.(App)
	if a.Running || a.Remaining != a.Duration {
		t.Fatal("reset failed")
	}
	a.next(false)
	if a.Data.Stats.CompletedFocuses != 0 || a.Data.Stats.FocusSeconds != 0 {
		t.Fatal("skipped focus counted")
	}
}
func TestSettingsApplyNextPeriodAndValidate(t *testing.T) {
	a, s, _ := setup()
	original := a.Duration
	a.openSettings()
	a.inputs[0].SetValue("0")
	a = key(a, tea.KeyEnter)
	if a.Err == nil || !a.Configuring {
		t.Fatal("invalid settings accepted")
	}
	a.inputs[0].SetValue("10")
	a = key(a, tea.KeyEnter)
	if a.Configuring || a.Err != nil || a.Duration != original || s.data.Settings.PomodoroMinutes != 10 {
		t.Fatal("save changed current period or failed")
	}
	a, _ = complete(a)
	if a.Data.Stats.FocusSeconds != 1500 {
		t.Fatal("current focus duration changed")
	}
	a.next(false)
	if a.Duration != 10*time.Minute {
		t.Fatal("new focus duration not applied")
	}
}
func TestAlertFailureAndSaveRetry(t *testing.T) {
	a, s, n := setup()
	s.err = errors.New("disk full")
	a, cmd := complete(a)
	if !a.dirty || a.Err == nil {
		t.Fatal("save failure hidden")
	}
	batch := cmd().(tea.BatchMsg)
	for _, c := range batch {
		msg := c()
		m, _ := a.Update(msg)
		a = m.(App)
	}
	if n.calls != 1 || a.Running || a.Phase != models.ShortBreak || a.Message == "" {
		t.Fatal("alert failure disrupted timer")
	}
	s.err = nil
	a = key(a, tea.KeyCtrlC)
	if a.dirty || s.data.Stats.CompletedFocuses != 1 {
		t.Fatal("save was not retried")
	}
}
