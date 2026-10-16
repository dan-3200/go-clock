package models

import (
	"fmt"
	"time"
)

type Phase int

const (
	Pomodoro Phase = iota
	ShortBreak
	LongBreak
)

func (p Phase) String() string { return [...]string{"Pomodoro", "Short Break", "Long Break"}[p] }

type Settings struct {
	PomodoroMinutes     int `json:"pomodoroMinutes"`
	ShortBreakMinutes   int `json:"shortBreakMinutes"`
	LongBreakMinutes    int `json:"longBreakMinutes"`
	FocusesPerLongBreak int `json:"focusesPerLongBreak"`
}

func DefaultSettings() Settings { return Settings{25, 5, 15, 4} }
func (s Settings) Validate() error {
	for _, v := range []int{s.PomodoroMinutes, s.ShortBreakMinutes, s.LongBreakMinutes} {
		if v < 1 || v > 10080 {
			return fmt.Errorf("durações devem estar entre 1 e 10080 minutos")
		}
	}
	if s.FocusesPerLongBreak < 1 || s.FocusesPerLongBreak > 1000 {
		return fmt.Errorf("focos por ciclo devem estar entre 1 e 1000")
	}
	return nil
}
func (s Settings) Duration(p Phase) time.Duration {
	return time.Duration([]int{s.PomodoroMinutes, s.ShortBreakMinutes, s.LongBreakMinutes}[p]) * time.Minute
}

type Stats struct {
	CompletedFocuses int64 `json:"completedFocuses"`
	FocusesInCycle   int   `json:"focusesInCycle"`
	FocusSeconds     int64 `json:"focusSeconds"`
}
type Data struct {
	Settings Settings `json:"settings"`
	Stats    Stats    `json:"stats"`
}
