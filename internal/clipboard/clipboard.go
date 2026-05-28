//go:build !windows

package clipboard

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// Copy places text on the system clipboard.
func Copy(text string) error {
	switch runtime.GOOS {
	case "darwin":
		return runCmd("pbcopy", text)
	case "linux":
		if err := runCmd("xclip", text, "-selection", "clipboard"); err != nil {
			return runCmd("xsel", text, "--clipboard", "--input")
		}
		return nil
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

// Paste simulates Ctrl+V to paste clipboard content.
func Paste() error {
	switch runtime.GOOS {
	case "darwin":
		cmd := exec.Command("osascript", "-e",
			`tell application "System Events" to keystroke "v" using command down`)
		return cmd.Run()
	case "linux":
		cmd := exec.Command("xdotool", "key", "ctrl+v")
		return cmd.Run()
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

// TypeText copies text and pastes into active window.
func TypeText(text string) error {
	if err := Copy(text); err != nil {
		return err
	}
	return Paste()
}

// CopyToClipboard is an alias for Copy on non-Windows.
func CopyToClipboard(text string) error {
	return Copy(text)
}

// PasteToActiveWindow is an alias for Paste on non-Windows.
func PasteToActiveWindow() error {
	return Paste()
}

func runCmd(name string, input string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdin = strings.NewReader(input)
	return cmd.Run()
}
