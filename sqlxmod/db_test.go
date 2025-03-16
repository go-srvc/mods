package sqlxmod_test

import (
	"testing"

	"github.com/go-srvc/mods/sqlxmod"
	"github.com/heppu/errgroup"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDB(t *testing.T) {
	db := &sqlx.DB{}
	dbx := sqlxmod.New(
		sqlxmod.WithDBx(db),
	)
	require.NoError(t, dbx.Init())
	assert.Equal(t, db, dbx.DB())

	wg := &errgroup.ErrGroup{}
	wg.Go(dbx.Run)
	require.NoError(t, dbx.Stop())
	require.NoError(t, wg.Wait())
	require.Equal(t, "sqlxmod", dbx.ID())
}

func TestDB_ErrFailedOpenDB(t *testing.T) {
	dbx := sqlxmod.New(
		sqlxmod.WithDSN("postgres", "user=foo dbname=bar sslmode=disable"),
	)
	require.ErrorIs(t, dbx.Init(), sqlxmod.ErrFailedOpenDB)
}

func TestDB_ErrDBNotSet(t *testing.T) {
	dbx := sqlxmod.New()
	require.ErrorIs(t, dbx.Init(), sqlxmod.ErrDBNotSet)
}
