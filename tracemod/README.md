[![](https://pkg.go.dev/badge/github.com/go-srvc/mods/tracemod.svg)](https://pkg.go.dev/github.com/go-srvc/mods/tracemod)

# tracemod

tracemod exports spans to the otel endpoint and flushes buffers before the application exits. Once installed, any otel tracer in the process feeds into the same exporter.

```go
package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/go-srvc/mods/httpmod"
	"github.com/go-srvc/mods/sigmod"
	"github.com/go-srvc/mods/tracemod"
	"github.com/go-srvc/srvc"
	"go.opentelemetry.io/otel"
)

func main() {
	srvc.RunAndExit(
		tracemod.New(),
		sigmod.New(os.Interrupt),
		httpmod.New(
			httpmod.WithAddr(":8080"),
			httpmod.WithHandler(http.HandlerFunc(hello)),
		),
	)
}

func hello(w http.ResponseWriter, r *http.Request) {
	_, span := otel.Tracer("hello").Start(r.Context(), "greet")
	defer span.End()
	fmt.Fprint(w, "hello, world")
}
```
