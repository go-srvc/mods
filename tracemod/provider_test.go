package tracemod_test

import (
	"context"
	"errors"
	"testing"

	"github.com/go-srvc/mods/tracemod"
	"github.com/heppu/errgroup"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/trace"
)

func TestProvider(t *testing.T) {
	p := tracemod.New(tracemod.WithDefaults())
	require.NoError(t, p.Init())
	wg := &errgroup.ErrGroup{}
	wg.Go(p.Run)
	require.NoError(t, p.Stop())
	require.NoError(t, wg.Wait())
	require.Equal(t, "tracemod", p.ID())
}

func TestProvider_ErrMissingProvider(t *testing.T) {
	p := tracemod.New()
	require.ErrorIs(t, p.Init(), tracemod.ErrMissingProvider)
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
		tracemod.WithDefaults(),
		tracemod.WithPropagator(prop),
	)
	require.NoError(t, p.Init())

	tmp := otel.GetTextMapPropagator()
	require.Equal(t, prop, tmp)
}
