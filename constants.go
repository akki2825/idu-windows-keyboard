package main

// Hook constants
const (
	WH_KEYBOARD_LL = 13
	WM_KEYDOWN     = 0x0100
	WM_KEYUP       = 0x0101
	WM_SYSKEYDOWN  = 0x0104
	WM_SYSKEYUP    = 0x0105
)

// Virtual key codes
const (
	VK_SHIFT    = 0x10
	VK_CONTROL  = 0x11
	VK_MENU     = 0x12 // Alt
	VK_LSHIFT   = 0xA0
	VK_RSHIFT   = 0xA1
	VK_LCONTROL = 0xA2
	VK_RCONTROL = 0xA3
	VK_LMENU    = 0xA4
	VK_RMENU    = 0xA5 // Right Alt = AltGr
	VK_CAPITAL  = 0x14 // Caps Lock
	VK_SCROLL   = 0x91 // Scroll Lock
	VK_LWIN     = 0x5B
	VK_RWIN     = 0x5C

	// Letter keys
	VK_A = 0x41
	VK_E = 0x45
	VK_I = 0x49
	VK_M = 0x4D
	VK_O = 0x4F
	VK_R = 0x52
	VK_U = 0x55
	VK_W = 0x57

	// OEM keys
	VK_OEM_1      = 0xBA // ;
	VK_OEM_3      = 0xC0 // ` / ~
	VK_OEM_7      = 0xDE // ' / "
	VK_OEM_PLUS   = 0xBB
	VK_OEM_COMMA  = 0xBC
	VK_OEM_MINUS  = 0xBD // - / _
	VK_OEM_PERIOD = 0xBE
)

// KBDLLHOOKSTRUCT flags
const (
	LLKHF_EXTENDED = 0x01
	LLKHF_INJECTED = 0x10
	LLKHF_ALTDOWN  = 0x20
	LLKHF_UP       = 0x80
)

// SendInput constants
const (
	INPUT_KEYBOARD    = 1
	KEYEVENTF_KEYUP   = 0x0002
	KEYEVENTF_UNICODE = 0x0004
)

// Window messages
const (
	WM_DESTROY   = 0x0002
	WM_COMMAND   = 0x0111
	WM_LBUTTONDOWN = 0x0201
	WM_LBUTTONUP   = 0x0202
	WM_RBUTTONUP   = 0x0205
	WM_USER      = 0x0400
	WM_TRAYICON  = WM_USER + 1
)

// Shell_NotifyIcon constants
const (
	NIM_ADD     = 0x00000000
	NIM_MODIFY  = 0x00000001
	NIM_DELETE  = 0x00000002
	NIF_ICON    = 0x00000002
	NIF_TIP     = 0x00000004
	NIF_MESSAGE = 0x00000001
)

// Menu constants
const (
	TPM_BOTTOMALIGN = 0x0020
	TPM_LEFTALIGN   = 0x0000
	MFT_STRING      = 0x00000000
	MIIM_ID         = 0x00000002
	MIIM_TYPE       = 0x00000010
)

// Window styles
const (
	WS_POPUP   = 0x80000000
	WS_CAPTION = 0x00C00000
	WS_SYSMENU = 0x00080000
	WS_CHILD   = 0x40000000
	WS_VISIBLE = 0x10000000
)

// Extended window styles
const (
	WS_EX_TOPMOST      = 0x00000008
	WS_EX_TRANSPARENT  = 0x00000020
	WS_EX_TOOLWINDOW   = 0x00000080
	WS_EX_LAYERED      = 0x00080000
)

// Control styles
const (
	SS_CENTER     = 0x00000001
	BS_PUSHBUTTON = 0x00000000
)

// ShowWindow commands
const (
	SW_HIDE           = 0
	SW_SHOWNOACTIVATE = 4
	SW_SHOW           = 5
)

// Additional messages
const (
	WM_CLOSE   = 0x0010
	WM_SETFONT = 0x0030
)

// SystemParametersInfo
const (
	SPI_GETWORKAREA = 0x0030
)

// Colors
const (
	COLOR_BTNFACE = 15
)

// Stock objects
const (
	DEFAULT_GUI_FONT = 17
)

// Drawing
const (
	BK_TRANSPARENT = 1 // SetBkMode
	DT_CENTER      = 0x01
	DT_VCENTER     = 0x04
	DT_SINGLELINE  = 0x20
	WM_PAINT       = 0x000F
)

// Layered window
const (
	LWA_ALPHA = 0x02
)

// Error constants
const (
	ERROR_ALREADY_EXISTS = 183
)

// MessageBox flags
const (
	MB_OK              = 0x00000000
	MB_ICONERROR       = 0x00000010
	MB_ICONINFORMATION = 0x00000040
)

// isIgnorableModifier returns true if the pressed key is a modifier key that
// we should always pass through without processing.
func isIgnorableModifier(vkCode uint32) bool {
	switch vkCode {
	case VK_SHIFT, VK_CONTROL, VK_MENU,
		VK_LSHIFT, VK_RSHIFT,
		VK_LCONTROL, VK_RCONTROL,
		VK_LMENU, VK_RMENU,
		VK_LWIN, VK_RWIN,
		VK_CAPITAL:
		return true
	}
	return false
}
