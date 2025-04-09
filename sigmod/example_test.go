package sigmod_test

import (
	"os"
	"syscall"
	"time"

	"github.com/go-srvc/mods/sigmod"
	"github.com/go-srvc/srvc"
)

func ExampleNew() {
	//nolint: errcheck
	go func() {
		// Send SIGINT after 1 second.
		time.Sleep(time.Second)
		p, _ := os.FindProcess(syscall.Getegid())
		p.Signal(os.Interrupt)
	}()

	srvc.RunAndExit(
		sigmod.New(os.Interrupt),
	)
}
