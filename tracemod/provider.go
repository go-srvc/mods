// Package tracemod provides OpenTelemetry trace provider as a module.
package tracemod

import (
	"context"
	"fmt"
	"log/slog"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/trace"
)

const ID = "tracemod"

const ErrMissingProvider = errStr("trace provider not set")

type errStr string

func (e errStr) Error() string { return string(e) }

type Provider struct {
	provider   *trace.TracerProvider
	propagator propagation.TextMapPropagator
	done       chan struct{}
	opts       []Opt
}

// New creates tracer provider module with given options.
func New(opts ...Opt) *Provider {
	return &Provider{opts: opts}
}

func (p *Provider) Init() error {
	p.done = make(chan struct{})
	p.propagator = propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)

	for _, opt := range p.opts {
		if err := opt(p); err != nil {
			return fmt.Errorf("failed to apply option: %w", err)
		}
	}

	if p.provider == nil {
		return ErrMissingProvider
	}

	otel.SetTracerProvider(p.provider)
	otel.SetTextMapPropagator(p.propagator)
	return nil
}

func (p *Provider) Run() error {
	<-p.done
	return nil
}

func (p *Provider) Stop() error {
	close(p.done)
	err := p.provider.ForceFlush(context.Background())
	if err != nil {
		slog.Warn("failed to export remaining spans",
			slog.String("error", err.Error()),
		)
	}
	return p.provider.Shutdown(context.Background())
}

func (p *Provider) ID() string { return ID }

type Opt func(*Provider) error

// WithProvider sets the underlying trace exporter for module.
func WithProvider(exp *trace.TracerProvider) Opt {
	return WithProviderFn(func() (*trace.TracerProvider, error) {
		return exp, nil
	})
}

// WithProviderFn sets the underlying trace exporter for module.
func WithProviderFn(fn func() (*trace.TracerProvider, error)) Opt {
	return func(p *Provider) error {
		prov, err := fn()
		if err != nil {
			return err
		}
		p.provider = prov
		return nil
	}
}

// WithDefaults creates GRPC exporter with batcher using OTEL_* env configs.
// Env spec: https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/configuration/sdk-environment-variables.md
func WithDefaults() Opt {
	return func(p *Provider) error {
		exp, err := otlptracegrpc.New(context.Background())
		if err != nil {
			return fmt.Errorf("failed to create grpc exporter: %w", err)
		}
		p.provider = trace.NewTracerProvider(trace.WithBatcher(exp))
		return nil
	}
}

// WithPropagator allows setting custom text map propagator.
func WithPropagator(prop propagation.TextMapPropagator) Opt {
	return func(p *Provider) error {
		p.propagator = prop
		return nil
	}
}
