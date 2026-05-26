package tickermod_test

import (
	"context"
	"errors"
	"testing"
	"testing/synctest"
	"time"

	"github.com/go-srvc/mods/tickermod"
	"github.com/heppu/errgroup"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestListener(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		called := make(chan struct{})
		tickerMod := tickermod.New(
			tickermod.WithInterval(time.Millisecond*10),
			tickermod.WithFunc(func() error {
				select {
				case called <- struct{}{}:
				default:
				}
				return nil
			}),
		)

		require.NoError(t, tickerMod.Init())
		wg := &errgroup.ErrGroup{}
		wg.Go(tickerMod.Run)
		<-called
		require.NoError(t, tickerMod.Stop())
		require.NoError(t, wg.Wait())
		require.Equal(t, "tickermod", tickerMod.ID())
	})
}

func TestListenerRunError(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		errRun := errors.New("run error")
		tickerMod := tickermod.New(
			tickermod.WithInterval(time.Millisecond*10),
			tickermod.WithFunc(func() error { return errRun }),
		)

		require.NoError(t, tickerMod.Init())
		require.ErrorIs(t, tickerMod.Run(), errRun)
		require.NoError(t, tickerMod.Stop())
	})
}

func TestListenerInitErrors(t *testing.T) {
	errOpt := errors.New("opt error")

	tests := []struct {
		name        string
		ticker      *tickermod.Ticker
		expectedErr error
	}{
		{
			name:        "ErrOpt",
			ticker:      tickermod.New(func(t *tickermod.Ticker) error { return errOpt }),
			expectedErr: errOpt,
		},
		{
			name:        "ErrMissingTickFunc",
			ticker:      tickermod.New(tickermod.WithInterval(time.Second)),
			expectedErr: tickermod.ErrMissingTickFunc,
		},
		{
			name:        "ErrMissingInterval",
			ticker:      tickermod.New(tickermod.WithFunc(func() error { return nil })),
			expectedErr: tickermod.ErrMissingInterval,
		},
		{
			name: "ErrInvalidInterval/zero",
			ticker: tickermod.New(
				tickermod.WithInterval(0),
				tickermod.WithFunc(func() error { return nil }),
			),
			expectedErr: tickermod.ErrInvalidInterval,
		},
		{
			name: "ErrInvalidInterval/negative",
			ticker: tickermod.New(
				tickermod.WithInterval(-time.Second),
				tickermod.WithFunc(func() error { return nil }),
			),
			expectedErr: tickermod.ErrInvalidInterval,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.ticker.Init()
			require.ErrorIs(t, err, tc.expectedErr)
		})
	}
}

func TestFuncCtxReceivesLifecycleContext(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		captured := make(chan context.Context, 1)
		tickerMod := tickermod.New(
			tickermod.WithInterval(time.Millisecond*10),
			tickermod.WithFuncCtx(func(ctx context.Context) error {
				select {
				case captured <- ctx:
				default:
				}
				return nil
			}),
		)

		require.NoError(t, tickerMod.Init())
		wg := &errgroup.ErrGroup{}
		wg.Go(tickerMod.Run)

		ctx := <-captured
		require.NoError(t, ctx.Err())
		require.NoError(t, tickerMod.Stop())
		require.NoError(t, wg.Wait())
		require.ErrorIs(t, ctx.Err(), context.Canceled)
	})
}

func TestShutdownDuringSlowTickReturnsPromptly(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		tickStarted := make(chan struct{})
		tickerMod := tickermod.New(
			tickermod.WithInterval(time.Millisecond*10),
			tickermod.WithFuncCtx(func(ctx context.Context) error {
				select {
				case tickStarted <- struct{}{}:
				default:
				}
				<-ctx.Done()
				return ctx.Err()
			}),
		)

		require.NoError(t, tickerMod.Init())
		wg := &errgroup.ErrGroup{}
		wg.Go(tickerMod.Run)

		<-tickStarted
		stopDone := make(chan error, 1)
		go func() { stopDone <- tickerMod.Stop() }()

		select {
		case err := <-stopDone:
			require.NoError(t, err)
		case <-time.After(time.Second):
			t.Fatal("Stop blocked: ctx cancellation not propagated to tick fn")
		}
		require.NoError(t, wg.Wait())
	})
}

func TestShutdownDoesNotSwallowUnrelatedTickError(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		errReal := errors.New("unrelated tick failure")
		tickRunning := make(chan struct{})
		unblock := make(chan struct{})
		tickerMod := tickermod.New(
			tickermod.WithInterval(time.Hour),
			tickermod.WithFireOnStart(),
			tickermod.WithFunc(func() error {
				select {
				case tickRunning <- struct{}{}:
				default:
				}
				<-unblock
				return errReal
			}),
		)

		require.NoError(t, tickerMod.Init())
		runErr := make(chan error, 1)
		go func() { runErr <- tickerMod.Run() }()

		<-tickRunning
		require.NoError(t, tickerMod.Stop())
		close(unblock)

		require.ErrorIs(t, <-runErr, errReal)
	})
}

func TestFireOnStartTicksImmediately(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		called := make(chan struct{}, 1)
		tickerMod := tickermod.New(
			tickermod.WithInterval(time.Hour),
			tickermod.WithFunc(func() error {
				select {
				case called <- struct{}{}:
				default:
				}
				return nil
			}),
			tickermod.WithFireOnStart(),
		)

		require.NoError(t, tickerMod.Init())
		wg := &errgroup.ErrGroup{}
		wg.Go(tickerMod.Run)

		select {
		case <-called:
		case <-time.After(time.Second):
			t.Fatal("WithFireOnStart did not fire the first tick before the interval")
		}

		require.NoError(t, tickerMod.Stop())
		require.NoError(t, wg.Wait())
	})
}

func TestFireOnStartPropagatesError(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		errBoom := errors.New("boom")
		tickerMod := tickermod.New(
			tickermod.WithInterval(time.Hour),
			tickermod.WithFunc(func() error { return errBoom }),
			tickermod.WithFireOnStart(),
		)
		require.NoError(t, tickerMod.Init())
		require.ErrorIs(t, tickerMod.Run(), errBoom)
		require.NoError(t, tickerMod.Stop())
	})
}

// TestFireOnStartResetsInterval verifies the second tick fires one full
// interval after the immediate one, not after Init.
func TestFireOnStartResetsInterval(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		const interval = time.Second
		ticks := make(chan time.Time, 4)
		tickerMod := tickermod.New(
			tickermod.WithInterval(interval),
			tickermod.WithFireOnStart(),
			tickermod.WithFunc(func() error {
				select {
				case ticks <- time.Now():
				default:
				}
				return nil
			}),
		)
		require.NoError(t, tickerMod.Init())
		wg := &errgroup.ErrGroup{}
		wg.Go(tickerMod.Run)

		first := <-ticks
		second := <-ticks
		require.InDelta(t, interval, second.Sub(first), float64(time.Millisecond),
			"second tick should be one interval after the immediate one")

		require.NoError(t, tickerMod.Stop())
		require.NoError(t, wg.Wait())
	})
}

func TestTickTimeoutBoundsTickFn(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		hit := make(chan error, 1)
		tickerMod := tickermod.New(
			tickermod.WithInterval(time.Hour),
			tickermod.WithFireOnStart(),
			tickermod.WithFuncCtx(func(ctx context.Context) error {
				<-ctx.Done()
				err := ctx.Err()
				select {
				case hit <- err:
				default:
				}
				return nil
			}),
			tickermod.WithTickTimeout(20*time.Millisecond),
		)

		require.NoError(t, tickerMod.Init())
		wg := &errgroup.ErrGroup{}
		wg.Go(tickerMod.Run)

		select {
		case err := <-hit:
			require.ErrorIs(t, err, context.DeadlineExceeded)
		case <-time.After(time.Second):
			t.Fatal("per-tick timeout did not fire")
		}

		require.NoError(t, tickerMod.Stop())
		require.NoError(t, wg.Wait())
	})
}

func TestTickTimeoutNoOpWithoutFuncCtx(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		completed := make(chan struct{}, 1)
		tickerMod := tickermod.New(
			tickermod.WithInterval(time.Hour),
			tickermod.WithFireOnStart(),
			tickermod.WithFunc(func() error {
				// Non-ctx fn can't observe cancellation; runs to completion
				// despite the aggressive per-tick timeout below.
				time.Sleep(time.Hour)
				select {
				case completed <- struct{}{}:
				default:
				}
				return nil
			}),
			tickermod.WithTickTimeout(time.Millisecond),
		)

		require.NoError(t, tickerMod.Init())
		wg := &errgroup.ErrGroup{}
		wg.Go(tickerMod.Run)

		<-completed
		require.NoError(t, tickerMod.Stop())
		require.NoError(t, wg.Wait())
	})
}
