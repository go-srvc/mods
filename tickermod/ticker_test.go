package tickermod_test

import (
	"errors"
	"testing"
	"time"

	"github.com/go-srvc/mods/tickermod"
	"github.com/heppu/errgroup"
	"github.com/stretchr/testify/require"
)

func TestListener(t *testing.T) {
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
}

func TestListenerRunError(t *testing.T) {
	errRun := errors.New("run error")
	tickerMod := tickermod.New(
		tickermod.WithInterval(time.Millisecond*10),
		tickermod.WithFunc(func() error { return errRun }),
	)

	require.NoError(t, tickerMod.Init())
	require.ErrorIs(t, tickerMod.Run(), errRun)
	require.NoError(t, tickerMod.Stop())
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
			name:        "ErrMissingWithFunc",
			ticker:      tickermod.New(tickermod.WithInterval(time.Second)),
			expectedErr: tickermod.ErrMissingWithFunc,
		},
		{
			name:        "ErrMissingWithInterval",
			ticker:      tickermod.New(tickermod.WithFunc(func() error { return nil })),
			expectedErr: tickermod.ErrMissingWithInterval,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.ticker.Init()
			require.ErrorIs(t, err, tc.expectedErr)
		})
	}
}
