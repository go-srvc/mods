[![codecov](https://codecov.io/github/go-srvc/mods/graph/badge.svg?token=H3u7Ui9PfC)](https://codecov.io/github/go-srvc/mods) ![main](https://github.com/go-srvc/mods/actions/workflows/go.yaml/badge.svg?branch=main)

# Modules for [srvc](https://github.com/go-srvc/srvc)

Each of the modules implements [srvc.Module](https://pkg.go.dev/github.com/go-srvc/srvc#Module) interface and can be used with [srvc.Run](https://pkg.go.dev/github.com/go-srvc/srvc#Run).

## Example

```go
package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/go-srvc/mods/httpmod"
	"github.com/go-srvc/mods/logmod"
	"github.com/go-srvc/mods/metermod"
	"github.com/go-srvc/mods/sigmod"
	"github.com/go-srvc/mods/sqlxmod"
	"github.com/go-srvc/mods/tracemod"
	"github.com/go-srvc/srvc"
)

func main() {
	db := sqlxmod.New()
	srvc.RunAndExit(
		sigmod.New(os.Interrupt),
		logmod.New(),
		tracemod.New(),
		metermod.New(),
		db,
		httpmod.New(
			httpmod.WithHandler(
				handler(db),
			),
		),
	)
}

func handler(db *sqlxmod.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := db.DB().PingContext(r.Context()); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprint(w, "OK")
	})
}
```

## Modules:

### [httpmod](https://github.com/go-srvc/mods/blob/main/httpmod)

HTTP server module wrapping `http.Server`.

### [logmod](https://github.com/go-srvc/mods/blob/main/logmod)

Wrapper for otel log provider.

### [metermod](https://github.com/go-srvc/mods/blob/main/metermod)

Wrapper for otel metrics provider.

### [sigmod](https://github.com/go-srvc/mods/blob/main/sigmod)

Signal handling module for graceful application shutdown.

### [sqlmod](https://github.com/go-srvc/mods/blob/main/sqlmod)

SQL module wrapping `sql.DB`.

### [sqlxmod](https://github.com/go-srvc/mods/blob/main/sqlxmod)

SQL module wrapping `sqlx.DB`.

### [tickermod](https://github.com/go-srvc/mods/blob/main/tickermod)

Ticker module wrapping `time.Ticker`.

### [tracemod](https://github.com/go-srvc/mods/blob/main/tracemod)

Wrapper for otel tracer provider.
