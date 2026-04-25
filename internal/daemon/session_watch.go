//go:build darwin || linux

package daemon

import (
	"syscall"
	"time"
)

// fallbackWaitPid polls kill(0) until the pid no longer exists. Used when
// the kernel-event fast path is unavailable (errors, unsupported OS, older
// kernels). Low-overhead: 2 s polling interval is fine for an inhibitor
// release signal.
func fallbackWaitPid(pid int) {
	for {
		if err := syscall.Kill(pid, 0); err != nil {
			return
		}
		time.Sleep(2 * time.Second)
	}
}
