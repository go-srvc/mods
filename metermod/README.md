[![](https://pkg.go.dev/badge/github.com/go-srvc/mods/metermod.svg)](https://pkg.go.dev/github.com/go-srvc/mods/metermod)

# metermod

Using metermod takes care of exporting metrics to otel endpoint and flushing buffers before the application exits.

```go
package main

import (
	"github.com/go-srvc/mods/metermod"
	"github.com/go-srvc/srvc"
)

func main() {
	srvc.RunAndExit(
		// By default metermod exports go runtime metrics to otel endpoint.
		metermod.New(),
	)
}

```
