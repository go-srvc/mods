package metermod_test

import (
	"github.com/go-srvc/mods/metermod"
	"github.com/go-srvc/srvc"
)

func ExampleNew() {
	srvc.RunAndExit(
		metermod.New(),
	)
}
