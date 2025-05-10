package sigmod_test

import (
	"testing"

	"github.com/go-srvc/mods/sigmod"
	"github.com/heppu/errgroup"
	"github.com/stretchr/testify/require"
)

func TestStop(t *testing.T) {
	l := sigmod.New()
	require.NoError(t, l.Init())

	wg := &errgroup.ErrGroup{}
	wg.Go(l.Run)
	require.NoError(t, l.Stop())
	require.NoError(t, wg.Wait())
}
