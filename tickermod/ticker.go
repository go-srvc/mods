// Package tickermod provides ticker functionality as a module.
package tickermod

import (
	"context"
	"errors"
	"fmt"
	"time"
)

const ID = "tickermod"

const (
	ErrMissingInterval = errStr("interval not set")
	ErrInvalidInterval = errStr("interval must be positive")
	ErrMissingTickFunc = errStr("tick function not set")
)

type errStr string

func (e errStr) Error() string { return string(e) }

type Ticker struct {
	t           *time.Ticker
	interval    time.Duration
	fn          func(context.Context) error
	opts        []Opt
	fireOnStart bool
	tickTimeout time.Duration

	ctx    context.Context
	cancel context.CancelFunc
}

// New creates ticker with given options.
// WithInterval and WithFunc options are mandatory.
func New(opts ...Opt) *Ticker {
	return &Ticker{opts: opts}
}

func (t *Ticker) Init() error {
	t.ctx, t.cancel = context.WithCancel(context.Background())
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
	if t.fireOnStart {
		if err := t.tick(t.ctx); err != nil {
			return err
		}
		// Sync next tick to fire one interval after the immediate one,
		// not interval after Init.
		t.t.Reset(t.interval)
	}
	for {
		select {
		case <-t.t.C:
			if err := t.tick(t.ctx); err != nil {
				return err
			}
		case <-t.ctx.Done():
			return nil
		}
	}
}

func (t *Ticker) tick(ctx context.Context) error {
	if t.tickTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, t.tickTimeout)
		defer cancel()
	}
	err := t.fn(ctx)
	if errors.Is(err, context.Canceled) {
		return nil
	}
	return err
}

func (t *Ticker) Stop() error {
	t.cancel()
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

// WithIntervalFn sets ticker interval using provided function. The
// returned duration must be greater than zero.
func WithIntervalFn(fn func() (time.Duration, error)) Opt {
	return func(t *Ticker) error {
		d, err := fn()
		if err != nil {
			return err
		}
		if d <= 0 {
			return ErrInvalidInterval
		}
		if t.t != nil {
			t.t.Stop()
		}
		t.interval = d
		t.t = time.NewTicker(d)
		return nil
	}
}

// WithFunc sets the function to be called on each tick.
func WithFunc(fn func() error) Opt {
	return WithFuncCtx(func(_ context.Context) error { return fn() })
}

// WithFuncCtx sets a context-aware tick function. The ctx is bound to
// the ticker's lifecycle and cancelled by Stop.
func WithFuncCtx(fn func(context.Context) error) Opt {
	return func(t *Ticker) error {
		t.fn = fn
		return nil
	}
}

// WithFireOnStart triggers the tick fn immediately on Run, before the
// first interval elapses.
func WithFireOnStart() Opt {
	return func(t *Ticker) error {
		t.fireOnStart = true
		return nil
	}
}

// WithTickTimeout bounds each individual tick by deriving a per-tick
// ctx with the supplied timeout. The tick fn must observe the ctx
// (registered via WithFuncCtx) for the timeout to take effect.
func WithTickTimeout(d time.Duration) Opt {
	return func(t *Ticker) error {
		t.tickTimeout = d
		return nil
	}
}
