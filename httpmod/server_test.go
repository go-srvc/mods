package httpmod_test

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/go-srvc/mods/httpmod"
	"github.com/heppu/errgroup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestErrOpt(t *testing.T) {
	optErr := errors.New("opt err")
	srv := httpmod.New(func(s *httpmod.Server) error { return optErr })
	err := srv.Init()
	require.ErrorIs(t, err, optErr)
}

func TestListenErr(t *testing.T) {
	srv := httpmod.New(httpmod.WithAddr(`sdf./43/s]\\][]"`))
	err := srv.Init()
	require.Error(t, err)
}

func TestHTTPS(t *testing.T) {
	srv := httpmod.New(httpmod.WithServer(&http.Server{
		Addr:              "127.0.0.1:0",
		ReadHeaderTimeout: time.Second,
		TLSConfig:         &tls.Config{}, //nolint:gosec
	}))
	err := srv.Init()
	require.NoError(t, err)
	require.Contains(t, srv.URL(), "https://127.0.0.1:")
}

func TestServer(t *testing.T) {
	srv := httpmod.New(
		httpmod.WithServer(&http.Server{ReadHeaderTimeout: time.Second}),
		httpmod.WithAddr("127.0.0.1:0"),
		httpmod.WithHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, "Hello")
		})),
	)

	require.Equal(t, "httpmod", srv.ID())
	require.NoError(t, srv.Init())
	wg := &errgroup.ErrGroup{}
	wg.Go(srv.Run)

	resp, err := http.Get(srv.URL())
	assert.NoError(t, err)
	data, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, "Hello", string(data))

	assert.NoError(t, srv.Stop())
	assert.NoError(t, wg.Wait())
}

func TestServerShutdownTimeout(t *testing.T) {
	block := make(chan struct{})
	srv := httpmod.New(
		httpmod.WithAddr("127.0.0.1:0"),
		httpmod.WithShutdownTimeout(time.Second),
		httpmod.WithHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			<-block
		})),
	)
	err := srv.Init()
	require.NoError(t, err)
	url := srv.URL()

	wg := &errgroup.ErrGroup{}
	wg.Go(srv.Run)

	go func() {
		resp, err := http.Get(url) //nolint:gosec
		assert.NoError(t, err)
		io.Copy(io.Discard, resp.Body) //nolint: errcheck
	}()

	time.Sleep(time.Millisecond * 10)
	err = srv.Stop()
	assert.ErrorContains(t, err, "context deadline exceeded")

	err = wg.Wait()
	assert.NoError(t, err)
}
