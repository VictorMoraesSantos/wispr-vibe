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

// Unregister is a no-op.
func (l *Listener) Unregister() {}
