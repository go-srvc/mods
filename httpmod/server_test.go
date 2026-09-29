package httpmod_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net"
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
	cert, roots := selfSignedCert(t)
	srv := httpmod.New(
		httpmod.WithAddr("127.0.0.1:0"),
		httpmod.WithTLSConfig(&tls.Config{Certificates: []tls.Certificate{cert}, MinVersion: tls.VersionTLS12}),
		httpmod.WithHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, r.Proto)
		})),
	)
	require.NoError(t, srv.Init())
	require.Contains(t, srv.URL(), "https://127.0.0.1:")
	wg := &errgroup.ErrGroup{}
	wg.Go(srv.Run)

	client := &http.Client{Transport: &http.Transport{
		TLSClientConfig:   &tls.Config{RootCAs: roots, MinVersion: tls.VersionTLS12},
		ForceAttemptHTTP2: true,
	}}
	resp, err := client.Get(srv.URL())
	require.NoError(t, err)
	data, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.NoError(t, resp.Body.Close())
	assert.NotNil(t, resp.TLS)
	assert.Equal(t, "HTTP/2.0", string(data))

	assert.NoError(t, srv.Stop())
	assert.NoError(t, wg.Wait())
}

func TestHTTPSWithoutCertificate(t *testing.T) {
	srv := httpmod.New(
		httpmod.WithAddr("127.0.0.1:0"),
		httpmod.WithTLSConfig(&tls.Config{MinVersion: tls.VersionTLS12}),
	)
	require.Error(t, srv.Init())
}

func selfSignedCert(t *testing.T) (tls.Certificate, *x509.CertPool) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	require.NoError(t, err)
	leaf, err := x509.ParseCertificate(der)
	require.NoError(t, err)
	roots := x509.NewCertPool()
	roots.AddCert(leaf)
	return tls.Certificate{Certificate: [][]byte{der}, PrivateKey: key, Leaf: leaf}, roots
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
