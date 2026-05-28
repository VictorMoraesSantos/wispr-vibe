package clipboard

import (
	"syscall"
	"time"
	"unsafe"
)

var (
	user32         = syscall.NewLazyDLL("user32.dll")
	sendInput      = user32.NewProc("SendInput")
	openClipboard  = user32.NewProc("OpenClipboard")
	closeClipboard = user32.NewProc("CloseClipboard")
	emptyClipboard = user32.NewProc("EmptyClipboard")
	setClipboard   = user32.NewProc("SetClipboardData")

	kernel32     = syscall.NewLazyDLL("kernel32.dll")
	globalAlloc  = kernel32.NewProc("GlobalAlloc")
	globalLock   = kernel32.NewProc("GlobalLock")
	globalUnlock = kernel32.NewProc("GlobalUnlock")
)

const (
	inputKeyboard = 1
	keyEventKeyUp = 0x0002
	vkControl     = 0x11
	vkV           = 0x56
	cfUnicodeText = 13
	gmemMoveable  = 0x0002
)

type keyboardInput struct {
	wVk         uint16
	wScan       uint16
	dwFlags     uint32
	time        uint32
	dwExtraInfo uintptr
}

type input struct {
	inputType uint32
	ki        keyboardInput
	padding   [8]byte
}

func CopyToClipboard(text string) error {
	utf16, err := syscall.UTF16FromString(text)
	if err != nil {
		return err
	}

	size := len(utf16) * 2
	hMem, _, _ := globalAlloc.Call(gmemMoveable, uintptr(size))
	if hMem == 0 {
		return syscall.GetLastError()
	}

	lock, _, _ := globalLock.Call(hMem)
	if lock == 0 {
		return syscall.GetLastError()
	}

	dst := unsafe.Slice((*uint16)(unsafe.Pointer(lock)), len(utf16))
	copy(dst, utf16)
	globalUnlock.Call(hMem)

	r, _, err := openClipboard.Call(0)
	if r == 0 {
		return err
	}
	defer closeClipboard.Call()

	emptyClipboard.Call()
	r, _, err = setClipboard.Call(cfUnicodeText, hMem)
	if r == 0 {
		return err
	}

	return nil
}

func PasteToActiveWindow() error {
	time.Sleep(50 * time.Millisecond)

	inputs := make([]input, 4)

	inputs[0].inputType = inputKeyboard
	inputs[0].ki.wVk = vkControl

	inputs[1].inputType = inputKeyboard
	inputs[1].ki.wVk = vkV

	inputs[2].inputType = inputKeyboard
	inputs[2].ki.wVk = vkV
	inputs[2].ki.dwFlags = keyEventKeyUp

	inputs[3].inputType = inputKeyboard
	inputs[3].ki.wVk = vkControl
	inputs[3].ki.dwFlags = keyEventKeyUp

	ret, _, err := sendInput.Call(
		uintptr(len(inputs)),
		uintptr(unsafe.Pointer(&inputs[0])),
		uintptr(unsafe.Sizeof(inputs[0])),
	)

	if ret == 0 {
		return err
	}
	return nil
}

func TypeText(text string) error {
	if err := CopyToClipboard(text); err != nil {
		return err
	}
	time.Sleep(50 * time.Millisecond)
	return PasteToActiveWindow()
}
