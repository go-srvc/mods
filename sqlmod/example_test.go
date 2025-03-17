package sqlmod_test

import (
	"github.com/XSAM/otelsql"
	"github.com/go-srvc/mods/sqlmod"
	"github.com/go-srvc/srvc"
	semconv "go.opentelemetry.io/otel/semconv/v1.30.0"
)

func ExampleNew() {
	srvc.RunAndExit(
		sqlmod.New(
			sqlmod.WithOtel("postgres", "user=foo dbname=bar sslmode=disable",
				otelsql.WithAttributes(semconv.DBSystemNamePostgreSQL),
			),
		),
	)
}
