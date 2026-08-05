//go:build windows

package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// logFile is the debug log in %LOCALAPPDATA%\IduMishmiKeyboard.
// It stays nil (and logf no-ops) if the log cannot be opened.
var (
	logFile *os.File
	logPID  int
)

func initLog() {
	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData == "" {
		return
	}
	logDir := filepath.Join(localAppData, "IduMishmiKeyboard")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return
	}
	// Append so a second instance doesn't wipe the first instance's log.
	f, err := os.OpenFile(filepath.Join(logDir, "idu-keyboard.log"),
		os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return
	}
	logFile = f
	logPID = os.Getpid()
}

func logf(format string, args ...any) {
	if logFile != nil {
		fmt.Fprintf(logFile, "[pid %d] "+format+"\n", append([]any{logPID}, args...)...)
		logFile.Sync()
	}
}
