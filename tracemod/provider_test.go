package tracemod_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/go-srvc/mods/tracemod"
	"github.com/heppu/errgroup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/trace"
)

func TestProvider(t *testing.T) {
	test := []struct {
		name        string
		opts        []tracemod.Opt
		contentType string
		err         error
	}{
		{
			name:        "Default",
			contentType: "application/x-protobuf",
		},
		{
			name:        "WithHTTP",
			opts:        []tracemod.Opt{tracemod.WithHTTP()},
			contentType: "application/x-protobuf",
		},
		{
			// We dont have proper GRPC server so exporter will retry until it reached OTEL_EXPORTER_OTLP_TRACES_TIMEOUT
			name:        "WithGRPC",
			opts:        []tracemod.Opt{tracemod.WithGRPC()},
			contentType: "",
			err:         tracemod.ErrFlushFailed,
		},
	}

	for _, tt := range test {
		t.Run(tt.name, func(t *testing.T) {
			called := &atomic.Bool{}
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called.Store(true)
				assert.Equal(t, tt.contentType, r.Header.Get("Content-Type"))
			}))
			t.Cleanup(srv.Close)

			t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", srv.URL)
			t.Setenv("OTEL_EXPORTER_OTLP_TRACES_TIMEOUT", "10") // Faster timeout for GRPC test

			p := tracemod.New(tt.opts...)
			require.NoError(t, p.Init())
			wg := &errgroup.ErrGroup{}
			wg.Go(p.Run)

			_, span := otel.GetTracerProvider().Tracer("test").Start(context.Background(), "test")
			span.End()

			require.ErrorIs(t, p.Stop(), tt.err)
			require.NoError(t, wg.Wait())
			require.Equal(t, "tracemod", p.ID())
			require.True(t, called.Load())
		})
	}
}

func TestProvider_WithProvider_nil(t *testing.T) {
	p := tracemod.New(tracemod.WithProvider(nil))
	require.ErrorIs(t, p.Init(), tracemod.ErrMissingProvider)
}

func TestProvider_WithProviderFn_Err(t *testing.T) {
	testErr := errors.New("test")
	p := tracemod.New(tracemod.WithProviderFn(func() (*trace.TracerProvider, error) {
		return nil, testErr
	}))
	require.ErrorIs(t, p.Init(), testErr)
}

func TestProvider_SetsGlobalProvider(t *testing.T) {
	exp, err := otlptracegrpc.New(context.Background())
	require.NoError(t, err)
	tracerProvider := trace.NewTracerProvider(trace.WithBatcher(exp))

	p := tracemod.New(tracemod.WithProvider(tracerProvider))
	require.NoError(t, p.Init())

	gp := otel.GetTracerProvider()
	require.Equal(t, tracerProvider, gp)
}

func TestProvider_SetsGlobalPropagator(t *testing.T) {
	prop := &propagation.TraceContext{}
	p := tracemod.New(
		tracemod.WithPropagator(prop),
	)
	require.NoError(t, p.Init())

	tmp := otel.GetTextMapPropagator()
	require.Equal(t, prop, tmp)
}
