[![](https://pkg.go.dev/badge/github.com/go-srvc/mods/sigmod.svg)](https://pkg.go.dev/github.com/go-srvc/mods/sigmod)

# sigmod

sigmod listens for OS signals and returns from Run when one is received, triggering srvc shutdown.

```go
package main

import (
	"fmt"
	"net/http"
	"os"
	"syscall"

	"github.com/go-srvc/mods/httpmod"
	"github.com/go-srvc/mods/sigmod"
	"github.com/go-srvc/srvc"
)

func main() {
	srvc.RunAndExit(
		// Trigger graceful shutdown of all modules on SIGINT or SIGTERM.
		sigmod.New(os.Interrupt, syscall.SIGTERM),
		httpmod.New(
			httpmod.WithAddr(":8080"),
			httpmod.WithHandler(http.HandlerFunc(hello)),
		),
	)
}

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "hello, world")
}
```
