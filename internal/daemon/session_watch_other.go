//go:build !darwin && !linux

package daemon

// waitPidExit on unsupported OSes returns immediately. The tracker will
// see the pid as "exited" the moment it registers, so the inhibitor never
// stays engaged — acceptable because the inhibitor is a no-op on those
// platforms anyway.
func waitPidExit(pid int) {}
