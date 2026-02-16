//go:build windows

package main

import (
	"syscall"
	"unsafe"
)

const (
	trayUID        = 1
	menuIDEnable   = 1001
	menuIDDisable  = 1002
	menuIDExit     = 1003
)

var (
	trayHWnd    uintptr
	iconActive  uintptr
	iconInactive uintptr
	hInstance   uintptr
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
	switch msg {
	case WM_TRAYICON:
		switch lParam {
		case WM_RBUTTONUP:
			showContextMenu(hwnd)
		case WM_LBUTTONUP:
			toggleActive()
		}
		return 0

	case WM_COMMAND:
		switch wParam & 0xFFFF {
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
	logf("  initTray: GetModuleHandleW...")
	hInstance, _, _ = procGetModuleHandleW.Call(0)
	logf("  initTray: hInstance=%d", hInstance)

	// Create icons: green = active (0x0000FF00 -> BGR green), gray = inactive
	logf("  initTray: creating icons...")
	iconActive = createProgrammaticIcon(0x0000AA00)  // Green
	iconInactive = createProgrammaticIcon(0x00808080) // Gray
	logf("  initTray: icons created (active=%d inactive=%d)", iconActive, iconInactive)

	// Register window class
	logf("  initTray: registering window class...")
	className := utf16Ptr("IduMishmiKBTray")
	wc := WNDCLASSEXW{
		CbSize:        uint32(unsafe.Sizeof(WNDCLASSEXW{})),
		LpfnWndProc:   syscall.NewCallback(trayWndProc),
		HInstance:     hInstance,
		LpszClassName: className,
	}
	atom, _, regErr := procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))
	logf("  initTray: RegisterClassExW atom=%d err=%v", atom, regErr)

	// Create message-only window (HWND_MESSAGE = -3)
	logf("  initTray: creating window...")
	hwnd, _, err := procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(utf16Ptr("Idu Mishmi Keyboard"))),
		0, 0, 0, 0, 0,
		uintptr(0xFFFFFFFFFFFFFFFD), // HWND_MESSAGE
		0, hInstance, 0,
	)
	logf("  initTray: CreateWindowExW hwnd=%d err=%v", hwnd, err)
	if hwnd == 0 {
		return err
	}
	trayHWnd = hwnd

	// Add tray icon
	logf("  initTray: adding tray icon...")
	addTrayIcon()
	logf("  initTray: done")

	return nil
}

// addTrayIcon adds the icon to the system notification area.
func addTrayIcon() {
	nid := newNotifyIconData()
	procShellNotifyIconW.Call(NIM_ADD, uintptr(unsafe.Pointer(&nid)))
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
