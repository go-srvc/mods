[![](https://pkg.go.dev/badge/github.com/go-srvc/mods/logmod.svg)](https://pkg.go.dev/github.com/go-srvc/mods/logmod)

# logmod

logmod exports logs to the otel endpoint and flushes log buffers before the application exits. After `logmod.New()` runs, `slog` calls are routed through the otelslog bridge.

```go
package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-srvc/mods/httpmod"
	"github.com/go-srvc/mods/logmod"
	"github.com/go-srvc/mods/sigmod"
	"github.com/go-srvc/srvc"
)

func main() {
	srvc.RunAndExit(
		logmod.New(),
		sigmod.New(os.Interrupt),
		httpmod.New(
			httpmod.WithAddr(":8080"),
			httpmod.WithHandler(http.HandlerFunc(hello)),
		),
	)
}

func hello(w http.ResponseWriter, r *http.Request) {
	slog.InfoContext(r.Context(), "request", "path", r.URL.Path)
	fmt.Fprint(w, "ok")
}
```
