//go:build !darwin && !linux

package daemon

func platformNewInhibitor() Inhibitor { return noopInhibitor{} }
