//go:build windows

package main

import (
	"syscall"
	"unsafe"
)

const (
	trayUID        = 1
	menuIDOpen     = 1000
	menuIDEnable   = 1001
	menuIDDisable  = 1002
	menuIDExit     = 1003
)

var (
	trayHWnd    uintptr
	iconActive  uintptr
	iconInactive uintptr
	hInstance   uintptr

	// Broadcast when the shell (re)starts; we re-add the tray icon then.
	taskbarCreatedMsg uintptr
)

// createProgrammaticIcon creates a small colored square icon using GDI.
// color is a COLORREF value (0x00BBGGRR).
func createProgrammaticIcon(color uint32) uintptr {
	const size = 16

	screenDC, _, _ := procGetDC.Call(0)
	defer procReleaseDC.Call(0, screenDC)

	memDC, _, _ := procCreateCompatibleDC.Call(screenDC)
	defer procDeleteDC.Call(memDC)

	bmp, _, _ := procCreateCompatibleBitmap.Call(screenDC, size, size)
	procSelectObject.Call(memDC, bmp)

	brush, _, _ := procCreateSolidBrush.Call(uintptr(color))
	rect := RECT{0, 0, size, size}
	procFillRect.Call(memDC, uintptr(unsafe.Pointer(&rect)), brush)
	procDeleteObject.Call(brush)

	// Create a monochrome mask bitmap (all zeros = opaque)
	maskBmp, _, _ := procCreateBitmap.Call(size, size, 1, 1, 0)

	iconInfo := ICONINFO{
		FIcon:   1, // TRUE = icon
		HbmMask: maskBmp,
		HbmColor: bmp,
	}
	icon, _, _ := procCreateIconIndirect.Call(uintptr(unsafe.Pointer(&iconInfo)))

	procDeleteObject.Call(maskBmp)
	procDeleteObject.Call(bmp)

	return icon
}

// trayWndProc handles messages for the hidden tray window.
func trayWndProc(hwnd, msg, wParam, lParam uintptr) uintptr {
	if msg == taskbarCreatedMsg && msg != 0 {
		addTrayIcon()
		return 0
	}

	switch msg {
	case WM_TRAYICON:
		switch lParam {
		case WM_RBUTTONUP:
			showContextMenu(hwnd)
		case WM_LBUTTONUP, WM_LBUTTONDOWN:
			toggleActive()
		}
		return 0

	case WM_COMMAND:
		switch wParam & 0xFFFF {
		case menuIDOpen:
			toggleActive()
		case menuIDEnable:
			setActive(true)
		case menuIDDisable:
			setActive(false)
		case menuIDExit:
			removeTrayIcon()
			uninstallHook()
			procPostQuitMessage.Call(0)
		}
		return 0

	case WM_DESTROY:
		removeTrayIcon()
		procPostQuitMessage.Call(0)
		return 0
	}

	ret, _, _ := procDefWindowProcW.Call(hwnd, msg, wParam, lParam)
	return ret
}

// initTray creates the hidden window, icons, and system tray icon.
func initTray() error {
	hInstance, _, _ = procGetModuleHandleW.Call(0)

	taskbarCreatedMsg, _, _ = procRegisterWindowMessageW.Call(
		uintptr(unsafe.Pointer(utf16Ptr("TaskbarCreated"))))

	// Create icons: green = active (0x0000FF00 -> BGR green), gray = inactive
	iconActive = createProgrammaticIcon(0x0000AA00)  // Green
	iconInactive = createProgrammaticIcon(0x00808080) // Gray

	// Register window class
	className := utf16Ptr("IduMishmiKBTray")
	wc := WNDCLASSEXW{
		CbSize:        uint32(unsafe.Sizeof(WNDCLASSEXW{})),
		LpfnWndProc:   syscall.NewCallback(trayWndProc),
		HInstance:     hInstance,
		LpszClassName: className,
	}
	procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))

	// Create a hidden window for tray icon messages.
	hwnd, _, err := procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(utf16Ptr("Idu Mishmi Keyboard"))),
		0, 0, 0, 0, 0,
		0, // NULL parent (regular hidden window, not HWND_MESSAGE)
		0, hInstance, 0,
	)
	if hwnd == 0 {
		return err
	}
	trayHWnd = hwnd

	// Add tray icon
	addTrayIcon()

	return nil
}

// addTrayIcon adds the icon to the system notification area.
func addTrayIcon() {
	nid := newNotifyIconData()
	ok, _, err := procShellNotifyIconW.Call(NIM_ADD, uintptr(unsafe.Pointer(&nid)))
	if ok == 0 {
		logf("Shell_NotifyIcon NIM_ADD failed: %v", err)
	}
}

// updateTrayIcon changes the tray icon to reflect the current state.
func updateTrayIcon() {
	nid := newNotifyIconData()
	procShellNotifyIconW.Call(NIM_MODIFY, uintptr(unsafe.Pointer(&nid)))
}

// removeTrayIcon removes the icon from the notification area.
func removeTrayIcon() {
	nid := NOTIFYICONDATAW{
		CbSize: uint32(unsafe.Sizeof(NOTIFYICONDATAW{})),
		HWnd:   trayHWnd,
		UID:    trayUID,
	}
	procShellNotifyIconW.Call(NIM_DELETE, uintptr(unsafe.Pointer(&nid)))
}

// newNotifyIconData constructs a NOTIFYICONDATAW with the correct icon and tooltip.
func newNotifyIconData() NOTIFYICONDATAW {
	nid := NOTIFYICONDATAW{
		CbSize:           uint32(unsafe.Sizeof(NOTIFYICONDATAW{})),
		HWnd:             trayHWnd,
		UID:              trayUID,
		UFlags:           NIF_ICON | NIF_TIP | NIF_MESSAGE,
		UCallbackMessage: WM_TRAYICON,
	}

	if isActive() {
		nid.HIcon = iconActive
		copy(nid.SzTip[:], syscall.StringToUTF16("Idu Mishmi Keyboard (Active)"))
	} else {
		nid.HIcon = iconInactive
		copy(nid.SzTip[:], syscall.StringToUTF16("Idu Mishmi Keyboard (Inactive)"))
	}

	return nid
}

// showContextMenu displays the right-click tray menu.
func showContextMenu(hwnd uintptr) {
	menu, _, _ := procCreatePopupMenu.Call()
	if menu == 0 {
		return
	}

	appendMenuItem(menu, menuIDEnable, "Enable")
	appendMenuItem(menu, menuIDDisable, "Disable")
	appendMenuItem(menu, menuIDExit, "Exit")

	var pt POINT
	procGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))

	// Required to make the menu dismiss when clicking elsewhere
	procSetForegroundWindow.Call(hwnd)

	procTrackPopupMenu.Call(
		menu,
		TPM_BOTTOMALIGN|TPM_LEFTALIGN,
		uintptr(pt.X), uintptr(pt.Y),
		0, hwnd, 0,
	)

	procDestroyMenu.Call(menu)
}

// appendMenuItem adds an item to a popup menu.
func appendMenuItem(menu uintptr, id uint32, text string) {
	mi := MENUITEMINFOW{
		CbSize:     uint32(unsafe.Sizeof(MENUITEMINFOW{})),
		FMask:      MIIM_ID | MIIM_TYPE,
		FType:      MFT_STRING,
		WID:        id,
		DwTypeData: utf16Ptr(text),
		Cch:        uint32(len(text)),
	}
	procInsertMenuItemW.Call(menu, 0xFFFFFFFF, 1, uintptr(unsafe.Pointer(&mi)))
}
