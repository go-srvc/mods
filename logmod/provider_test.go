package logmod_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/go-srvc/mods/logmod"
	"github.com/heppu/errgroup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/sdk/log"
)

func TestProvider(t *testing.T) {
	test := []struct {
		name         string
		opts         []logmod.Opt
		contentType  string
		expectCalled bool
	}{
		{
			name:         "Defaults",
			contentType:  "application/x-protobuf",
			expectCalled: false,
		},
		{
			name:         "WithHTTP",
			opts:         []logmod.Opt{logmod.WithHTTP()},
			contentType:  "application/x-protobuf",
			expectCalled: true,
		},
		{
			// We dont have proper GRPC server so exporter will retry until it reached OTEL_EXPORTER_OTLP_LOGS_TIMEOUT
			name:         "WithGRPC",
			opts:         []logmod.Opt{logmod.WithGRPC()},
			contentType:  "",
			expectCalled: true,
		},
	}

	for _, tt := range test {
		t.Run(tt.name, func(t *testing.T) {
			called := &atomic.Bool{}
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called.Store(true)
				assert.Equal(t, tt.contentType, r.Header.Get("Content-Type"))
				io.Copy(io.Discard, r.Body) //nolint: errcheck
			}))
			t.Cleanup(srv.Close)

			t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", srv.URL)
			t.Setenv("OTEL_EXPORTER_OTLP_LOGS_TIMEOUT", "10") // Faster timeout for GRPC test

			p := logmod.New(tt.opts...)
			require.NoError(t, p.Init())
			wg := &errgroup.ErrGroup{}
			wg.Go(p.Run)

			l := otelslog.NewLogger("test")
			l.Info("test")

			require.NoError(t, p.Stop())
			require.NoError(t, wg.Wait())
			require.Equal(t, "logmod", p.ID())
			require.Equal(t, tt.expectCalled, called.Load())
		})
	}
}

func TestProvider_WithProvider_nil(t *testing.T) {
	p := logmod.New(logmod.WithProvider(nil))
	require.ErrorIs(t, p.Init(), logmod.ErrMissingProvider)
}

func TestProvider_WithProviderFn_Err(t *testing.T) {
	testErr := errors.New("test")
	p := logmod.New(logmod.WithProviderFn(func() (*log.LoggerProvider, error) {
		return nil, testErr
	}))
	require.ErrorIs(t, p.Init(), testErr)
}

func TestProvider_SetsGlobalProvider(t *testing.T) {
	exp, err := otlploghttp.New(context.Background())
	require.NoError(t, err)
	lp := log.NewLoggerProvider(log.WithProcessor(log.NewBatchProcessor(exp)))

	p := logmod.New(logmod.WithProvider(lp))
	require.NoError(t, p.Init())

	gp := global.GetLoggerProvider()
	require.Equal(t, lp, gp)
}
