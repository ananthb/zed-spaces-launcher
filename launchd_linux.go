package main

// ensureLaunchdAgent is a no-op on Linux (no launchd).
func ensureLaunchdAgent() bool { return false }
