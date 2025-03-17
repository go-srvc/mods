// Package tracemod provides OpenTelemetry trace provider as a module.
package tracemod

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/trace"
)

const ID = "tracemod"

const (
	ErrMissingProvider = errStr("trace provider not set")
	ErrFlushFailed     = errStr("failed to flush remaining spans")
)

type errStr string

func (e errStr) Error() string { return string(e) }

type Provider struct {
	provider   *trace.TracerProvider
	propagator propagation.TextMapPropagator
	done       chan struct{}
	opts       []Opt
}

// New creates tracer provider module with sane defaults.
// Default options are: WithEnv.
func New(opts ...Opt) *Provider {
	if len(opts) == 0 {
		opts = []Opt{WithEnv()}
	}
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
	flushErr := p.provider.ForceFlush(context.Background())
	if flushErr != nil {
		flushErr = fmt.Errorf("%w: %w", ErrFlushFailed, flushErr)
	}
	shutdowErr := p.provider.Shutdown(context.Background())
	return errors.Join(flushErr, shutdowErr)
}

func (p *Provider) ID() string { return ID }

type Opt func(*Provider) error

// WithProvider sets the underlying trace provider for module.
func WithProvider(exp *trace.TracerProvider) Opt {
	return WithProviderFn(func() (*trace.TracerProvider, error) {
		return exp, nil
	})
}

// WithProviderFn sets the underlying trace provider for module using given function.
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

// WithHTTP creates meter provider with periodic reader using http exporter from OTEL_* env configs.
// Env variables: https://pkg.go.dev/go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp
func WithHTTP() Opt {
	return func(p *Provider) error {
		exp, err := otlptracehttp.New(context.Background())
		if err != nil {
			return fmt.Errorf("failed to create http exporter: %w", err)
		}
		p.provider = trace.NewTracerProvider(trace.WithBatcher(exp))
		return nil
	}
}

// WithGRPC creates meter provider with periodic reader using grpc exporter from OTEL_* env configs.
// Env variables: https://pkg.go.dev/go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc
func WithGRPC() Opt {
	return func(p *Provider) error {
		exp, err := otlptracegrpc.New(context.Background())
		if err != nil {
			return fmt.Errorf("failed to create grpc exporter: %w", err)
		}
		p.provider = trace.NewTracerProvider(trace.WithBatcher(exp))
		return nil
	}
}

// WithStdout creates trace provider with stdout exporter.
func WithStdout(opt ...stdouttrace.Option) Opt {
	return func(p *Provider) error {
		exp, err := stdouttrace.New()
		if err != nil {
			return fmt.Errorf("failed to create stdout exporter: %w", err)
		}
		p.provider = trace.NewTracerProvider(trace.WithBatcher(exp))
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

// WithPropagator allows setting custom text map propagator.
func WithPropagator(prop propagation.TextMapPropagator) Opt {
	return func(p *Provider) error {
		p.propagator = prop
		return nil
	}
}
