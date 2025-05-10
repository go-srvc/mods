//go:build !windows

package sigmod

import (
	"os"
	"syscall"
)

var DefaultSignals = []os.Signal{
	os.Interrupt,
	syscall.SIGTERM,
}
