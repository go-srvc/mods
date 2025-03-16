package metermod_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/go-srvc/mods/metermod"
	"github.com/heppu/errgroup"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/sdk/metric"
)

func TestProvider(t *testing.T) {
	called := &atomic.Bool{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called.Store(true) }))
	t.Cleanup(srv.Close)
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", srv.URL)

	p := metermod.New(metermod.WithDefaults())
	require.NoError(t, p.Init())
	wg := &errgroup.ErrGroup{}
	wg.Go(p.Run)
	require.NoError(t, p.Stop())
	require.NoError(t, wg.Wait())
	require.Equal(t, "metermod", p.ID())
	require.True(t, called.Load())
}

func TestProvider_ErrMissingProvider(t *testing.T) {
	p := metermod.New()
	require.ErrorIs(t, p.Init(), metermod.ErrMissingProvider)
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
