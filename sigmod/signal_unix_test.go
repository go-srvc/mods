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
	// Decoupled from sigmod.defaultSignals to detect breaking changes.
	testSignals := []syscall.Signal{syscall.SIGINT}

	for _, sig := range testSignals {
		t.Run(sig.String(), func(t *testing.T) {
			l := sigmod.New()
			require.NoError(t, l.Init())
			require.NoError(t, syscall.Kill(syscall.Getpid(), sig))
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
