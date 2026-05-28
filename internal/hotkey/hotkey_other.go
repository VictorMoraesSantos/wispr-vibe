//go:build !windows

package hotkey

import "fmt"

const (
	ModAlt     = 0x0001
	ModControl = 0x0002
	ModShift   = 0x0004
	ModWin     = 0x0008
)

// Listener is a no-op on non-Windows platforms.
type Listener struct{}

// Register is not supported on this platform.
func Register(id int32, modifiers uint32, vk uint32, callback func()) (*Listener, error) {
	return nil, fmt.Errorf("global hotkeys not supported on this platform")
}

// RegisterFromString is not supported on this platform.
func RegisterFromString(id int32, combo string, callback func()) (*Listener, error) {
	return nil, fmt.Errorf("global hotkeys not supported on this platform")
}

// ParseHotkey parses a hotkey string (works on all platforms for config validation).
func ParseHotkey(combo string) (uint32, uint32, error) {
	return 0, 0, fmt.Errorf("hotkey parsing not supported on this platform")
}

// FormatHotkey returns the combo string as-is on non-Windows.
func FormatHotkey(mods uint32, vk uint32) string {
	return "N/A"
}

// Unregister is a no-op.
func (l *Listener) Unregister() {}
