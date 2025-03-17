package sqlxmod_test

import (
	"testing"

	"github.com/XSAM/otelsql"
	"github.com/go-srvc/mods/sqlxmod"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	semconv "go.opentelemetry.io/otel/semconv/v1.30.0"

	_ "github.com/lib/pq"
)

func TestDB(t *testing.T) {
	db := &sqlx.DB{}
	dbx := sqlxmod.New(
		sqlxmod.WithDBx(db),
	)
	require.NoError(t, dbx.Init())
	assert.Equal(t, db, dbx.DB())

	// TODO: Test against the real database so close works
	// wg := &errgroup.ErrGroup{}
	// wg.Go(dbx.Run)
	// require.NoError(t, dbx.Stop())
	// require.NoError(t, wg.Wait())
	require.Equal(t, "sqlxmod", dbx.ID())
}

func TestDB_ErrFailedOpenDB(t *testing.T) {
	dbx := sqlxmod.New(
		sqlxmod.WithDSN("not valid driver", ""),
	)
	require.ErrorIs(t, dbx.Init(), sqlxmod.ErrFailedOpenDB)
}

func TestDB_ErrDBNotSet(t *testing.T) {
	dbx := sqlxmod.New()
	require.ErrorIs(t, dbx.Init(), sqlxmod.ErrDBNotSet)
}

func TestDB_WithOtel(t *testing.T) {
	dbx := sqlxmod.New(
		sqlxmod.WithOtel("postgres", "user=foo dbname=bar sslmode=disable", otelsql.WithAttributes(semconv.DBSystemNamePostgreSQL)),
	)
	require.NoError(t, dbx.Init())
}
