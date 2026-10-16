package notifications

import (
	"context"
	"fmt"
	"github.com/gen2brain/beeep"
	"os/exec"
	"runtime"
	"time"
)

type Notifier interface {
	Notify(title, body string) error
}
type Desktop struct{}

func (Desktop) Notify(title, body string) error { return beeep.Notify(title, body, "") }

// Sound uses a native sound player; the UI emits a terminal bell on failure.
func Sound() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.CommandContext(ctx, "afplay", "/System/Library/Sounds/Glass.aiff")
	case "windows":
		cmd = exec.CommandContext(ctx, "powershell", "-NoProfile", "-NonInteractive", "-Command", "[System.Media.SystemSounds]::Exclamation.Play(); Start-Sleep -Milliseconds 700")
	default:
		cmd = exec.CommandContext(ctx, "canberra-gtk-play", "-i", "complete")
	}
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("som indisponível: %w", err)
	}
	return nil
}
