[![](https://pkg.go.dev/badge/github.com/go-srvc/mods/httpmod.svg)](https://pkg.go.dev/github.com/go-srvc/mods/httpmod)

# httpmod

```go
package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/go-srvc/mods/httpmod"
	"github.com/go-srvc/srvc"
)

func main() {
	srvc.RunAndExit(
		httpmod.New(
			httpmod.WithServerFn(CreateServer),
		),
	)
}

func CreateServer() (*http.Server, error) {
	addr := os.Getenv("ADDR")
	if addr == "" {
		return nil, errors.New("ADDR is required")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "World!")
	})

	return &http.Server{
		Addr:    addr,
		Handler: mux,
	}, nil
}
```
