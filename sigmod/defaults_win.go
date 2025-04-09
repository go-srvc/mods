//go:build windows

package sigmod

import "os"

var defaultSignals = []os.Signal{os.Kill}
