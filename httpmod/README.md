[![](https://pkg.go.dev/badge/github.com/go-srvc/mods/httpmod.svg)](https://pkg.go.dev/github.com/go-srvc/mods/httpmod)

# httpmod

httpmod runs an `http.Server` as a srvc module and shuts it down gracefully on exit.

```go
package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/go-srvc/mods/httpmod"
	"github.com/go-srvc/mods/sigmod"
	"github.com/go-srvc/srvc"
)

func main() {
	srvc.RunAndExit(
		sigmod.New(os.Interrupt),
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
