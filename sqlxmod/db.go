// Package sqlxmod wraps https://pkg.go.dev/github.com/jmoiron/sqlx as a module.
package sqlxmod

import (
	"database/sql"
	"fmt"

	"github.com/XSAM/otelsql"
	"github.com/jmoiron/sqlx"
)

const ID = "sqlxmod"

const (
	ErrDBNotSet     = errStr("db not set")
	ErrFailedOpenDB = errStr("failed to open db")
)

type errStr string

func (e errStr) Error() string { return string(e) }

type DB struct {
	dbx  *sqlx.DB
	done chan struct{}
	opts []Opt
}

// New creates new sqlx module with given options.
// Remember to import required driver for your database.
func New(opts ...Opt) *DB {
	return &DB{opts: opts}
}

func (d *DB) Init() error {
	d.done = make(chan struct{})
	for _, opt := range d.opts {
		if err := opt(d); err != nil {
			return fmt.Errorf("failed to apply option: %w", err)
		}
	}

	if d.dbx == nil {
		return ErrDBNotSet
	}

	return nil
}

func (d *DB) Run() error {
	<-d.done
	return nil
}

func (d *DB) Stop() error {
	defer close(d.done)
	d.dbx.Close()
	return nil
}

func (d *DB) ID() string { return ID }

// DB returns *sqlx.DB instance.
// This should be only call after Init.
func (d *DB) DB() *sqlx.DB { return d.dbx }

type Opt func(*DB) error

func WithDSN(driver, dsn string) Opt {
	return func(d *DB) error {
		dbx, err := sqlx.Open(driver, dsn)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrFailedOpenDB, err)
		}
		d.dbx = dbx
		return nil
	}
}

// WithDBx sets *sqlx.DB for module.
func WithDBx(dbx *sqlx.DB) Opt {
	return WithDBxFn(func() (*sqlx.DB, error) {
		return dbx, nil
	})
}

// WithDBxFn sets *sqlx.DB using value returned from fn.
func WithDBxFn(fn func() (*sqlx.DB, error)) Opt {
	return func(d *DB) error {
		dbx, err := fn()
		if err != nil {
			return err
		}
		d.dbx = dbx
		return nil
	}
}

// WithDB creates *sqlx.DB from *sql.DB.
func WithDB(db *sql.DB, driver string) Opt {
	return func(d *DB) error {
		d.dbx = sqlx.NewDb(db, driver)
		return nil
	}
}

// WithOtel registers *sqlx.DB with OpenTelemetry.
func WithOtel(driver, dsn string, opts ...otelsql.Option) Opt {
	return func(d *DB) error {
		db, err := otelsql.Open(driver, dsn, opts...)
		if err != nil {
			return err
		}

		err = otelsql.RegisterDBStatsMetrics(db, opts...)
		if err != nil {
			return err
		}

		return WithDB(db, driver)(d)
	}
}
