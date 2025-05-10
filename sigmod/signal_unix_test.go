//go:build !windows

package sigmod_test

import (
	"os"
	"syscall"
	"testing"

	"github.com/go-srvc/mods/sigmod"
	"github.com/stretchr/testify/require"
)

func TestListenerDefaultSignal(t *testing.T) {
	for _, sig := range sigmod.DefaultSignals {
		t.Run(sig.String(), func(t *testing.T) {
			var ok bool
			var s syscall.Signal
			if s, ok = sig.(syscall.Signal); !ok {
				t.Errorf("Couldn't convert signal %s to syscall", sig.String())
				return
			}

			l := sigmod.New()
			require.NoError(t, l.Init())
			require.NoError(t, syscall.Kill(syscall.Getpid(), s))
			require.NoError(t, l.Stop())
			require.Equal(t, "sigmod", l.ID())
		})
	}
}

func TestListenerOpts(t *testing.T) {
	l := sigmod.New(os.Interrupt)
	require.NoError(t, l.Init())
	require.NoError(t, syscall.Kill(syscall.Getpid(), syscall.SIGINT))
	require.NoError(t, l.Run())
	require.NoError(t, l.Stop())
}
