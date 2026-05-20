[![](https://pkg.go.dev/badge/github.com/go-srvc/mods/metermod.svg)](https://pkg.go.dev/github.com/go-srvc/mods/metermod)

# metermod

metermod exports metrics to the otel endpoint and flushes buffers before the application exits. By default it also collects Go runtime metrics.

```go
package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/go-srvc/mods/httpmod"
	"github.com/go-srvc/mods/metermod"
	"github.com/go-srvc/mods/sigmod"
	"github.com/go-srvc/srvc"
)

func main() {
	srvc.RunAndExit(
		metermod.New(),
		sigmod.New(os.Interrupt),
		httpmod.New(
			httpmod.WithAddr(":8080"),
			httpmod.WithHandler(http.HandlerFunc(hello)),
		),
	)
}

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "ok")
}
```
