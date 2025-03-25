// Package sigmod provides signal listening as a module.
package sigmod

import (
	"os"
	"os/signal"
	"syscall"
)

const ID = "sigmod"

type Listener struct {
	ch   chan os.Signal
	sigs []os.Signal
}

// New creates signal listener for given signals.
// If no signal is provided, os.Interrupt and SIGTERM will be used.
func New(signals ...os.Signal) *Listener {
	if len(signals) == 0 {
		signals = []os.Signal{
			os.Interrupt,
			syscall.SIGTERM, // XXX Not available on Windows
		}
	}
	return &Listener{
		sigs: signals,
	}
}

func (l *Listener) Init() error {
	l.ch = make(chan os.Signal, 1)
	signal.Notify(l.ch, l.sigs...)
	return nil
}

func (l *Listener) Run() error {
	<-l.ch
	return nil
}

func (l *Listener) Stop() error {
	defer close(l.ch)
	signal.Stop(l.ch)
	return nil
}

func (l *Listener) ID() string { return ID }
