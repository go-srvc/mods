package sigmod_test

import (
	"os"
	"runtime"
	"syscall"
	"testing"

	"github.com/go-srvc/mods/sigmod"
	"github.com/heppu/errgroup"
	"github.com/stretchr/testify/require"
)

func TestListener(t *testing.T) {
	tests := []struct {
		name     string
		signals  []os.Signal
		inputSig syscall.Signal
	}{
		{
			name:     "Defaults",
			signals:  []os.Signal{},
			inputSig: syscall.SIGINT,
		},
		{
			name:     "Interrupt",
			signals:  []os.Signal{os.Interrupt},
			inputSig: syscall.SIGINT,
		},
		{
			name:     "SIGTERM",
			signals:  []os.Signal{syscall.SIGTERM},
			inputSig: syscall.SIGTERM,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := sigmod.New(tt.signals...)
			require.NoError(t, l.Init())
			if tt.inputSig == syscall.SIGTERM && runtime.GOOS == "windows" {
				t.Skip("Windows doesn't support SIGTERM")
			}
			require.NoError(t, syscall.Kill(syscall.Getpid(), tt.inputSig))
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
