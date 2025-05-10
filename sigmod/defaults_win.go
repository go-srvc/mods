//go:build windows

package sigmod

import "os"

var DefaultSignals = []os.Signal{os.Kill}
