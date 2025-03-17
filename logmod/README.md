[![](https://pkg.go.dev/badge/github.com/go-srvc/mods/logmod.svg)](https://pkg.go.dev/github.com/go-srvc/mods/logmod)

# logmod

Using logmod takes care of exporting logs to otel endpoint and flushing log buffers before the application exits.

```go
package main

import (
	"log/slog"
	"time"

	"github.com/go-srvc/mods/logmod"
	"github.com/go-srvc/mods/tickermod"
	"github.com/go-srvc/srvc"
)

func main() {
	srvc.RunAndExit(
		logmod.New(),
		tickermod.New(
			tickermod.WithInterval(5*time.Second),
			tickermod.WithFunc(func() {
				// Slog uses now otelslog bridge configured by logmod.
				slog.Info("Hello, World!")
			}),
		),
	)
}
```
