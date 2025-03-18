package sigmod_test

import (
	"os"
	"syscall"
	"testing"

	"github.com/go-srvc/mods/sigmod"
	"github.com/heppu/errgroup"
	"github.com/stretchr/testify/require"
)

func TestListener(t *testing.T) {
	tests := []struct {
		name    string
		signals []os.Signal
	}{
		{
			name:    "Defaults",
			signals: []os.Signal{},
		},
		{
			name:    "Interrupt",
			signals: []os.Signal{os.Interrupt},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := sigmod.New(tt.signals...)
			require.NoError(t, l.Init())
			require.NoError(t, syscall.Kill(syscall.Getpid(), syscall.SIGINT))
			require.NoError(t, l.Run())
			require.NoError(t, l.Stop())
			require.Equal(t, "sigmod", l.ID())
		})
	}
}

func TestStop(t *testing.T) {
	l := sigmod.New(os.Interrupt)
	require.NoError(t, l.Init())

	wg := &errgroup.ErrGroup{}
	wg.Go(l.Run)
	require.NoError(t, l.Stop())
	require.NoError(t, wg.Wait())
}
