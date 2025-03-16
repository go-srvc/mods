package sqlxmod_test

import (
	"github.com/go-srvc/mods/sqlxmod"
	"github.com/go-srvc/srvc"
)

func ExampleNew() {
	srvc.RunAndExit(
		sqlxmod.New(
			sqlxmod.WithDSN("postgres", "user=foo dbname=bar sslmode=disable"),
		),
	)
}
