package daemon

import (
	"fmt"
	"strings"

	"golang.design/x/hotkey"
)

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
