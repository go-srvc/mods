[![](https://pkg.go.dev/badge/github.com/go-srvc/mods/sigmod.svg)](https://pkg.go.dev/github.com/go-srvc/mods/sigmod)

# sigmod

```go
package main

import (
	"net/http"
	"os"

	"github.com/go-srvc/mods/httpmod"
	"github.com/go-srvc/mods/sigmod"
	"github.com/go-srvc/srvc"
)

func main() {
	srvc.RunAndExit(
		// Register the signal module to handle SIGINT
		// and trigger graceful shutdown of all modules.
		sigmod.New(os.Interrupt),
		httpmod.New(
			httpmod.WithAddr(":8080"),
			httpmod.WithHandler(http.DefaultServeMux),
		),
	)
}
```
