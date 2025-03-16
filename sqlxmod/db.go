// Package sqlxmod wraps https://pkg.go.dev/github.com/jmoiron/sqlx as a module.
package sqlxmod

import (
	"fmt"

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

func (db *DB) Init() error {
	db.done = make(chan struct{})
	for _, opt := range db.opts {
		if err := opt(db); err != nil {
			return fmt.Errorf("failed to apply option: %w", err)
		}
	}

	if db.dbx == nil {
		return ErrDBNotSet
	}

	return nil
}

func (db *DB) Run() error {
	<-db.done
	return nil
}

func (db *DB) Stop() error {
	defer close(db.done)
	// db.dbx.
	return nil
}

func (db *DB) ID() string { return ID }

// DB returns *sqlx.DB instance.
// This should be only call after Init.
func (db *DB) DB() *sqlx.DB { return db.dbx }

type Opt func(*DB) error

func WithDSN(driver, dsn string) Opt {
	return func(db *DB) error {
		dbx, err := sqlx.Open(driver, dsn)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrFailedOpenDB, err)
		}
		db.dbx = dbx
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
	return func(db *DB) error {
		dbx, err := fn()
		if err != nil {
			return err
		}
		db.dbx = dbx
		return nil
	}
}
