package daemon

import (
	"fmt"
	"strings"

	"golang.design/x/hotkey"
)

func parseModifier(s string) (hotkey.Modifier, error) {
	switch strings.ToLower(s) {
	case "super", "mod", "cmd", "command":
		return hotkey.Mod4, nil // Super/Meta key
	case "ctrl", "control":
		return hotkey.ModCtrl, nil
	case "shift":
		return hotkey.ModShift, nil
	case "alt", "option", "opt":
		return hotkey.Mod1, nil // Alt key
	default:
		return 0, fmt.Errorf("unknown modifier %q", s)
	}
}
