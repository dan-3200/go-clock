package engine

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"app/internal/components"
	"app/internal/models"
	"app/internal/notifications"
	"app/internal/storage"
)

type tickMsg struct {
	generation uint64
	now        time.Time
}
type alertMsg struct {
	err  error
	bell bool
}
type App struct {
	Data                         models.Data
	Phase                        models.Phase
	Remaining, Duration          time.Duration
	Running, Paused, Configuring bool
	Deadline                     time.Time
	generation                   uint64
	store                        storage.Store
	notifier                     notifications.Notifier
	sound                        func() error
	inputs                       []textinput.Model
	selected                     int
	Width, Height                int
	Message                      string
	Err                          error
	dirty                        bool
}

func New(d models.Data, s storage.Store, n notifications.Notifier) App {
	return App{Data: d, Phase: models.Pomodoro, Duration: d.Settings.Duration(models.Pomodoro), Remaining: d.Settings.Duration(models.Pomodoro), store: s, notifier: n, sound: notifications.Sound}
}
func (a App) Init() tea.Cmd { return nil }
func tick(g uint64) tea.Cmd {
	return tea.Tick(200*time.Millisecond, func(t time.Time) tea.Msg { return tickMsg{g, t} })
}
func (a *App) save() { a.Err = a.store.Save(a.Data); a.dirty = a.Err != nil }
func (a *App) next(completed bool) {
	if a.Phase == models.Pomodoro {
		if completed {
			a.Data.Stats.CompletedFocuses++
			a.Data.Stats.FocusesInCycle++
			a.Data.Stats.FocusSeconds += int64(a.Duration / time.Second)
			a.save()
		}
		a.Phase = models.ShortBreak
		if a.Data.Stats.FocusesInCycle >= a.Data.Settings.FocusesPerLongBreak {
			a.Phase = models.LongBreak
		}
	} else {
		if a.Phase == models.LongBreak {
			a.Data.Stats.FocusesInCycle = 0
			a.save()
		}
		a.Phase = models.Pomodoro
	}
	a.Running, a.Paused = false, false
	a.generation++
	a.Duration = a.Data.Settings.Duration(a.Phase)
	a.Remaining = a.Duration
}
func (a *App) openSettings() {
	a.Configuring = true
	a.selected = 0
	a.inputs = nil
	for i, v := range []int{a.Data.Settings.PomodoroMinutes, a.Data.Settings.ShortBreakMinutes, a.Data.Settings.LongBreakMinutes, a.Data.Settings.FocusesPerLongBreak} {
		in := textinput.New()
		in.SetValue(strconv.Itoa(v))
		in.CharLimit = 5
		in.Width = 10
		if i == 0 {
			in.Focus()
		}
		a.inputs = append(a.inputs, in)
	}
}
func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		a.Width, a.Height = m.Width, m.Height
	case alertMsg:
		if m.err != nil {
			a.Message += " • Notificação desktop indisponível"
		}
		if m.bell {
			a.Message += " • Som nativo indisponível; usando campainha do terminal"
			return a, func() tea.Msg { _, _ = os.Stdout.WriteString("\a"); return nil }
		}
	case tickMsg:
		if !a.Running || m.generation != a.generation {
			return a, nil
		}
		a.Remaining = a.Deadline.Sub(m.now)
		if a.Remaining <= 0 {
			phase := a.Phase.String()
			a.next(true)
			a.Message = phase + " concluído! Inicie " + a.Phase.String() + " quando quiser."
			n := a.notifier
			sound := a.sound
			return a, tea.Batch(
				func() tea.Msg { return alertMsg{err: n.Notify("Go Clock", phase+" concluído!")} },
				func() tea.Msg { return alertMsg{bell: sound() != nil} },
			)
		}
		return a, tick(a.generation)
	case tea.KeyMsg:
		key := m.String()
		if key == "ctrl+c" {
			if a.dirty {
				a.save()
				if a.Err != nil {
					return a, nil
				}
			}
			return a, tea.Quit
		}
		if a.Configuring {
			switch key {
			case "esc":
				a.Configuring = false
				return a, nil
			case "tab", "down", "shift+tab", "up":
				a.inputs[a.selected].Blur()
				delta := 1
				if key == "up" || key == "shift+tab" {
					delta = -1
				}
				a.selected = (a.selected + delta + 4) % 4
				return a, a.inputs[a.selected].Focus()
			case "enter":
				var values [4]int
				for i, in := range a.inputs {
					v, err := strconv.Atoi(in.Value())
					if err != nil {
						a.Err = fmt.Errorf("digite apenas números inteiros positivos")
						return a, nil
					}
					values[i] = v
				}
				settings := models.Settings{PomodoroMinutes: values[0], ShortBreakMinutes: values[1], LongBreakMinutes: values[2], FocusesPerLongBreak: values[3]}
				if err := settings.Validate(); err != nil {
					a.Err = err
					return a, nil
				}
				candidate := a.Data
				candidate.Settings = settings
				if err := a.store.Save(candidate); err != nil {
					a.Err = err
					return a, nil
				}
				a.Data = candidate
				a.dirty = false
				a.Err = nil
				a.Configuring = false
				a.Message = "Configurações salvas; aplicadas ao próximo período."
				return a, nil
			}
			var cmd tea.Cmd
			a.inputs[a.selected], cmd = a.inputs[a.selected].Update(msg)
			return a, cmd
		}
		switch key {
		case "q", "esc":
			if a.dirty {
				a.save()
				if a.Err != nil {
					return a, nil
				}
			}
			return a, tea.Quit
		case "s":
			a.openSettings()
			return a, textinput.Blink
		case "space", " ", "enter":
			a.generation++
			if a.Running {
				a.Remaining = time.Until(a.Deadline)
				if a.Remaining < 0 {
					a.Remaining = 0
				}
				a.Running = false
				a.Paused = true
			} else {
				a.Running = true
				a.Paused = false
				a.Deadline = time.Now().Add(a.Remaining)
				return a, tick(a.generation)
			}
		case "r":
			a.generation++
			a.Running = false
			a.Paused = false
			a.Remaining = a.Duration
			a.Message = "Período reiniciado."
		case "n":
			a.next(false)
			a.Message = "Período avançado; nenhum foco contabilizado."
		}
	}
	return a, nil
}
func (a App) View() string {
	if a.Configuring {
		fields := make([]string, len(a.inputs))
		for i, in := range a.inputs {
			fields[i] = in.View()
		}
		return components.Settings(fields, a.Err, a.Width)
	}
	state := "Parado"
	if a.Running {
		state = "Em execução"
	} else if a.Paused {
		state = "Pausado"
	}
	return components.Timer(a.Phase, a.Remaining, a.Duration, state, a.Data.Stats, a.Data.Settings.FocusesPerLongBreak, a.Message, a.Err, a.Width)
}
