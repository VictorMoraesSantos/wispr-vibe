package hotkey

import (
	"fmt"
	"runtime"
	"strings"
	"syscall"
	"unsafe"
)

var (
	user32           = syscall.NewLazyDLL("user32.dll")
	registerHotKey   = user32.NewProc("RegisterHotKey")
	unregisterHotKey = user32.NewProc("UnregisterHotKey")
	getMessage       = user32.NewProc("GetMessageW")
	postThreadMsg    = user32.NewProc("PostThreadMessageW")
	kernel32         = syscall.NewLazyDLL("kernel32.dll")
	getCurrentTID    = kernel32.NewProc("GetCurrentThreadId")
)

const (
	ModAlt     = 0x0001
	ModControl = 0x0002
	ModShift   = 0x0004
	ModWin     = 0x0008

	wmHotKey = 0x0312
	wmQuit   = 0x0012
)

type msg struct {
	hwnd    uintptr
	message uint32
	wParam  uintptr
	lParam  uintptr
	time    uint32
	pt      struct{ x, y int32 }
}

// Listener listens for a registered global hotkey.
type Listener struct {
	id       int32
	threadID uint32
	stopCh   chan struct{}
	callback func()
}

// Register registers a global hotkey and calls callback when triggered.
// Both RegisterHotKey and GetMessage run on the SAME locked OS thread.
func Register(id int32, modifiers uint32, vk uint32, callback func()) (*Listener, error) {
	l := &Listener{
		id:       id,
		stopCh:   make(chan struct{}),
		callback: callback,
	}

	errCh := make(chan error, 1)

	go func() {
		// Lock this goroutine to a single OS thread — critical for Win32 message loop
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		// Get thread ID for PostThreadMessage on unregister
		tid, _, _ := getCurrentTID.Call()
		l.threadID = uint32(tid)

		// Register hotkey on THIS thread
		ret, _, err := registerHotKey.Call(0, uintptr(id), uintptr(modifiers), uintptr(vk))
		if ret == 0 {
			errCh <- fmt.Errorf("RegisterHotKey failed: %w", err)
			return
		}
		errCh <- nil

		// Message loop on same thread — receives WM_HOTKEY
		var m msg
		for {
			ret, _, _ := getMessage.Call(
				uintptr(unsafe.Pointer(&m)),
				0, 0, 0,
			)
			if ret == 0 || ret == uintptr(^uintptr(0)) {
				break
			}
			if m.message == wmHotKey && int32(m.wParam) == id {
				l.callback()
			}
			if m.message == wmQuit {
				break
			}
		}

		unregisterHotKey.Call(0, uintptr(id))
	}()

	if err := <-errCh; err != nil {
		return nil, err
	}
	return l, nil
}

// RegisterFromString parses a hotkey string like "Ctrl+Shift+R" and registers it.
func RegisterFromString(id int32, combo string, callback func()) (*Listener, error) {
	mods, vk, err := ParseHotkey(combo)
	if err != nil {
		return nil, err
	}
	return Register(id, mods, vk, callback)
}

// Unregister removes the hotkey and stops the listener.
func (l *Listener) Unregister() {
	// Post WM_QUIT to the listener thread to break the message loop
	if l.threadID != 0 {
		postThreadMsg.Call(uintptr(l.threadID), wmQuit, 0, 0)
	}
}

// ParseHotkey parses "Ctrl+Shift+R" → (modifiers, virtualKey, error)
func ParseHotkey(combo string) (uint32, uint32, error) {
	parts := strings.Split(strings.TrimSpace(combo), "+")
	if len(parts) == 0 {
		return 0, 0, fmt.Errorf("empty hotkey")
	}

	var mods uint32
	var key string

	for _, p := range parts {
		p = strings.TrimSpace(p)
		switch strings.ToLower(p) {
		case "ctrl", "control":
			mods |= ModControl
		case "shift":
			mods |= ModShift
		case "alt":
			mods |= ModAlt
		case "win", "super":
			mods |= ModWin
		default:
			key = p
		}
	}

	if key == "" {
		return 0, 0, fmt.Errorf("no key specified in hotkey: %s", combo)
	}

	vk, ok := keyNameToVK(key)
	if !ok {
		return 0, 0, fmt.Errorf("unknown key: %s", key)
	}

	return mods, vk, nil
}

// FormatHotkey converts modifiers + VK back to display string.
func FormatHotkey(mods uint32, vk uint32) string {
	var parts []string
	if mods&ModControl != 0 {
		parts = append(parts, "Ctrl")
	}
	if mods&ModAlt != 0 {
		parts = append(parts, "Alt")
	}
	if mods&ModShift != 0 {
		parts = append(parts, "Shift")
	}
	if mods&ModWin != 0 {
		parts = append(parts, "Win")
	}
	parts = append(parts, vkToKeyName(vk))
	return strings.Join(parts, "+")
}

// keyNameToVK maps key names to Windows virtual key codes.
func keyNameToVK(name string) (uint32, bool) {
	upper := strings.ToUpper(strings.TrimSpace(name))

	// Single letter A-Z
	if len(upper) == 1 && upper[0] >= 'A' && upper[0] <= 'Z' {
		return uint32(upper[0]), true
	}

	// Single digit 0-9
	if len(upper) == 1 && upper[0] >= '0' && upper[0] <= '9' {
		return uint32(upper[0]), true
	}

	// Function keys and special keys
	special := map[string]uint32{
		"F1": 0x70, "F2": 0x71, "F3": 0x72, "F4": 0x73,
		"F5": 0x74, "F6": 0x75, "F7": 0x76, "F8": 0x77,
		"F9": 0x78, "F10": 0x79, "F11": 0x7A, "F12": 0x7B,
		"SPACE": 0x20, "ENTER": 0x0D, "RETURN": 0x0D,
		"TAB": 0x09, "ESC": 0x1B, "ESCAPE": 0x1B,
		"BACKSPACE": 0x08, "DELETE": 0x2E, "DEL": 0x2E,
		"INSERT": 0x2D, "INS": 0x2D,
		"HOME": 0x24, "END": 0x23,
		"PAGEUP": 0x21, "PGUP": 0x21,
		"PAGEDOWN": 0x22, "PGDN": 0x22,
		"UP": 0x26, "DOWN": 0x28, "LEFT": 0x25, "RIGHT": 0x27,
		"CAPSLOCK": 0x14, "NUMLOCK": 0x90, "SCROLLLOCK": 0x91,
		"PRINTSCREEN": 0x2C, "PRTSC": 0x2C,
		"PAUSE": 0x13, "BREAK": 0x13,
		"`": 0xC0, "~": 0xC0, "GRAVE": 0xC0,
		"-": 0xBD, "MINUS": 0xBD,
		"=": 0xBB, "EQUALS": 0xBB, "PLUS": 0xBB,
		"[": 0xDB, "]": 0xDD,
		"\\": 0xDC, ";": 0xBA, "'": 0xDE,
		",": 0xBC, ".": 0xBE, "/": 0xBF,
	}

	if vk, ok := special[upper]; ok {
		return vk, true
	}

	return 0, false
}

// vkToKeyName converts VK code back to display name.
func vkToKeyName(vk uint32) string {
	if vk >= 'A' && vk <= 'Z' {
		return string(rune(vk))
	}
	if vk >= '0' && vk <= '9' {
		return string(rune(vk))
	}

	names := map[uint32]string{
		0x70: "F1", 0x71: "F2", 0x72: "F3", 0x73: "F4",
		0x74: "F5", 0x75: "F6", 0x76: "F7", 0x77: "F8",
		0x78: "F9", 0x79: "F10", 0x7A: "F11", 0x7B: "F12",
		0x20: "Space", 0x0D: "Enter", 0x09: "Tab", 0x1B: "Esc",
		0x08: "Backspace", 0x2E: "Delete", 0x2D: "Insert",
		0x24: "Home", 0x23: "End", 0x21: "PageUp", 0x22: "PageDown",
		0x26: "Up", 0x28: "Down", 0x25: "Left", 0x27: "Right",
	}

	if name, ok := names[vk]; ok {
		return name
	}
	return fmt.Sprintf("0x%02X", vk)
}
