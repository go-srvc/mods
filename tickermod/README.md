[![](https://pkg.go.dev/badge/github.com/go-srvc/mods/tickermod.svg)](https://pkg.go.dev/github.com/go-srvc/mods/tickermod)

# tickermod

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
			tickermod.WithFunc(func() {
				slog.Info("Hello, World!")
			}),
		),
	)
}
```
