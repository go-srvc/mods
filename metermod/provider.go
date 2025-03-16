// Package metermod provides OpenTelemetry meter provider as a module.
package metermod

import (
	"context"
	"fmt"
	"log/slog"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/sdk/metric"
)

const ID = "metermod"

const ErrMissingProvider = errStr("meter provider not set")

type errStr string

func (e errStr) Error() string { return string(e) }

type Provider struct {
	provider *metric.MeterProvider
	done     chan struct{}
	opts     []Opt
}

// New creates tracer provider module with given options.
func New(opts ...Opt) *Provider {
	return &Provider{opts: opts}
}

func (p *Provider) Init() error {
	p.done = make(chan struct{})
	for _, opt := range p.opts {
		if err := opt(p); err != nil {
			return fmt.Errorf("failed to apply option: %w", err)
		}
	}

	if p.provider == nil {
		return ErrMissingProvider
	}

	otel.SetMeterProvider(p.provider)
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

// WithProvider sets the underlying meter provider for module.
func WithProvider(exp *metric.MeterProvider) Opt {
	return WithProviderFn(func() (*metric.MeterProvider, error) {
		return exp, nil
	})
}

// WithProviderFn sets the underlying meter provider for module using given function.
func WithProviderFn(fn func() (*metric.MeterProvider, error)) Opt {
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
		exp, err := otlpmetrichttp.New(context.Background())
		if err != nil {
			return fmt.Errorf("failed to create http exporter: %w", err)
		}
		p.provider = metric.NewMeterProvider(metric.WithReader(metric.NewPeriodicReader(exp)))
		return nil
	}
}
