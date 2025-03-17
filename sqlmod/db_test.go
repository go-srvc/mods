package sqlmod_test

import (
	"database/sql"
	"testing"

	"github.com/go-srvc/mods/sqlmod"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDB(t *testing.T) {
	db := &sql.DB{}
	dbx := sqlmod.New(
		sqlmod.WithDBx(db),
	)
	require.NoError(t, dbx.Init())
	assert.Equal(t, db, dbx.DB())

	// TODO: Test against the real database so close works
	// wg := &errgroup.ErrGroup{}
	// wg.Go(dbx.Run)
	// require.NoError(t, dbx.Stop())
	// require.NoError(t, wg.Wait())
	require.Equal(t, "sqlmod", dbx.ID())
}

func TestDB_ErrFailedOpenDB(t *testing.T) {
	dbx := sqlmod.New(
		sqlmod.WithDSN("postgres", "user=foo dbname=bar sslmode=disable"),
	)
	require.ErrorIs(t, dbx.Init(), sqlmod.ErrFailedOpenDB)
}

func TestDB_ErrDBNotSet(t *testing.T) {
	dbx := sqlmod.New()
	require.ErrorIs(t, dbx.Init(), sqlmod.ErrDBNotSet)
}
