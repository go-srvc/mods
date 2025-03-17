// Package metermod provides OpenTelemetry meter provider as a module.
package metermod

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/sdk/metric"
)

const ID = "metermod"

const (
	ErrMissingProvider = errStr("meter provider not set")
	ErrFlushFailed     = errStr("failed to flush remaining metrics")
)

type errStr string

func (e errStr) Error() string { return string(e) }

type Provider struct {
	provider *metric.MeterProvider
	done     chan struct{}
	opts     []Opt
}

// New creates meter provider module with sane defaults.
// Default options are: WithEnv and WithRuntimeMetrics.
func New(opts ...Opt) *Provider {
	if len(opts) == 0 {
		opts = []Opt{WithEnv(), WithRuntimeMetrics()}
	}
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
	flushErr := p.provider.ForceFlush(context.Background())
	if flushErr != nil {
		flushErr = fmt.Errorf("%w: %w", ErrFlushFailed, flushErr)
	}
	shutdowErr := p.provider.Shutdown(context.Background())
	return errors.Join(flushErr, shutdowErr)
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

// WithHTTP creates meter provider with periodic reader using http exporter from OTEL_* env configs.
// Env variables: https://pkg.go.dev/go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp
func WithHTTP() Opt {
	return func(p *Provider) error {
		exp, err := otlpmetrichttp.New(context.Background())
		if err != nil {
			return fmt.Errorf("failed to create http exporter: %w", err)
		}
		p.provider = metric.NewMeterProvider(metric.WithReader(metric.NewPeriodicReader(exp)))
		return nil
	}
}

// WithGRPC creates meter provider with periodic reader using grpc exporter from OTEL_* env configs.
// Env variables: https://pkg.go.dev/go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc
func WithGRPC() Opt {
	return func(p *Provider) error {
		exp, err := otlpmetricgrpc.New(context.Background())
		if err != nil {
			return fmt.Errorf("failed to create grpc exporter: %w", err)
		}
		p.provider = metric.NewMeterProvider(metric.WithReader(metric.NewPeriodicReader(exp)))
		return nil
	}
}

// WithStdout creates meter provider with stdout exporter.
func WithStdout(opt ...stdoutmetric.Option) Opt {
	return func(p *Provider) error {
		exp, err := stdoutmetric.New()
		if err != nil {
			return fmt.Errorf("failed to create stdout exporter: %w", err)
		}
		p.provider = metric.NewMeterProvider(metric.WithReader(metric.NewPeriodicReader(exp)))
		return nil
	}
}

// WithEnv uses OTEL_EXPORTER_OTLP_TRACES_PROTO and OTEL_EXPORTER_OTLP_PROTO environment variable to set exporter.
// Accepted values are:
//   - http
//   - grpc
//   - stdout
//
// If no value is provided, stdout is used.
func WithEnv() Opt {
	return func(p *Provider) error {
		switch strings.ToLower(cmp.Or(os.Getenv("OTEL_EXPORTER_OTLP_TRACES_PROTO"), os.Getenv("OTEL_EXPORTER_OTLP_PROTO"))) {
		case "http":
			return WithHTTP()(p)
		case "grpc":
			return WithGRPC()(p)
		default:
			return WithStdout()(p)
		}
	}
}

// WithRuntimeMetrics starts runtime metrics collection.
func WithRuntimeMetrics(opts ...runtime.Option) Opt {
	return func(p *Provider) error {
		return runtime.Start(opts...)
	}
}
