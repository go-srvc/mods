package sigmod

import (
	"os"
	"syscall"
)

var defaultSignals = []os.Signal{
	os.Interrupt,
	syscall.SIGTERM,
}
