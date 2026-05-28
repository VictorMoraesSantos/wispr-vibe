//go:build !windows

package hotkey

import "fmt"

const (
	ModAlt     = 0x0001
	ModControl = 0x0002
	ModShift   = 0x0004
	ModWin     = 0x0008
)

type Listener struct{}

func Register(id int32, modifiers uint32, vk uint32, callback func()) (*Listener, error) {
	return nil, fmt.Errorf("global hotkeys not supported on this platform")
}

func RegisterFromString(id int32, combo string, callback func()) (*Listener, error) {
	return nil, fmt.Errorf("global hotkeys not supported on this platform")
}

func ParseHotkey(combo string) (uint32, uint32, error) {
	return 0, 0, fmt.Errorf("hotkey parsing not supported on this platform")
}

func FormatHotkey(mods uint32, vk uint32) string {
	return "N/A"
}

func (l *Listener) Unregister() {}
