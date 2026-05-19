[![](https://pkg.go.dev/badge/github.com/go-srvc/mods/tickermod.svg)](https://pkg.go.dev/github.com/go-srvc/mods/tickermod)

# tickermod

tickermod runs a function on a fixed interval. The ticker stops when the function returns a non-nil error.

```go
package main

import (
	"log/slog"
	"time"

	"github.com/go-srvc/mods/tickermod"
	"github.com/go-srvc/srvc"
)

func main() {
	srvc.RunAndExit(
		tickermod.New(
			tickermod.WithInterval(5*time.Second),
			tickermod.WithFunc(func() error {
				slog.Info("Hello, World!")
				return nil
			}),
		),
	)
}
```
