[![](https://pkg.go.dev/badge/github.com/go-srvc/mods/tracemod.svg)](https://pkg.go.dev/github.com/go-srvc/mods/tracemod)

# tracemod

Using tracemod takes care of exporting spans to otel endpoint and flushing buffers before the application exits.

```go
package main

import (
	"github.com/go-srvc/mods/tracemod"
	"github.com/go-srvc/srvc"
)

func main() {
	srvc.RunAndExit(
		tracemod.New(),
	)
}
```
