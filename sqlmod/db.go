// Package sqlmod wraps database/sql as a module.
package sqlmod

import (
	"database/sql"
	"fmt"

	"github.com/XSAM/otelsql"
)

const ID = "sqlmod"

const (
	ErrDBNotSet     = errStr("db not set")
	ErrFailedOpenDB = errStr("failed to open db")
)

type errStr string

func (e errStr) Error() string { return string(e) }

type DB struct {
	db   *sql.DB
	done chan struct{}
	opts []Opt
}

// New creates new sql module with given options.
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

	if d.db == nil {
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
	return d.db.Close()
}

func (d *DB) ID() string { return ID }

// DB returns *sql.DB instance.
// This should be only call after Init.
func (d *DB) DB() *sql.DB { return d.db }

type Opt func(*DB) error

func WithDSN(driver, dsn string) Opt {
	return func(d *DB) error {
		db, err := sql.Open(driver, dsn)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrFailedOpenDB, err)
		}
		d.db = db
		return nil
	}
}

// WithDB sets *sql.DB for module.
func WithDB(db *sql.DB) Opt {
	return WithDBFn(func() (*sql.DB, error) {
		return db, nil
	})
}

// WithDBFn sets *sql.DB using value returned from fn.
func WithDBFn(fn func() (*sql.DB, error)) Opt {
	return func(d *DB) error {
		db, err := fn()
		if err != nil {
			return err
		}
		d.db = db
		return nil
	}
}

// WithOtel creates *sql.DB and instruments it with OpenTelemetry.
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

		return WithDB(db)(d)
	}
}
