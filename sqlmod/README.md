[![](https://pkg.go.dev/badge/github.com/go-srvc/mods/sqlmod.svg)](https://pkg.go.dev/github.com/go-srvc/mods/sqlmod)

# sqlmod

sqlmod takes care of gracefully closing connection pool when application exits.

```go
package main

import (
	"context"
	"os"

	"github.com/XSAM/otelsql"
	"github.com/go-srvc/mods/sqlmod"
	"github.com/go-srvc/srvc"
	semconv "go.opentelemetry.io/otel/semconv/v1.30.0"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	srvc.RunAndExit(
		New(),
	)
}

// Store embeds the sqlmod.DB
type Store struct {
	*sqlmod.DB
}

func New() *Store {
	db := sqlmod.New(
		sqlmod.WithOtel("pgx", os.Getenv("DSN"),
			otelsql.WithAttributes(semconv.DBSystemNamePostgreSQL),
		),
	)
	return &Store{DB: db}
}

func (s *Store) Healthy(ctx context.Context) error {
	return s.DB.DB().PingContext(ctx)
}
```
