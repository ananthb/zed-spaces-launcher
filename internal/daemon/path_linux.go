package daemon

// enrichPath is a no-op on Linux: systemd user services inherit
// the graphical session's environment which typically includes the
// user's full PATH.
func enrichPath() {}
