package daemon

import (
	"fmt"
	"log"
	"runtime"
	"strings"

	"golang.design/x/hotkey"

	"github.com/ananth/codespace-zed/internal/history"
)

// defaultHotkey returns the platform default hotkey string.
func defaultHotkey() string {
	if runtime.GOOS == "darwin" {
		return "Cmd+Shift+C"
	}
	return "Super+Shift+C"
}

func (d *Daemon) startHotkeyListener() {
	hotkeyStr := defaultHotkey()
	if d.Cfg != nil && d.Cfg.Daemon != nil && d.Cfg.Daemon.Hotkey != "" {
		hotkeyStr = d.Cfg.Daemon.Hotkey
	}

	mods, key, err := parseHotkeyString(hotkeyStr)
	if err != nil {
		log.Printf("hotkey: invalid config %q: %v", hotkeyStr, err)
		return
	}

	hk := hotkey.New(mods, key)
	if err := hk.Register(); err != nil {
		log.Printf("hotkey: failed to register %q: %v", hotkeyStr, err)
		return
	}
	defer hk.Unregister()

	log.Printf("hotkey: registered %s", hotkeyStr)

	for {
		select {
		case <-hk.Keydown():
			go d.hotkeyAction()
		case <-d.stopCh:
			return
		}
	}
}

// hotkeyAction determines what to do when the hotkey is pressed,
// based on the daemon.hotkeyAction config: "picker" (default),
// "previous" (most recent repo from history), or "default" (default target).
func (d *Daemon) hotkeyAction() {
	action := "picker"
	if d.Cfg != nil && d.Cfg.Daemon != nil && d.Cfg.Daemon.HotkeyAction != "" {
		action = d.Cfg.Daemon.HotkeyAction
	}

	switch action {
	case "previous":
		hist := history.Load()
		if len(hist.Entries) > 0 {
			repo := hist.Entries[len(hist.Entries)-1].Repository
			if name := d.targetNameForRepo(repo); name != "" {
				d.showPopover(name)
				return
			}
		}
		// No history — fall through to picker.
		d.showPopover()

	case "default":
		if d.Cfg != nil && d.Cfg.DefaultTarget != "" {
			if _, ok := d.Cfg.Targets[d.Cfg.DefaultTarget]; ok {
				d.showPopover(d.Cfg.DefaultTarget)
				return
			}
		}
		// No default target — fall through to picker.
		d.showPopover()

	default: // "picker"
		d.showPopover()
	}
}

// parseHotkeyString converts a string like "Cmd+Shift+C" to hotkey modifiers and key.
func parseHotkeyString(s string) ([]hotkey.Modifier, hotkey.Key, error) {
	parts := strings.Split(s, "+")
	if len(parts) < 2 {
		return nil, 0, fmt.Errorf("expected modifier+key (e.g. Cmd+Shift+C)")
	}

	keyPart := strings.TrimSpace(parts[len(parts)-1])
	modParts := parts[:len(parts)-1]

	var mods []hotkey.Modifier
	for _, m := range modParts {
		mod, err := parseModifier(strings.TrimSpace(m))
		if err != nil {
			return nil, 0, err
		}
		mods = append(mods, mod)
	}

	key, err := parseKey(keyPart)
	if err != nil {
		return nil, 0, err
	}

	return mods, key, nil
}

func parseModifier(s string) (hotkey.Modifier, error) {
	switch strings.ToLower(s) {
	case "cmd", "command", "super", "mod":
		return hotkey.ModCmd, nil
	case "ctrl", "control":
		return hotkey.ModCtrl, nil
	case "shift":
		return hotkey.ModShift, nil
	case "alt", "option", "opt":
		return hotkey.ModOption, nil
	default:
		return 0, fmt.Errorf("unknown modifier %q", s)
	}
}

// keyMap maps lowercase key names to hotkey.Key constants.
// These constants are platform-specific virtual key codes, not ASCII.
var keyMap = map[string]hotkey.Key{
	"a": hotkey.KeyA, "b": hotkey.KeyB, "c": hotkey.KeyC, "d": hotkey.KeyD,
	"e": hotkey.KeyE, "f": hotkey.KeyF, "g": hotkey.KeyG, "h": hotkey.KeyH,
	"i": hotkey.KeyI, "j": hotkey.KeyJ, "k": hotkey.KeyK, "l": hotkey.KeyL,
	"m": hotkey.KeyM, "n": hotkey.KeyN, "o": hotkey.KeyO, "p": hotkey.KeyP,
	"q": hotkey.KeyQ, "r": hotkey.KeyR, "s": hotkey.KeyS, "t": hotkey.KeyT,
	"u": hotkey.KeyU, "v": hotkey.KeyV, "w": hotkey.KeyW, "x": hotkey.KeyX,
	"y": hotkey.KeyY, "z": hotkey.KeyZ,
	"0": hotkey.Key0, "1": hotkey.Key1, "2": hotkey.Key2, "3": hotkey.Key3,
	"4": hotkey.Key4, "5": hotkey.Key5, "6": hotkey.Key6, "7": hotkey.Key7,
	"8": hotkey.Key8, "9": hotkey.Key9,
	"space":     hotkey.KeySpace,
	"return":    hotkey.KeyReturn, "enter": hotkey.KeyReturn,
	"escape":    hotkey.KeyEscape, "esc": hotkey.KeyEscape,
	"tab":       hotkey.KeyTab,
	"delete":    hotkey.KeyDelete, "backspace": hotkey.KeyDelete,
	"up":        hotkey.KeyUp, "down": hotkey.KeyDown,
	"left":      hotkey.KeyLeft, "right": hotkey.KeyRight,
	"f1":  hotkey.KeyF1, "f2": hotkey.KeyF2, "f3": hotkey.KeyF3,
	"f4":  hotkey.KeyF4, "f5": hotkey.KeyF5, "f6": hotkey.KeyF6,
	"f7":  hotkey.KeyF7, "f8": hotkey.KeyF8, "f9": hotkey.KeyF9,
	"f10": hotkey.KeyF10, "f11": hotkey.KeyF11, "f12": hotkey.KeyF12,
}

func parseKey(s string) (hotkey.Key, error) {
	if k, ok := keyMap[strings.ToLower(s)]; ok {
		return k, nil
	}
	return 0, fmt.Errorf("unknown key %q", s)
}
