package sqlxmod_test

import (
	"github.com/XSAM/otelsql"
	"github.com/go-srvc/mods/sqlxmod"
	"github.com/go-srvc/srvc"
	semconv "go.opentelemetry.io/otel/semconv/v1.30.0"
)

func ExampleNew() {
	srvc.RunAndExit(
		sqlxmod.New(
			sqlxmod.WithOtel("postgres", "user=foo dbname=bar sslmode=disable",
				otelsql.WithAttributes(semconv.DBSystemNamePostgreSQL),
			),
		),
	)
}
