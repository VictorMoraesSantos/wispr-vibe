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
	case "windows":
		return runCmd("clip", text)
	case "darwin":
		return runCmd("pbcopy", text)
	case "linux":
		// Try xclip first, fallback to xsel
		if err := runCmd("xclip", text, "-selection", "clipboard"); err != nil {
			return runCmd("xsel", text, "--clipboard", "--input")
		}
		return nil
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

// Paste simulates Ctrl+V to paste clipboard content.
// On Windows uses PowerShell SendKeys.
func Paste() error {
	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("powershell", "-Command",
			`Add-Type -AssemblyName System.Windows.Forms; [System.Windows.Forms.SendKeys]::SendWait("^v")`)
		return cmd.Run()
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

func runCmd(name string, input string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdin = strings.NewReader(input)
	return cmd.Run()
}
