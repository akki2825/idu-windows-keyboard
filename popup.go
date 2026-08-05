//go:build windows

package main

import (
	"syscall"
	"unsafe"
)

const (
	popupWidth  = 280
	popupHeight = 140

	ctrlIDStatus = 2001
	ctrlIDToggle = 2002
)

var (
	popupHWnd  uintptr
	statusHWnd uintptr
	toggleHWnd uintptr
)

// popupWndProc handles messages for the popup settings window.
func popupWndProc(hwnd, msg, wParam, lParam uintptr) uintptr {
	switch msg {
	case WM_COMMAND:
		if wParam&0xFFFF == ctrlIDToggle {
			toggleActive()
		}
		return 0
	case WM_CLOSE:
		// Hide instead of destroy so it can be reopened.
		procShowWindow.Call(hwnd, SW_HIDE)
		return 0
	}
	ret, _, _ := procDefWindowProcW.Call(hwnd, msg, wParam, lParam)
	return ret
}

// initPopup creates the popup window (hidden initially) with a status label and toggle button.
func initPopup() error {
	className := utf16Ptr("IduMishmiKBPopup")
	wc := WNDCLASSEXW{
		CbSize:        uint32(unsafe.Sizeof(WNDCLASSEXW{})),
		LpfnWndProc:   syscall.NewCallback(popupWndProc),
		HInstance:     hInstance,
		HbrBackground: COLOR_BTNFACE + 1,
		LpszClassName: className,
	}
	procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))

	// Position near the bottom-right of the work area (above the taskbar).
	var workArea RECT
	procSystemParametersInfoW.Call(SPI_GETWORKAREA, 0, uintptr(unsafe.Pointer(&workArea)), 0)
	x := workArea.Right - popupWidth - 10
	y := workArea.Bottom - popupHeight - 10

	hwnd, _, err := procCreateWindowExW.Call(
		WS_EX_TOPMOST|WS_EX_TOOLWINDOW,
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(utf16Ptr("Idu Mishmi Keyboard"))),
		WS_POPUP|WS_CAPTION|WS_SYSMENU,
		uintptr(x), uintptr(y),
		popupWidth, popupHeight,
		0, 0, hInstance, 0,
	)
	if hwnd == 0 {
		return err
	}
	popupHWnd = hwnd

	guiFont, _, _ := procGetStockObject.Call(DEFAULT_GUI_FONT)

	// Status label (centered).
	statusHWnd, _, _ = procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(utf16Ptr("STATIC"))),
		0,
		WS_CHILD|WS_VISIBLE|SS_CENTER,
		20, 20, 240, 25,
		hwnd, ctrlIDStatus, hInstance, 0,
	)
	procSendMessageW.Call(statusHWnd, WM_SETFONT, guiFont, 1)

	// Toggle button (centered).
	toggleHWnd, _, _ = procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(utf16Ptr("BUTTON"))),
		0,
		WS_CHILD|WS_VISIBLE|BS_PUSHBUTTON,
		80, 60, 120, 35,
		hwnd, ctrlIDToggle, hInstance, 0,
	)
	procSendMessageW.Call(toggleHWnd, WM_SETFONT, guiFont, 1)

	updatePopup()
	return nil
}

// updatePopup refreshes the status label and button text to reflect the current state.
func updatePopup() {
	if popupHWnd == 0 {
		return
	}
	if isActive() {
		procSetWindowTextW.Call(statusHWnd, uintptr(unsafe.Pointer(utf16Ptr("Keyboard is active"))))
		procSetWindowTextW.Call(toggleHWnd, uintptr(unsafe.Pointer(utf16Ptr("Disable"))))
	} else {
		procSetWindowTextW.Call(statusHWnd, uintptr(unsafe.Pointer(utf16Ptr("Keyboard is inactive"))))
		procSetWindowTextW.Call(toggleHWnd, uintptr(unsafe.Pointer(utf16Ptr("Enable"))))
	}
}

// showPopup toggles visibility of the popup window.
func showPopup() {
	if popupHWnd == 0 {
		return
	}
	visible, _, _ := procIsWindowVisible.Call(popupHWnd)
	if visible != 0 {
		procShowWindow.Call(popupHWnd, SW_HIDE)
	} else {
		updatePopup()
		procShowWindow.Call(popupHWnd, SW_SHOW)
		procSetForegroundWindow.Call(popupHWnd)
	}
}
