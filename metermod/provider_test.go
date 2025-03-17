package metermod_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/go-srvc/mods/metermod"
	"github.com/heppu/errgroup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/sdk/metric"
)

func TestProvider(t *testing.T) {
	test := []struct {
		name         string
		opts         []metermod.Opt
		contentType  string
		expectCalled bool
		err          error
	}{
		{
			name:         "Default",
			contentType:  "application/x-protobuf",
			expectCalled: false,
		},
		{
			name:         "WithHTTP",
			opts:         []metermod.Opt{metermod.WithHTTP()},
			contentType:  "application/x-protobuf",
			expectCalled: true,
		},
		{
			// We dont have proper GRPC server so exporter will retry until it reached OTEL_EXPORTER_OTLP_METRICS_TIMEOUT
			name:         "WithGRPC",
			opts:         []metermod.Opt{metermod.WithGRPC()},
			contentType:  "",
			expectCalled: true,
			err:          metermod.ErrFlushFailed,
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
			t.Setenv("OTEL_EXPORTER_OTLP_METRICS_TIMEOUT", "10") // Faster timeout for GRPC test

			p := metermod.New(tt.opts...)
			require.NoError(t, p.Init())
			wg := &errgroup.ErrGroup{}
			wg.Go(p.Run)

			_, span := otel.GetTracerProvider().Tracer("test").Start(context.Background(), "test")
			span.End()

			require.ErrorIs(t, p.Stop(), tt.err)
			require.NoError(t, wg.Wait())
			require.Equal(t, "metermod", p.ID())
			require.Equal(t, tt.expectCalled, called.Load())
		})
	}
}

func TestProvider_WithProvider_nil(t *testing.T) {
	p := metermod.New(metermod.WithProvider(nil))
	require.ErrorIs(t, p.Init(), metermod.ErrMissingProvider)
}

func TestProvider_WithProviderFn_Err(t *testing.T) {
	testErr := errors.New("test")
	p := metermod.New(metermod.WithProviderFn(func() (*metric.MeterProvider, error) {
		return nil, testErr
	}))
	require.ErrorIs(t, p.Init(), testErr)
}

func TestProvider_SetsGlobalProvider(t *testing.T) {
	exp, err := otlpmetrichttp.New(context.Background())
	require.NoError(t, err)
	mp := metric.NewMeterProvider(metric.WithReader(metric.NewPeriodicReader(exp)))
	p := metermod.New(metermod.WithProvider(mp))
	require.NoError(t, p.Init())

	gp := otel.GetMeterProvider()
	require.Equal(t, mp, gp)
}
