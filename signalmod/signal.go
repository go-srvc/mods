// Package signalmod provides signal listening as a module.
package signalmod

import (
	"os"
	"os/signal"
)

const ID = "signalmod"

type Listener struct {
	ch   chan os.Signal
	sigs []os.Signal
}

// New creates signal listener for given signals.
func New(signals ...os.Signal) *Listener {
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
