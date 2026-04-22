package main

import (
	"fmt"
	"os"
	"os/exec"
)

// ensureLaunchdAgent kicks the launchd agent if it exists. Returns true
// if the agent was started (caller should exit), false if there's no
// agent registered and the caller should run the applet inline.
func ensureLaunchdAgent() bool {
	const label = "com.cosmonaut.daemon"
	uid := fmt.Sprintf("gui/%d", os.Getuid())

	// Check if the agent is registered with launchd.
	if err := exec.Command("launchctl", "print", uid+"/"+label).Run(); err != nil {
		return false // not registered — fall back to inline
	}

	// Kick it (starts if not running, no-op if already running).
	_ = exec.Command("launchctl", "kickstart", uid+"/"+label).Run()
	return true
}
