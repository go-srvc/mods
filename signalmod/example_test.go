package signalmod_test

import (
	"os"
	"syscall"
	"time"

	"github.com/go-srvc/mods/signalmod"
	"github.com/go-srvc/srvc"
)

func ExampleNew() {
	go func() {
		// Send SIGINT after 1 second.
		time.Sleep(time.Second)
		syscall.Kill(syscall.Getpid(), syscall.SIGINT) //nolint: errcheck
	}()

	srvc.RunAndExit(
		signalmod.New(os.Interrupt),
	)
}
