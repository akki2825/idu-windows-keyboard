// Idu Mishmi Keyboard – Linux test tool.
//
// Grabs the physical keyboard via evdev, creates a virtual keyboard via uinput
// for passthrough keys, and injects remapped Unicode characters via clipboard.
// Open any editor, run this tool, and type to see the Idu Mishmi layout in action.
//
// Usage: sudo -E ./testkeys [/dev/input/eventN]
// Press ESC to release keyboard and exit.
package main

import (
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

// ── evdev types ──

const inputEventSize = 24

const (
	evSYN = 0x00
	evKEY = 0x01

	valRelease = 0
	valPress   = 1
	valRepeat  = 2
)

// ── ioctl constants ──

const (
	eviocGrab    = 0x40044590
	uiSetEvbit   = 0x40045564
	uiSetKeybit  = 0x40045565
	uiDevCreate  = 0x5501
	uiDevDestroy = 0x5502
)

// ── Linux key codes ──

const (
	keyESC        = 1
	key1          = 2
	key2          = 3
	key3          = 4
	key4          = 5
	key5          = 6
	key6          = 7
	key7          = 8
	key8          = 9
	key9          = 10
	key0          = 11
	keyMINUS      = 12
	keyEQUAL      = 13
	keyBACKSPACE  = 14
	keyTAB        = 15
	keyQ          = 16
	keyW          = 17
	keyE          = 18
	keyR          = 19
	keyT          = 20
	keyY          = 21
	keyU          = 22
	keyI          = 23
	keyO          = 24
	keyP          = 25
	keyLBRACE     = 26
	keyRBRACE     = 27
	keyENTER      = 28
	keyLCTRL      = 29
	keyA          = 30
	keyS          = 31
	keyD          = 32
	keyF          = 33
	keyG          = 34
	keyH          = 35
	keyJ          = 36
	keyK          = 37
	keyL          = 38
	keySEMICOLON  = 39
	keyAPOSTROPHE = 40
	keyGRAVE      = 41
	keyLSHIFT     = 42
	keyBACKSLASH  = 43
	keyZ          = 44
	keyX          = 45
	keyC          = 46
	keyV          = 47
	keyB          = 48
	keyN          = 49
	keyM          = 50
	keyCOMMA      = 51
	keyDOT        = 52
	keySLASH      = 53
	keyRSHIFT     = 54
	keyLALT       = 56
	keySPACE      = 57
	keyCAPSLOCK   = 58
	keySCROLLLOCK = 70
	keyRCTRL      = 97
	keyRALT       = 100
	keyHOME       = 102
	keyUP         = 103
	keyPAGEUP     = 104
	keyLEFT       = 105
	keyRIGHT      = 106
	keyEND        = 107
	keyDOWN       = 108
	keyPAGEDOWN   = 109
	keyINSERT     = 110
	keyDELETE     = 111
	keyLMETA      = 125
	keyRMETA      = 126
	keyMAX        = 200
)

// ── Modifier state ──

var (
	lShift, rShift bool
	lCtrl, rCtrl   bool
	lAlt           bool
	rAlt           bool // AltGr
	capsLock       bool
	kbActive       = true
)

func shiftHeld() bool    { return lShift || rShift }
func logicalShift() bool { return shiftHeld() != capsLock }
func altGrHeld() bool    { return rAlt }

// ── Combining accent constants ──

const (
	noAccent    rune = 0
	accentGrave rune = '\u0300'
	accentAcute rune = '\u0301'
	accentTilde rune = '\u0303'
	accentMacro rune = '\u0304'
)

// ── Dead key state ──

var pendingAccent rune = noAccent

// ── Idu Mishmi keymap ──

// Direct AltGr mappings: [0]=lowercase, [1]=uppercase
var directAltGr = map[uint16][2]string{
	keyE: {"\u0259", "\u018F"},             // ə, Ə
	keyR: {"\u0259\u02DE", "\u018F\u02DE"}, // ə˞, Ə˞
}

// Dead key triggers: AltGr+key → [0]=no shift, [1]=shift
var deadKeyTriggers = map[uint16][2]rune{
	keyGRAVE:      {accentGrave, accentTilde}, // ` → grave, ~ → tilde
	keyAPOSTROPHE: {accentAcute, 0},           // ' → acute
	keyMINUS:      {accentMacro, 0},           // - → macron
}

// Precomposed accented vowels: accent → base rune → output
var precomposed = map[rune]map[rune]string{
	accentGrave: {
		'a': "\u00E0", 'A': "\u00C0",
		'e': "\u00E8", 'E': "\u00C8",
		'i': "\u00EC", 'I': "\u00CC",
		'o': "\u00F2", 'O': "\u00D2",
		'u': "\u00F9", 'U': "\u00D9",
	},
	accentAcute: {
		'a': "\u00E1", 'A': "\u00C1",
		'e': "\u00E9", 'E': "\u00C9",
		'i': "\u00ED", 'I': "\u00CD",
		'o': "\u00F3", 'O': "\u00D3",
		'u': "\u00FA", 'U': "\u00DA",
	},
	accentTilde: {
		'a': "\u00E3", 'A': "\u00C3",
		'e': "\u1EBD", 'E': "\u1EBC",
		'i': "\u0129", 'I': "\u0128",
		'o': "\u00F5", 'O': "\u00D5",
		'u': "\u0169", 'U': "\u0168",
	},
	accentMacro: {
		'a': "\u0101", 'A': "\u0100",
		'e': "\u0113", 'E': "\u0112",
		'i': "\u012B", 'I': "\u012A",
		'o': "\u014D", 'O': "\u014C",
		'u': "\u016B", 'U': "\u016A",
	},
}

// vowelKeys maps Linux key codes to [lowercase, uppercase] rune
var vowelKeys = map[uint16][2]rune{
	keyA: {'a', 'A'}, keyE: {'e', 'E'}, keyI: {'i', 'I'},
	keyO: {'o', 'O'}, keyU: {'u', 'U'},
}

// applyAccent inserts combining accent after first rune.
func applyAccent(accent rune, s string) string {
	runes := []rune(s)
	if len(runes) == 0 {
		return string(accent)
	}
	result := make([]rune, 0, len(runes)+1)
	result = append(result, runes[0])
	result = append(result, accent)
	result = append(result, runes[1:]...)
	return string(result)
}

func accentName(a rune) string {
	switch a {
	case accentGrave:
		return "grave"
	case accentAcute:
		return "acute"
	case accentTilde:
		return "tilde"
	case accentMacro:
		return "macron"
	}
	return "?"
}

// ── uinput virtual keyboard ──

type uinputUserDev struct {
	Name         [80]byte
	ID           [8]byte
	FFEffectsMax uint32
	Absmax       [64]int32
	Absmin       [64]int32
	Absfuzz      [64]int32
	Absflat      [64]int32
}

func uiIoctl(fd uintptr, req uintptr, val uintptr) error {
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd, req, val)
	if errno != 0 {
		return errno
	}
	return nil
}

func createVirtualKeyboard() (*os.File, error) {
	f, err := os.OpenFile("/dev/uinput", os.O_WRONLY|syscall.O_NONBLOCK, 0)
	if err != nil {
		return nil, fmt.Errorf("open /dev/uinput: %w (try sudo)", err)
	}
	fd := f.Fd()
	if err := uiIoctl(fd, uiSetEvbit, evKEY); err != nil {
		f.Close()
		return nil, err
	}
	if err := uiIoctl(fd, uiSetEvbit, evSYN); err != nil {
		f.Close()
		return nil, err
	}
	for i := uintptr(1); i <= keyMAX; i++ {
		uiIoctl(fd, uiSetKeybit, i)
	}
	var dev uinputUserDev
	copy(dev.Name[:], "Idu Mishmi Virtual KB")
	dev.ID[0] = 0x03
	if _, err := f.Write((*[unsafe.Sizeof(dev)]byte)(unsafe.Pointer(&dev))[:]); err != nil {
		f.Close()
		return nil, err
	}
	if err := uiIoctl(fd, uiDevCreate, 0); err != nil {
		f.Close()
		return nil, err
	}
	time.Sleep(200 * time.Millisecond)
	return f, nil
}

func destroyVirtualKeyboard(f *os.File) {
	if f != nil {
		uiIoctl(f.Fd(), uiDevDestroy, 0)
		f.Close()
	}
}

func emitKey(f *os.File, code uint16, value int32) {
	var buf [inputEventSize]byte
	binary.LittleEndian.PutUint16(buf[16:18], evKEY)
	binary.LittleEndian.PutUint16(buf[18:20], code)
	binary.LittleEndian.PutUint32(buf[20:24], uint32(value))
	f.Write(buf[:])
	var syn [inputEventSize]byte
	binary.LittleEndian.PutUint16(syn[16:18], evSYN)
	f.Write(syn[:])
}

func releaseModifiersForInjection(vkb *os.File) func() {
	var restore []uint16
	if lShift {
		emitKey(vkb, keyLSHIFT, valRelease)
		restore = append(restore, keyLSHIFT)
	}
	if rShift {
		emitKey(vkb, keyRSHIFT, valRelease)
		restore = append(restore, keyRSHIFT)
	}
	return func() {
		for _, k := range restore {
			emitKey(vkb, k, valPress)
		}
	}
}

// ── Clipboard injection ──

func injectString(vkb *os.File, s string) {
	cmd := exec.Command("xclip", "-selection", "clipboard")
	cmd.Stdin = strings.NewReader(s)
	if err := cmd.Run(); err != nil {
		return
	}
	time.Sleep(10 * time.Millisecond)
	emitKey(vkb, keyLCTRL, valPress)
	emitKey(vkb, keyV, valPress)
	emitKey(vkb, keyV, valRelease)
	emitKey(vkb, keyLCTRL, valRelease)
}

// ── Device discovery ──

func findKeyboard() string {
	for _, pattern := range []string{
		"/dev/input/by-id/*-event-kbd",
		"/dev/input/by-path/*-event-kbd",
	} {
		matches, _ := filepath.Glob(pattern)
		for _, m := range matches {
			if resolved, err := filepath.EvalSymlinks(m); err == nil {
				return resolved
			}
		}
	}
	data, err := os.ReadFile("/proc/bus/input/devices")
	if err == nil {
		for _, block := range strings.Split(string(data), "\n\n") {
			low := strings.ToLower(block)
			if strings.Contains(low, "keyboard") && !strings.Contains(low, "consumer") {
				for _, line := range strings.Split(block, "\n") {
					if strings.HasPrefix(line, "H: Handlers=") {
						for _, tok := range strings.Fields(line) {
							if strings.HasPrefix(tok, "event") {
								return "/dev/input/" + tok
							}
						}
					}
				}
			}
		}
	}
	return ""
}

func keyName(code uint16) string {
	names := map[uint16]string{
		keyE: "e", keyR: "r", keyA: "a", keyI: "i", keyO: "o", keyU: "u",
		keyW: "w", keyM: "m", keySEMICOLON: ";", keyAPOSTROPHE: "'",
		keyGRAVE: "`", keyMINUS: "-",
		keyB: "b", keyC: "c", keyD: "d", keyF: "f", keyG: "g", keyH: "h",
		keyJ: "j", keyK: "k", keyL: "l", keyN: "n", keyP: "p", keyQ: "q",
		keyS: "s", keyT: "t", keyV: "v", keyX: "x", keyY: "y", keyZ: "z",
		keySPACE: "Space", keyCOMMA: ",", keyDOT: ".", keySLASH: "/",
		keyEQUAL: "=", keyLBRACE: "[", keyRBRACE: "]", keyBACKSLASH: "\\",
		key1: "1", key2: "2", key3: "3", key4: "4", key5: "5",
		key6: "6", key7: "7", key8: "8", key9: "9", key0: "0",
	}
	if n, ok := names[code]; ok {
		return n
	}
	return fmt.Sprintf("key_%d", code)
}

// ── Main ──

func main() {
	if _, err := exec.LookPath("xclip"); err != nil {
		fmt.Fprintln(os.Stderr, "xclip not found. Install: sudo apt install xclip")
		os.Exit(1)
	}

	devPath := ""
	if len(os.Args) > 1 {
		devPath = os.Args[1]
	} else {
		devPath = findKeyboard()
	}
	if devPath == "" {
		fmt.Fprintln(os.Stderr, "Could not find keyboard. Specify: sudo -E ./testkeys /dev/input/eventN")
		os.Exit(1)
	}

	evdev, err := os.Open(devPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot open %s: %v\nTry: sudo -E ./testkeys\n", devPath, err)
		os.Exit(1)
	}
	defer evdev.Close()
	fmt.Printf("Keyboard: %s\n", devPath)

	vkb, err := createVirtualKeyboard()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot create virtual keyboard: %v\n", err)
		os.Exit(1)
	}
	defer destroyVirtualKeyboard(vkb)
	fmt.Println("Virtual keyboard created")

	one := int32(1)
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, evdev.Fd(), eviocGrab, uintptr(unsafe.Pointer(&one)))
	if errno != 0 {
		fmt.Fprintf(os.Stderr, "Cannot grab keyboard: %v\n", errno)
		destroyVirtualKeyboard(vkb)
		os.Exit(1)
	}
	release := func() {
		zero := int32(0)
		syscall.Syscall(syscall.SYS_IOCTL, evdev.Fd(), eviocGrab, uintptr(unsafe.Pointer(&zero)))
	}
	defer release()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		release()
		destroyVirtualKeyboard(vkb)
		fmt.Println("\nReleased keyboard. Bye!")
		os.Exit(0)
	}()

	fmt.Println("")
	fmt.Println("╔════════════════════════════════════════════════════════╗")
	fmt.Println("║         Idu Mishmi Keyboard — Linux Test Mode          ║")
	fmt.Println("╠════════════════════════════════════════════════════════╣")
	fmt.Println("║  Direct:     AltGr+e → ə       AltGr+Shift+e → Ə     ║")
	fmt.Println("║              AltGr+r → ə˞      AltGr+Shift+r → Ə˞    ║")
	fmt.Println("║                                                        ║")
	fmt.Println("║  Dead keys:  AltGr+` then vowel → grave accent        ║")
	fmt.Println("║              AltGr+~ then vowel → tilde/nasalized      ║")
	fmt.Println("║              AltGr+' then vowel → acute accent         ║")
	fmt.Println("║              AltGr+- then vowel → macron (long)        ║")
	fmt.Println("║                                                        ║")
	fmt.Println("║  Examples:   AltGr+` then a → à                       ║")
	fmt.Println("║              AltGr+~ then AltGr+e → ə̃                 ║")
	fmt.Println("║              AltGr+` then AltGr+r → ə̀˞               ║")
	fmt.Println("║                                                        ║")
	fmt.Println("║  ESC = exit    ScrollLock = toggle on/off              ║")
	fmt.Println("╚════════════════════════════════════════════════════════╝")
	fmt.Println("")
	fmt.Println("  [ACTIVE] Type in your editor now...")

	buf := make([]byte, inputEventSize)
	suppressed := make(map[uint16]bool)

	for {
		n, err := evdev.Read(buf)
		if err != nil || n != inputEventSize {
			continue
		}
		evType := binary.LittleEndian.Uint16(buf[16:18])
		evCode := binary.LittleEndian.Uint16(buf[18:20])
		evValue := int32(binary.LittleEndian.Uint32(buf[20:24]))

		if evType != evKEY {
			vkb.Write(buf[:n])
			continue
		}

		pressed := evValue == valPress
		repeated := evValue == valRepeat
		released := evValue == valRelease

		// ── ESC ──
		if evCode == keyESC && pressed {
			fmt.Println("\n  ESC — releasing keyboard. Bye!")
			return
		}

		// ── ScrollLock ──
		if evCode == keySCROLLLOCK && pressed {
			kbActive = !kbActive
			pendingAccent = noAccent
			if kbActive {
				fmt.Println("  [ACTIVE]")
			} else {
				fmt.Println("  [INACTIVE]")
			}
			continue
		}

		// ── Modifiers ──
		switch evCode {
		case keyLSHIFT:
			lShift = !released
			emitKey(vkb, evCode, evValue)
			continue
		case keyRSHIFT:
			rShift = !released
			emitKey(vkb, evCode, evValue)
			continue
		case keyLCTRL:
			lCtrl = !released
			emitKey(vkb, evCode, evValue)
			continue
		case keyRCTRL:
			rCtrl = !released
			emitKey(vkb, evCode, evValue)
			continue
		case keyLALT:
			lAlt = !released
			emitKey(vkb, evCode, evValue)
			continue
		case keyRALT:
			rAlt = !released
			continue // Don't forward AltGr
		case keyCAPSLOCK:
			if pressed {
				capsLock = !capsLock
			}
			emitKey(vkb, evCode, evValue)
			continue
		}

		// ── Key release ──
		if released {
			if suppressed[evCode] {
				delete(suppressed, evCode)
				continue
			}
			emitKey(vkb, evCode, evValue)
			continue
		}

		// ── Inactive: forward everything ──
		if !kbActive {
			emitKey(vkb, evCode, evValue)
			continue
		}

		// ── Ctrl/Alt/Win combos: pass through ──
		if lCtrl || rCtrl || lAlt {
			pendingAccent = noAccent
			emitKey(vkb, evCode, evValue)
			continue
		}

		// Only process key press / repeat
		if !pressed && !repeated {
			emitKey(vkb, evCode, evValue)
			continue
		}

		ag := altGrHeld()
		sh := logicalShift()

		// ── Dead key triggers (AltGr + `, ~, ', -) ──
		if ag {
			if accents, ok := deadKeyTriggers[evCode]; ok {
				accent := accents[0]
				if shiftHeld() && accents[1] != 0 {
					accent = accents[1]
				}
				pendingAccent = accent
				suppressed[evCode] = true
				fmt.Printf("  [dead key: %s] waiting for vowel...\n", accentName(accent))
				continue
			}
		}

		// ── Resolve pending dead key ──
		if pendingAccent != noAccent {
			accent := pendingAccent
			pendingAccent = noAccent

			// Try AltGr+E (ə) or AltGr+R (ə˞)
			if ag {
				if m, ok := directAltGr[evCode]; ok {
					base := m[0]
					if sh {
						base = m[1]
					}
					output := applyAccent(accent, base)
					suppressed[evCode] = true
					restore := releaseModifiersForInjection(vkb)
					injectString(vkb, output)
					restore()
					fmt.Printf("  %s + AltGr+%s → %s\n", accentName(accent), keyName(evCode), output)
					continue
				}
			}

			// Try regular vowel (a, e, i, o, u)
			if !ag {
				if vowels, ok := vowelKeys[evCode]; ok {
					v := vowels[0]
					if sh {
						v = vowels[1]
					}
					output := ""
					if accentMap, ok := precomposed[accent]; ok {
						if pre, ok := accentMap[v]; ok {
							output = pre
						}
					}
					if output == "" {
						output = string(v) + string(accent)
					}
					suppressed[evCode] = true
					restore := releaseModifiersForInjection(vkb)
					injectString(vkb, output)
					restore()
					fmt.Printf("  %s + %s → %s\n", accentName(accent), keyName(evCode), output)
					continue
				}
			}

			// Dead key not resolved — cancel and process key normally
			fmt.Printf("  [dead key cancelled]\n")
		}

		// ── Direct AltGr mappings (ə, ə˞) ──
		if ag {
			if m, ok := directAltGr[evCode]; ok {
				output := m[0]
				if sh {
					output = m[1]
				}
				suppressed[evCode] = true
				restore := releaseModifiersForInjection(vkb)
				injectString(vkb, output)
				restore()
				combo := "AltGr+"
				if sh {
					combo = "AltGr+Shift+"
				}
				fmt.Printf("  %s%s → %s\n", combo, keyName(evCode), output)
				continue
			}
		}

		// ── Passthrough ──
		emitKey(vkb, evCode, evValue)
	}
}
