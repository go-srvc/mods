// Package tickermod provides ticker functionality as a module.
package tickermod

import (
	"fmt"
	"time"
)

const ID = "tickermod"

const (
	ErrMissingInterval = errStr("interval not set")
	ErrMissingTickFunc = errStr("tick function not set")
)

type errStr string

func (e errStr) Error() string { return string(e) }

type Ticker struct {
	t       *time.Ticker
	stopped chan struct{}
	fn      func() error
	opts    []Opt
}

// New creates ticker with given options.
// WithInterval and WithFunc options are mandatory.
func New(opts ...Opt) *Ticker {
	return &Ticker{opts: opts}
}

func (t *Ticker) Init() error {
	t.stopped = make(chan struct{})
	for _, opt := range t.opts {
		if err := opt(t); err != nil {
			return fmt.Errorf("failed to apply option: %w", err)
		}
	}

	switch {
	case t.t == nil:
		return ErrMissingInterval
	case t.fn == nil:
		return ErrMissingTickFunc
	}

	return nil
}

func (t *Ticker) Run() error {
	for {
		select {
		case <-t.t.C:
			if err := t.fn(); err != nil {
				return err
			}
		case <-t.stopped:
			return nil
		}
	}
}

func (t *Ticker) Stop() error {
	defer close(t.stopped)
	t.t.Stop()
	return nil
}

func (t *Ticker) ID() string { return ID }

type Opt func(*Ticker) error

// WithInterval sets ticker interval.
func WithInterval(d time.Duration) Opt {
	return WithIntervalFn(func() (time.Duration, error) {
		return d, nil
	})
}

// WithIntervalFn sets ticker interval using provided function.
func WithIntervalFn(fn func() (time.Duration, error)) Opt {
	return func(t *Ticker) error {
		d, err := fn()
		if err != nil {
			return err
		}
		t.t = time.NewTicker(d)
		return nil
	}
}

// WithFunc sets function to be called on each tick.
// If non nil error is returned, ticker stops.
func WithFunc(fn func() error) Opt {
	return func(t *Ticker) error {
		t.fn = fn
		return nil
	}
}
