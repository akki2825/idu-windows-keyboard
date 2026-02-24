//go:build windows

package main

import (
	_ "embed"
	"os"
	"path/filepath"
	"syscall"
	"unsafe"
)

//go:embed fonts/NotoSans-Regular.ttf
var notoSansTTF []byte

// advapi32 DLL for registry operations.
var advapi32 = syscall.NewLazyDLL("advapi32.dll")

var (
	procRegOpenKeyExW   = advapi32.NewProc("RegOpenKeyExW")
	procRegSetValueExW  = advapi32.NewProc("RegSetValueExW")
	procRegCloseKey     = advapi32.NewProc("RegCloseKey")
	procAddFontResourceW = gdi32.NewProc("AddFontResourceW")
)

// Registry constants.
const (
	HKEY_CURRENT_USER uintptr = 0x80000001
	KEY_SET_VALUE     uint32  = 0x0002
	REG_SZ            uint32  = 1
)

// Font broadcast constants.
const (
	WM_FONTCHANGE  = 0x001D
	HWND_BROADCAST = 0xFFFF
)

// ensureNotoSansInstalled installs Noto Sans as a per-user font if not already present.
// All errors are logged but non-fatal — font install failure never prevents the keyboard from working.
func ensureNotoSansInstalled() {
	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData == "" {
		logf("font: LOCALAPPDATA not set, skipping font install")
		return
	}

	fontDir := filepath.Join(localAppData, "Microsoft", "Windows", "Fonts")
	fontPath := filepath.Join(fontDir, "NotoSans-Regular.ttf")

	// Check if font file already exists.
	if _, err := os.Stat(fontPath); err == nil {
		logf("font: %s already exists, skipping write", fontPath)
	} else {
		// Create font directory if needed.
		if err := os.MkdirAll(fontDir, 0755); err != nil {
			logf("font: failed to create directory %s: %v", fontDir, err)
			return
		}

		// Write embedded font to disk.
		if err := os.WriteFile(fontPath, notoSansTTF, 0644); err != nil {
			logf("font: failed to write %s: %v", fontPath, err)
			return
		}
		logf("font: wrote %s (%d bytes)", fontPath, len(notoSansTTF))

		// Register in HKCU fonts registry.
		registerFontInRegistry(fontPath)
	}

	// Load font for current session.
	fontPathPtr := utf16Ptr(fontPath)
	ret, _, _ := procAddFontResourceW.Call(uintptr(unsafe.Pointer(fontPathPtr)))
	if ret == 0 {
		logf("font: AddFontResourceW failed")
	} else {
		logf("font: AddFontResourceW succeeded (added %d fonts)", ret)
	}

	// Notify other applications of the font change.
	procPostMessageW.Call(HWND_BROADCAST, WM_FONTCHANGE, 0, 0)
	logf("font: broadcast WM_FONTCHANGE")
}

// registerFontInRegistry adds the font to HKCU\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Fonts.
func registerFontInRegistry(fontPath string) {
	subKey := utf16Ptr(`SOFTWARE\Microsoft\Windows NT\CurrentVersion\Fonts`)
	var hKey uintptr
	ret, _, _ := procRegOpenKeyExW.Call(
		HKEY_CURRENT_USER,
		uintptr(unsafe.Pointer(subKey)),
		0,
		uintptr(KEY_SET_VALUE),
		uintptr(unsafe.Pointer(&hKey)),
	)
	if ret != 0 {
		logf("font: RegOpenKeyExW failed (error %d)", ret)
		return
	}
	defer procRegCloseKey.Call(hKey)

	valueName := utf16Ptr("Noto Sans Regular (TrueType)")
	valueData, _ := syscall.UTF16FromString(fontPath)
	dataSize := uint32(len(valueData) * 2) // UTF-16 byte count

	ret, _, _ = procRegSetValueExW.Call(
		hKey,
		uintptr(unsafe.Pointer(valueName)),
		0,
		uintptr(REG_SZ),
		uintptr(unsafe.Pointer(&valueData[0])),
		uintptr(dataSize),
	)
	if ret != 0 {
		logf("font: RegSetValueExW failed (error %d)", ret)
		return
	}
	logf("font: registered in HKCU fonts registry")
}
