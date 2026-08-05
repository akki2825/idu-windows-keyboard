//go:build windows

package main

import (
	"syscall"
	"unsafe"
)

// DLLs
var (
	user32   = syscall.NewLazyDLL("user32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	shell32  = syscall.NewLazyDLL("shell32.dll")
	gdi32    = syscall.NewLazyDLL("gdi32.dll")
)

// user32 procs
var (
	procSetWindowsHookExW   = user32.NewProc("SetWindowsHookExW")
	procUnhookWindowsHookEx = user32.NewProc("UnhookWindowsHookEx")
	procCallNextHookEx      = user32.NewProc("CallNextHookEx")
	procGetMessageW         = user32.NewProc("GetMessageW")
	procTranslateMessage    = user32.NewProc("TranslateMessage")
	procDispatchMessageW    = user32.NewProc("DispatchMessageW")
	procSendInput           = user32.NewProc("SendInput")
	procGetAsyncKeyState    = user32.NewProc("GetAsyncKeyState")
	procGetKeyState         = user32.NewProc("GetKeyState")
	procRegisterClassExW    = user32.NewProc("RegisterClassExW")
	procCreateWindowExW     = user32.NewProc("CreateWindowExW")
	procDefWindowProcW      = user32.NewProc("DefWindowProcW")
	procPostQuitMessage     = user32.NewProc("PostQuitMessage")
	procCreatePopupMenu     = user32.NewProc("CreatePopupMenu")
	procInsertMenuItemW     = user32.NewProc("InsertMenuItemW")
	procTrackPopupMenu      = user32.NewProc("TrackPopupMenu")
	procDestroyMenu         = user32.NewProc("DestroyMenu")
	procSetForegroundWindow = user32.NewProc("SetForegroundWindow")
	procGetCursorPos        = user32.NewProc("GetCursorPos")
	procDestroyIcon         = user32.NewProc("DestroyIcon")
	procCreateIconIndirect  = user32.NewProc("CreateIconIndirect")
	procPostMessageW          = user32.NewProc("PostMessageW")
	procGetKeyboardLayout     = user32.NewProc("GetKeyboardLayout")
	procShowWindow                 = user32.NewProc("ShowWindow")
	procSetWindowTextW             = user32.NewProc("SetWindowTextW")
	procIsWindowVisible            = user32.NewProc("IsWindowVisible")
	procSystemParametersInfoW      = user32.NewProc("SystemParametersInfoW")
	procSendMessageW               = user32.NewProc("SendMessageW")
	procMessageBoxW                = user32.NewProc("MessageBoxW")
	procRegisterWindowMessageW     = user32.NewProc("RegisterWindowMessageW")
	procBeginPaint                 = user32.NewProc("BeginPaint")
	procEndPaint                   = user32.NewProc("EndPaint")
	procGetClientRect              = user32.NewProc("GetClientRect")
	procDrawTextW                  = user32.NewProc("DrawTextW")
	procSetLayeredWindowAttributes = user32.NewProc("SetLayeredWindowAttributes")
)

// shell32 procs
var (
	procShellNotifyIconW = shell32.NewProc("Shell_NotifyIconW")
)

// kernel32 procs
var (
	procGetModuleHandleW = kernel32.NewProc("GetModuleHandleW")
	procCreateMutexW     = kernel32.NewProc("CreateMutexW")
)

// gdi32 procs
var (
	procCreateCompatibleDC     = gdi32.NewProc("CreateCompatibleDC")
	procCreateCompatibleBitmap = gdi32.NewProc("CreateCompatibleBitmap")
	procSelectObject           = gdi32.NewProc("SelectObject")
	procDeleteDC               = gdi32.NewProc("DeleteDC")
	procDeleteObject           = gdi32.NewProc("DeleteObject")
	procCreateSolidBrush       = gdi32.NewProc("CreateSolidBrush")
	procFillRect               = user32.NewProc("FillRect")
	procGetDC                  = user32.NewProc("GetDC")
	procReleaseDC              = user32.NewProc("ReleaseDC")
	procCreateBitmap           = gdi32.NewProc("CreateBitmap")
	procGetStockObject         = gdi32.NewProc("GetStockObject")
	procSetBkMode              = gdi32.NewProc("SetBkMode")
	procSetTextColor           = gdi32.NewProc("SetTextColor")
)

// KBDLLHOOKSTRUCT matches the Win32 KBDLLHOOKSTRUCT layout.
type KBDLLHOOKSTRUCT struct {
	VkCode      uint32
	ScanCode    uint32
	Flags       uint32
	Time        uint32
	DwExtraInfo uintptr
}

// KEYBDINPUT matches the Win32 KEYBDINPUT layout.
type KEYBDINPUT struct {
	WVk         uint16
	WScan       uint16
	DwFlags     uint32
	Time        uint32
	DwExtraInfo uintptr
}

// INPUT matches the Win32 INPUT structure (keyboard variant).
type INPUT struct {
	Type uint32
	Ki   KEYBDINPUT
	_    [8]byte // padding to match union size
}

// NOTIFYICONDATAW matches the Win32 NOTIFYICONDATAW structure.
type NOTIFYICONDATAW struct {
	CbSize           uint32
	HWnd             uintptr
	UID              uint32
	UFlags           uint32
	UCallbackMessage uint32
	HIcon            uintptr
	SzTip            [128]uint16
	DwState          uint32
	DwStateMask      uint32
	SzInfo           [256]uint16
	UVersion         uint32
	SzInfoTitle      [64]uint16
	DwInfoFlags      uint32
	GuidItem         [16]byte
	HBalloonIcon     uintptr
}

// MSG matches the Win32 MSG structure.
type MSG struct {
	HWnd    uintptr
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      POINT
}

// POINT matches the Win32 POINT structure.
type POINT struct {
	X, Y int32
}

// RECT matches the Win32 RECT structure.
type RECT struct {
	Left, Top, Right, Bottom int32
}

// WNDCLASSEXW matches the Win32 WNDCLASSEXW structure.
type WNDCLASSEXW struct {
	CbSize        uint32
	Style         uint32
	LpfnWndProc   uintptr
	CbClsExtra    int32
	CbWndExtra    int32
	HInstance     uintptr
	HIcon         uintptr
	HCursor       uintptr
	HbrBackground uintptr
	LpszMenuName  *uint16
	LpszClassName *uint16
	HIconSm       uintptr
}

// MENUITEMINFOW matches the Win32 MENUITEMINFOW structure.
type MENUITEMINFOW struct {
	CbSize        uint32
	FMask         uint32
	FType         uint32
	FState        uint32
	WID           uint32
	HSubMenu      uintptr
	HbmpChecked   uintptr
	HbmpUnchecked uintptr
	DwItemData    uintptr
	DwTypeData    *uint16
	Cch           uint32
	HbmpItem      uintptr
}

// PAINTSTRUCT matches the Win32 PAINTSTRUCT structure.
type PAINTSTRUCT struct {
	HDC         uintptr
	FErase      int32
	RcPaint     RECT
	FRestore    int32
	FIncUpdate  int32
	RgbReserved [32]byte
}

// ICONINFO matches the Win32 ICONINFO structure.
type ICONINFO struct {
	FIcon    int32
	XHotspot uint32
	YHotspot uint32
	HbmMask  uintptr
	HbmColor uintptr
}

// messageBox shows a message box with no owner window.
func messageBox(text, caption string, flags uintptr) {
	procMessageBoxW.Call(0,
		uintptr(unsafe.Pointer(utf16Ptr(text))),
		uintptr(unsafe.Pointer(utf16Ptr(caption))),
		flags)
}

// utf16Ptr converts a Go string to a *uint16 pointer for Win32 APIs.
func utf16Ptr(s string) *uint16 {
	p, _ := syscall.UTF16PtrFromString(s)
	return p
}

// sizeofInput returns the size of INPUT struct as expected by SendInput.
func sizeofInput() int {
	return int(unsafe.Sizeof(INPUT{}))
}
