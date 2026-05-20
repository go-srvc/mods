[![](https://pkg.go.dev/badge/github.com/go-srvc/mods/sqlmod.svg)](https://pkg.go.dev/github.com/go-srvc/mods/sqlmod)

# sqlmod

sqlmod wraps `database/sql` as a module and closes the connection pool gracefully when the application exits.

```go
package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/XSAM/otelsql"
	"github.com/go-srvc/mods/httpmod"
	"github.com/go-srvc/mods/sigmod"
	"github.com/go-srvc/mods/sqlmod"
	"github.com/go-srvc/srvc"
	semconv "go.opentelemetry.io/otel/semconv/v1.30.0"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	db := sqlmod.New(
		sqlmod.WithOtel("pgx", os.Getenv("DSN"),
			otelsql.WithAttributes(semconv.DBSystemNamePostgreSQL),
		),
	)
	srvc.RunAndExit(
		sigmod.New(os.Interrupt),
		db,
		httpmod.New(
			httpmod.WithAddr(":8080"),
			httpmod.WithHandler(handler(db)),
		),
	)
}

func handler(db *sqlmod.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var version string
		if err := db.DB().QueryRowContext(r.Context(), "SELECT version()").Scan(&version); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "db version: %s\n", version)
	})
}
```
