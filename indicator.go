//go:build windows

package main

import (
	"syscall"
	"unsafe"
)

const (
	indicatorWidth  = 46
	indicatorHeight = 24
)

var indicatorHWnd uintptr

// indicatorWndProc paints the "IDU" badge.
func indicatorWndProc(hwnd, msg, wParam, lParam uintptr) uintptr {
	switch msg {
	case WM_PAINT:
		var ps PAINTSTRUCT
		hdc, _, _ := procBeginPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))

		var rc RECT
		procGetClientRect.Call(hwnd, uintptr(unsafe.Pointer(&rc)))

		// Dark green background.
		brush, _, _ := procCreateSolidBrush.Call(0x00008B00) // BGR dark green
		procFillRect.Call(hdc, uintptr(unsafe.Pointer(&rc)), brush)
		procDeleteObject.Call(brush)

		// White "IDU" text, centered.
		procSetBkMode.Call(hdc, BK_TRANSPARENT)
		procSetTextColor.Call(hdc, 0x00FFFFFF) // White
		guiFont, _, _ := procGetStockObject.Call(DEFAULT_GUI_FONT)
		procSelectObject.Call(hdc, guiFont)

		text := utf16Ptr("IDU")
		procDrawTextW.Call(hdc, uintptr(unsafe.Pointer(text)),
			uintptr(0xFFFFFFFF), // -1 = null-terminated
			uintptr(unsafe.Pointer(&rc)),
			DT_CENTER|DT_VCENTER|DT_SINGLELINE)

		procEndPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
		return 0
	}
	ret, _, _ := procDefWindowProcW.Call(hwnd, msg, wParam, lParam)
	return ret
}

// initIndicator creates the floating badge window (hidden initially).
func initIndicator() error {
	className := utf16Ptr("IduMishmiKBIndicator")
	wc := WNDCLASSEXW{
		CbSize:        uint32(unsafe.Sizeof(WNDCLASSEXW{})),
		LpfnWndProc:   syscall.NewCallback(indicatorWndProc),
		HInstance:     hInstance,
		LpszClassName: className,
	}
	procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))

	var workArea RECT
	procSystemParametersInfoW.Call(SPI_GETWORKAREA, 0, uintptr(unsafe.Pointer(&workArea)), 0)
	x := workArea.Right - indicatorWidth - 8
	y := workArea.Bottom - indicatorHeight - 8

	hwnd, _, err := procCreateWindowExW.Call(
		WS_EX_TOPMOST|WS_EX_TOOLWINDOW|WS_EX_LAYERED|WS_EX_TRANSPARENT,
		uintptr(unsafe.Pointer(className)),
		0,
		WS_POPUP,
		uintptr(x), uintptr(y),
		indicatorWidth, indicatorHeight,
		0, 0, hInstance, 0,
	)
	if hwnd == 0 {
		return err
	}
	indicatorHWnd = hwnd

	// 85% opacity.
	procSetLayeredWindowAttributes.Call(hwnd, 0, 216, LWA_ALPHA)

	updateIndicator()
	return nil
}

// updateIndicator shows or hides the badge based on the active state.
func updateIndicator() {
	if indicatorHWnd == 0 {
		return
	}
	if isActive() {
		procShowWindow.Call(indicatorHWnd, SW_SHOWNOACTIVATE)
	} else {
		procShowWindow.Call(indicatorHWnd, SW_HIDE)
	}
}
