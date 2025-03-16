package tracemod_test

import (
	"github.com/go-srvc/mods/tracemod"
	"github.com/go-srvc/srvc"
)

func ExampleNew() {
	srvc.RunAndExit(
		tracemod.New(
			tracemod.WithDefaults(),
		),
	)
}
