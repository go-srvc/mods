package sqlmod_test

import (
	"github.com/go-srvc/mods/sqlmod"
	"github.com/go-srvc/srvc"
)

func ExampleNew() {
	srvc.RunAndExit(
		sqlmod.New(
			sqlmod.WithDSN("postgres", "user=foo dbname=bar sslmode=disable"),
		),
	)
}
