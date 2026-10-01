package sigmod_test

import (
	"os"
	"testing"

	"github.com/go-srvc/mods/sigmod"
	"github.com/go-srvc/srvc"
	"github.com/stretchr/testify/require"
)

func TestStop(t *testing.T) {
	l := sigmod.New(os.Interrupt)
	require.NoError(t, l.Init())

	wg := &srvc.ErrGroup{}
	wg.Go(l.Run)
	require.NoError(t, l.Stop())
	require.NoError(t, wg.Wait())
}
