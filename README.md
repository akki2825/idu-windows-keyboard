# Idu Mishmi Keyboard for Windows

The official Windows desktop keyboard for the Idu Mishmi language. A standalone, portable program that lets you type the full Idu Mishmi character set (Idu Azobra) in any Windows application.

Does not require installation or administrator access. Works fully offline with no internet connection needed.

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [How to Use](#how-to-use)
- [How It Works](#how-it-works)
- [Building from Source](#building-from-source)
- [License](#license)

# Features

## Idu Mishmi Language Support
<ul>
  <li>🎯 <strong>Full Idu Azobra character set</strong> including schwa (ə), retracted vowels (ə̱, o̱, u̱), and all accented forms</li>
  <li>⌨️ <strong>AltGr-based input</strong> — use the Right Alt key to access special characters</li>
  <li>🔤 <strong>Dead key composition</strong> — type accented vowels in two steps (accent key, then vowel)</li>
  <li>✅ <strong>Works with any app</strong> — Notepad, Word, Chrome, or any other Windows program</li>
  <li>🔒 <strong>Fully offline</strong> — no internet connection required, no data leaves your computer</li>
  <li>📦 <strong>Portable</strong> — single .exe file, no installation or admin access needed</li>
  <li>🤝 <strong>Coexists with other keyboards</strong> — does not interfere with Hindi, English, or other input methods</li>
</ul>

# Installation

1. Download `Idu Mishmi Keyboard.exe` from the [latest release](https://github.com/akki2825/idu-keyboard-windows/releases/latest)
2. Place it anywhere on your computer
3. Double-click to run

The keyboard runs in the system tray. Only one instance can run at a time.

# How to Use

Open any app where you can type and use the **Right Alt key (AltGr)** for special characters.

## Direct Keys

| Shortcut | Output | Description |
|---|---|---|
| Right Alt + E | ə | Schwa |
| Right Alt + Shift + E | Ə | Capital schwa |
| Right Alt + R | ə̱ | Retracted schwa |
| Right Alt + Shift + R | Ə̱ | Capital retracted schwa |
| Right Alt + O | o̱ | Retracted o |
| Right Alt + Shift + O | O̱ | Capital retracted o |
| Right Alt + U | u̱ | Retracted u |
| Right Alt + Shift + U | U̱ | Capital retracted u |

## Accented Vowels (two-step)

Press the accent key first, release, then press a vowel (a, e, i, o, or u):

| Accent Key | Then Vowel | Output |
|---|---|---|
| Right Alt + `` ` `` | a e i o u | à è ì ò ù (grave) |
| Right Alt + `'` | a e i o u | á é í ó ú (acute) |
| Right Alt + `~` | a e i o u | ã ẽ ĩ õ ũ (nasalized) |
| Right Alt + `-` | a e i o u | ā ē ī ō ū (macron) |

## Accented Schwas (two-step)

Press the accent key first, then Right Alt + E (for schwa) or Right Alt + R (for retracted schwa):

| Accent Key | Then Right Alt + E | Then Right Alt + R |
|---|---|---|
| Right Alt + `` ` `` | ə̀ | ə̱̀ |
| Right Alt + `'` | ə́ | ə̱́ |
| Right Alt + `~` | ə̃ | ə̱̃ |
| Right Alt + `-` | ə̄ | ə̱̄ |

Hold **Shift** for uppercase variants.

# How It Works

The program uses a low-level keyboard hook (`WH_KEYBOARD_LL`) to intercept keystrokes at the OS level. When you press an AltGr key combination, it suppresses the original keystroke and sends the corresponding Idu Mishmi character to the active application. All other keystrokes, including standard shortcuts (Ctrl+C, Alt+Tab, etc.), pass through unchanged.

# Building from Source

Requires Go 1.21 or later.

```bash
# Build the Windows executable (cross-compilation from Linux/macOS)
make build

# Or directly with Go
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-H windowsgui -s -w" -o "Idu Mishmi Keyboard.exe" .
```

# License

The Idu Mishmi Windows Keyboard is open-source software licensed under the GNU General Public License v3.0.

> Permissions of this strong copyleft license are conditioned on making available complete source code of licensed works and modifications, which include larger works using a licensed work, under the same license. Copyright and license notices must be preserved. Contributors provide an express grant of patent rights.
