// Package stopmod external trigger to exit module.
// This is mostly convinient for testing purposes.
package stopmod

import (
	"sync"
)

const ID = "stopmod"

type Stopper struct {
	done chan struct{}
	stop sync.Once
}

func New() *Stopper {
	return &Stopper{}
}

func (s *Stopper) Init() error {
	s.done = make(chan struct{})
	return nil
}

func (s *Stopper) Run() error {
	<-s.done
	return nil
}

func (s *Stopper) Stop() error {
	s.stop.Do(func() {
		close(s.done)
	})
	return nil
}

func (s *Stopper) ID() string { return ID }
