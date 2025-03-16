package logmod_test

import (
	"github.com/go-srvc/mods/logmod"
	"github.com/go-srvc/srvc"
)

func ExampleNew() {
	srvc.RunAndExit(
		logmod.New(),
	)
}
