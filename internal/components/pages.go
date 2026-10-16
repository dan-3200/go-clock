package components

import (
	"app/internal/models"
	"fmt"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/lipgloss"
	"math"
	"strings"
	"time"
)

var muted = lipgloss.NewStyle().Foreground(lipgloss.Color("#9095A5"))
var title = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#87CEEB"))

func frame(content string, width int) string {
	if width > 76 {
		return lipgloss.NewStyle().Padding(1, 2).Width(72).Render(content)
	}
	return lipgloss.NewStyle().Padding(1).Render(content)
}
func errorText(err error) string {
	if err == nil {
		return ""
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#FAA0A0")).Render("Erro: " + err.Error())
}
func Timer(phase models.Phase, remaining, duration time.Duration, state string, stats models.Stats, cycle int, message string, err error, width int) string {
	seconds := int(math.Ceil(remaining.Seconds()))
	if seconds < 0 {
		seconds = 0
	}
	barWidth := 40
	if width > 0 && width < 46 {
		barWidth = max(8, width-6)
	}
	bar := progress.New(progress.WithDefaultGradient(), progress.WithWidth(barWidth), progress.WithoutPercentage())
	fraction := 1 - float64(remaining)/float64(duration)
	fraction = max(0, min(1, fraction))
	clock := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#CF9FFF")).Render(fmt.Sprintf("%02d:%02d", seconds/60, seconds%60))
	return frame(strings.Join([]string{title.Render("GO CLOCK  •  " + phase.String()), "", clock + "  " + muted.Render(state), bar.ViewAs(fraction), "", fmt.Sprintf("Ciclos de foco concluídos: %d", stats.CompletedFocuses), fmt.Sprintf("Tempo de foco: %dh %02dmin", stats.FocusSeconds/3600, (stats.FocusSeconds%3600)/60), fmt.Sprintf("Focos no ciclo: %d / %d", stats.FocusesInCycle, cycle), "", message, errorText(err), "", muted.Render("[Espaço/Enter] Iniciar/Pausar  [r] Reiniciar  [n] Avançar\n[s] Settings  [q/Esc/Ctrl+C] Sair")}, "\n"), width)
}
func Settings(fields []string, err error, width int) string {
	labels := []string{"Pomodoro (min)", "Short Break (min)", "Long Break (min)", "Focos antes de Long Break"}
	lines := []string{title.Render("GO CLOCK  •  Settings"), ""}
	for i, label := range labels {
		lines = append(lines, label+": "+fields[i])
	}
	lines = append(lines, "", "Durações: 1–10080 min. Focos por ciclo: 1–1000.", "O relógio continua enquanto Settings está aberto.", "Alterações valem a partir do próximo período.", errorText(err), "", muted.Render("[Tab/↑/↓] Campo  [Enter] Salvar  [Esc] Cancelar"))
	return frame(strings.Join(lines, "\n"), width)
}
