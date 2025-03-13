package stopmod_test

import (
	"testing"

	"github.com/go-srvc/mods/stopmod"
	"github.com/heppu/errgroup"
	"github.com/stretchr/testify/require"
)

func TestStopper(t *testing.T) {
	l := stopmod.New()
	require.NoError(t, l.Init())
	eg := &errgroup.ErrGroup{}
	eg.Go(l.Run)
	require.NoError(t, l.Stop())
	require.NoError(t, eg.Wait())
	require.Equal(t, "stopmod", l.ID())
	// verify that second stop doesn't panic
	require.NoError(t, l.Stop())
}
