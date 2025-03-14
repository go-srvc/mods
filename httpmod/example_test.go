package httpmod_test

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-srvc/mods/httpmod"
	"github.com/go-srvc/srvc"
)

func ExampleNew() {
	srvc.RunAndExit(
		httpmod.New(
			httpmod.WithServer(&http.Server{ReadHeaderTimeout: time.Second}),
			httpmod.WithAddr("127.0.0.1:0"),
			httpmod.WithHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, "Hello")
			})),
		),
	)
}
