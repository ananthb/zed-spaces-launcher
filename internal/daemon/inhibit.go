package daemon

// Inhibitor prevents the system from sleeping (and optionally shutting down)
// while at least one codespace session is active. Implementations are
// platform-specific; see inhibit_darwin.go and inhibit_linux.go.
type Inhibitor interface {
	// Engage asks the OS to hold a sleep inhibit assertion. Mode is one of
	// "sleep" or "sleep+shutdown". Safe to call when already engaged (no-op).
	Engage(mode string) error
	// Release drops any held assertion. Safe to call when not engaged.
	Release() error
}

type noopInhibitor struct{}

func (noopInhibitor) Engage(string) error { return nil }
func (noopInhibitor) Release() error      { return nil }

// newInhibitor picks a platform-appropriate inhibitor for the given mode.
// Returns a no-op when mode == "off" or the platform has no implementation.
func newInhibitor(mode string) Inhibitor {
	if mode == "" || mode == "off" {
		return noopInhibitor{}
	}
	return platformNewInhibitor()
}
