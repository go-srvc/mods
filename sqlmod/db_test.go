package sqlmod_test

import (
	"database/sql"
	"testing"

	"github.com/XSAM/otelsql"
	"github.com/go-srvc/mods/sqlmod"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	semconv "go.opentelemetry.io/otel/semconv/v1.30.0"

	_ "github.com/lib/pq"
)

func TestDB(t *testing.T) {
	db := &sql.DB{}
	dbx := sqlmod.New(
		sqlmod.WithDB(db),
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
		sqlmod.WithDSN("not valid driver", ""),
	)
	require.ErrorIs(t, dbx.Init(), sqlmod.ErrFailedOpenDB)
}

func TestDB_ErrDBNotSet(t *testing.T) {
	dbx := sqlmod.New()
	require.ErrorIs(t, dbx.Init(), sqlmod.ErrDBNotSet)
}

func TestDB_WithOtel(t *testing.T) {
	dbx := sqlmod.New(
		sqlmod.WithOtel("postgres", "user=foo dbname=bar sslmode=disable", otelsql.WithAttributes(semconv.DBSystemNamePostgreSQL)),
	)
	require.NoError(t, dbx.Init())
}
