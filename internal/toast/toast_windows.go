package toast

import (
	"fmt"
	"runtime"
	"sync"
	"syscall"
	"unsafe"
)

var (
	user32            = syscall.NewLazyDLL("user32.dll")
	kernel32          = syscall.NewLazyDLL("kernel32.dll")
	gdi32             = syscall.NewLazyDLL("gdi32.dll")
	createWindowEx    = user32.NewProc("CreateWindowExW")
	defWindowProc     = user32.NewProc("DefWindowProcW")
	registerClassEx   = user32.NewProc("RegisterClassExW")
	showWindow        = user32.NewProc("ShowWindow")
	updateWindow      = user32.NewProc("UpdateWindow")
	destroyWindow     = user32.NewProc("DestroyWindow")
	getMessage        = user32.NewProc("GetMessageW")
	translateMessage  = user32.NewProc("TranslateMessage")
	dispatchMessage   = user32.NewProc("DispatchMessageW")
	postMessage       = user32.NewProc("PostMessageW")
	getSystemMetrics  = user32.NewProc("GetSystemMetrics")
	setWindowPos      = user32.NewProc("SetWindowPos")
	invalidateRect    = user32.NewProc("InvalidateRect")
	beginPaint        = user32.NewProc("BeginPaint")
	endPaint          = user32.NewProc("EndPaint")
	getModuleHandle   = kernel32.NewProc("GetModuleHandleW")
	createSolidBrush  = gdi32.NewProc("CreateSolidBrush")
	createFontIndirect = gdi32.NewProc("CreateFontIndirectW")
	selectObject      = gdi32.NewProc("SelectObject")
	deleteObject      = gdi32.NewProc("DeleteObject")
	setBkMode         = gdi32.NewProc("SetBkMode")
	setTextColor      = gdi32.NewProc("SetTextColor")
	drawTextW         = user32.NewProc("DrawTextW")
	fillRect          = user32.NewProc("FillRect")
	roundRect         = gdi32.NewProc("RoundRect")
	createPen         = gdi32.NewProc("CreatePen")
	setLayeredWindowAttributes = user32.NewProc("SetLayeredWindowAttributes")
)

const (
	wsExTopmost     = 0x00000008
	wsExToolwindow  = 0x00000080
	wsExLayered     = 0x00080000
	wsExNoactivate  = 0x08000000
	wsPopup         = 0x80000000
	wsVisible       = 0x10000000
	swShow          = 5
	swHide          = 0
	smCxScreen      = 0
	smCyScreen      = 1
	wmDestroy       = 0x0002
	wmPaint         = 0x000F
	wmUser          = 0x0400
	wmClose         = 0x0010
	wmShowToast     = wmUser + 1
	wmHideToast     = wmUser + 2
	wmUpdateText    = wmUser + 3
	csHredraw       = 0x0002
	csVredraw       = 0x0001
	transparent     = 1
	dtCenter        = 0x0001
	dtVcenter       = 0x0004
	dtSingleline    = 0x0020
	lwaAlpha        = 0x0002
	hwndTopmost     = -1
	swpNosize       = 0x0001
	swpNomove       = 0x0002
	swpNoactivate   = 0x0010
	swpShowwindow   = 0x0040
)

type wndClassEx struct {
	size       uint32
	style      uint32
	wndProc    uintptr
	clsExtra   int32
	wndExtra   int32
	instance   uintptr
	icon       uintptr
	cursor     uintptr
	background uintptr
	menuName   *uint16
	className  *uint16
	iconSm     uintptr
}

type point struct{ x, y int32 }
type msg struct {
	hwnd    uintptr
	message uint32
	wParam  uintptr
	lParam  uintptr
	time    uint32
	pt      point
}

type rect struct{ left, top, right, bottom int32 }

type paintStruct struct {
	hdc         uintptr
	fErase      int32
	rcPaint     rect
	fRestore    int32
	fIncUpdate  int32
	rgbReserved [32]byte
}

type logFont struct {
	height         int32
	width          int32
	escapement     int32
	orientation    int32
	weight         int32
	italic         byte
	underline      byte
	strikeOut      byte
	charSet        byte
	outPrecision   byte
	clipPrecision  byte
	quality        byte
	pitchAndFamily byte
	faceName       [32]uint16
}

// Toast is a floating overlay window for recording status.
type Toast struct {
	hwnd    uintptr
	mu      sync.Mutex
	text    string
	visible bool
}

var globalToast *Toast

func wndProc(hwnd, msg_, wParam, lParam uintptr) uintptr {
	switch uint32(msg_) {
	case wmPaint:
		if globalToast != nil {
			globalToast.paint(hwnd)
		}
		return 0
	case wmShowToast:
		showWindow.Call(hwnd, swShow)
		return 0
	case wmHideToast:
		showWindow.Call(hwnd, swHide)
		return 0
	case wmUpdateText:
		invalidateRect.Call(hwnd, 0, 1)
		updateWindow.Call(hwnd)
		return 0
	case wmDestroy:
		return 0
	}
	ret, _, _ := defWindowProc.Call(hwnd, msg_, wParam, lParam)
	return ret
}

func (t *Toast) paint(hwnd uintptr) {
	var ps paintStruct
	hdc, _, _ := beginPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
	if hdc == 0 {
		return
	}
	defer endPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))

	// Background: dark rounded rectangle
	bgColor := rgb(30, 30, 30)
	brush, _, _ := createSolidBrush.Call(uintptr(bgColor))
	pen, _, _ := createPen.Call(0, 1, uintptr(rgb(80, 80, 80)))
	oldBrush, _, _ := selectObject.Call(hdc, brush)
	oldPen, _, _ := selectObject.Call(hdc, pen)

	r := rect{0, 0, toastWidth, toastHeight}
	roundRect.Call(hdc, 0, 0, uintptr(r.right), uintptr(r.bottom), 20, 20)

	selectObject.Call(hdc, oldBrush)
	selectObject.Call(hdc, oldPen)
	deleteObject.Call(brush)
	deleteObject.Call(pen)

	// Font
	var lf logFont
	lf.height = -16
	lf.weight = 400
	copy(lf.faceName[:], utf16("Segoe UI"))
	font, _, _ := createFontIndirect.Call(uintptr(unsafe.Pointer(&lf)))
	oldFont, _, _ := selectObject.Call(hdc, font)

	// Text
	setBkMode.Call(hdc, transparent)
	setTextColor.Call(hdc, uintptr(rgb(255, 255, 255)))

	t.mu.Lock()
	txt := t.text
	t.mu.Unlock()

	txtPtr, _ := syscall.UTF16PtrFromString(txt)
	drawTextW.Call(hdc, uintptr(unsafe.Pointer(txtPtr)), uintptr(len([]rune(txt))),
		uintptr(unsafe.Pointer(&r)), dtCenter|dtVcenter|dtSingleline)

	selectObject.Call(hdc, oldFont)
	deleteObject.Call(font)
}

const (
	toastWidth  = 280
	toastHeight = 44
)

// New creates and starts the toast overlay (hidden initially).
func New() *Toast {
	t := &Toast{
		text: "🎙 Recording...",
	}
	globalToast = t

	ready := make(chan struct{})
	go t.run(ready)
	<-ready
	return t
}

func (t *Toast) run(ready chan struct{}) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	hInstance, _, _ := getModuleHandle.Call(0)
	className, _ := syscall.UTF16PtrFromString("WisprVibe_Toast")

	// Register window class
	wc := wndClassEx{
		size:      uint32(unsafe.Sizeof(wndClassEx{})),
		style:     csHredraw | csVredraw,
		wndProc:   syscall.NewCallback(wndProc),
		instance:  hInstance,
		className: className,
	}
	registerClassEx.Call(uintptr(unsafe.Pointer(&wc)))

	// Get screen size for positioning at bottom-center
	screenW, _, _ := getSystemMetrics.Call(smCxScreen)
	screenH, _, _ := getSystemMetrics.Call(smCyScreen)

	x := (int(screenW) - toastWidth) / 2
	y := int(screenH) - toastHeight - 60 // 60px from bottom

	// Create popup window: topmost, no-activate, layered, tool window
	hwnd, _, _ := createWindowEx.Call(
		wsExTopmost|wsExToolwindow|wsExLayered|wsExNoactivate,
		uintptr(unsafe.Pointer(className)),
		0,
		wsPopup,
		uintptr(x), uintptr(y), toastWidth, toastHeight,
		0, 0, hInstance, 0,
	)

	if hwnd == 0 {
		close(ready)
		return
	}

	t.hwnd = hwnd

	// Set opacity (220/255 ≈ 86%)
	setLayeredWindowAttributes.Call(hwnd, 0, 220, lwaAlpha)

	close(ready)

	// Message loop
	var m msg
	for {
		ret, _, _ := getMessage.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		if ret == 0 || ret == uintptr(^uintptr(0)) {
			break
		}
		translateMessage.Call(uintptr(unsafe.Pointer(&m)))
		dispatchMessage.Call(uintptr(unsafe.Pointer(&m)))
	}
}

// Show makes the toast visible with the given text.
func (t *Toast) Show(text string) {
	t.mu.Lock()
	t.text = text
	t.visible = true
	t.mu.Unlock()

	if t.hwnd != 0 {
		postMessage.Call(t.hwnd, wmShowToast, 0, 0)
		postMessage.Call(t.hwnd, wmUpdateText, 0, 0)
	}
}

// Hide hides the toast.
func (t *Toast) Hide() {
	t.mu.Lock()
	t.visible = false
	t.mu.Unlock()

	if t.hwnd != 0 {
		postMessage.Call(t.hwnd, wmHideToast, 0, 0)
	}
}

// SetText updates the toast text without changing visibility.
func (t *Toast) SetText(text string) {
	t.mu.Lock()
	t.text = text
	t.mu.Unlock()

	if t.hwnd != 0 {
		postMessage.Call(t.hwnd, wmUpdateText, 0, 0)
	}
}

// Destroy cleans up the toast window.
func (t *Toast) Destroy() {
	if t.hwnd != 0 {
		destroyWindow.Call(t.hwnd)
	}
}

func rgb(r, g, b uint8) uint32 {
	return uint32(r) | uint32(g)<<8 | uint32(b)<<16
}

func utf16(s string) []uint16 {
	r, _ := syscall.UTF16FromString(s)
	return r
}

// FormatRecordingText returns the display text for recording state.
func FormatRecordingText(seconds int) string {
	return fmt.Sprintf("🎙 Recording... %ds", seconds)
}
