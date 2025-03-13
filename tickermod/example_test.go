package tickermod_test

import (
	"errors"
	"log/slog"
	"time"

	"github.com/go-srvc/mods/tickermod"
	"github.com/go-srvc/srvc"
)

func ExampleNew() {
	srvc.RunAndExit(
		tickermod.New(
			tickermod.WithInterval(time.Second),
			tickermod.WithFunc(func() error {
				slog.Info("Hello from ticker")
				return errors.New("ticker error")
			}),
		),
	)
}
