package hotkey

import (
	"fmt"
	"syscall"
	"unsafe"
)

var (
	user32          = syscall.NewLazyDLL("user32.dll")
	registerHotKey  = user32.NewProc("RegisterHotKey")
	unregisterHotKey = user32.NewProc("UnregisterHotKey")
	getMessage      = user32.NewProc("GetMessageW")
)

const (
	ModAlt     = 0x0001
	ModControl = 0x0002
	ModShift   = 0x0004
	ModWin     = 0x0008

	wmHotKey = 0x0312
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
	stopCh   chan struct{}
	callback func()
}

// Register registers a global hotkey and calls callback when triggered.
// modifiers: combination of ModAlt, ModControl, ModShift, ModWin
// vk: virtual key code (e.g., 0x52 = 'R')
func Register(id int32, modifiers uint32, vk uint32, callback func()) (*Listener, error) {
	ret, _, err := registerHotKey.Call(0, uintptr(id), uintptr(modifiers), uintptr(vk))
	if ret == 0 {
		return nil, fmt.Errorf("RegisterHotKey failed: %w", err)
	}

	l := &Listener{
		id:       id,
		stopCh:   make(chan struct{}),
		callback: callback,
	}

	go l.listen()
	return l, nil
}

func (l *Listener) listen() {
	var m msg
	for {
		select {
		case <-l.stopCh:
			return
		default:
		}
		// PeekMessage with PM_REMOVE to avoid blocking forever
		ret, _, _ := getMessage.Call(
			uintptr(unsafe.Pointer(&m)),
			0, 0, 0,
		)
		if ret == 0 {
			return
		}
		if m.message == wmHotKey && int32(m.wParam) == l.id {
			l.callback()
		}
	}
}

// Unregister removes the hotkey.
func (l *Listener) Unregister() {
	close(l.stopCh)
	unregisterHotKey.Call(0, uintptr(l.id))
}
