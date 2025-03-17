[![](https://pkg.go.dev/badge/github.com/go-srvc/mods/sqlxmod.svg)](https://pkg.go.dev/github.com/go-srvc/mods/sqlxmod)

# sqlxmod

sqlxmod wraps [jmoiron/sqlx](https://github.com/jmoiron/sqlx) and takes care of gracefully closing connection pool when application exits.

```go
package main

import (
	"context"
	"os"

	"github.com/XSAM/otelsql"
	"github.com/go-srvc/mods/sqlxmod"
	"github.com/go-srvc/srvc"
	semconv "go.opentelemetry.io/otel/semconv/v1.30.0"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	srvc.RunAndExit(
		New(),
	)
}

// Store embeds the sqlxmod.DB
type Store struct {
	*sqlxmod.DB
}

func New() *Store {
	db := sqlxmod.New(
		sqlxmod.WithOtel("pgx", os.Getenv("DSN"),
			otelsql.WithAttributes(semconv.DBSystemNamePostgreSQL),
		),
	)
	return &Store{DB: db}
}

func (s *Store) Healthy(ctx context.Context) error {
	return s.DB.DB().PingContext(ctx)
}
```
