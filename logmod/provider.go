// Package logmod provides OpenTelemetry trace provider as a module.
package logmod

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/sdk/log"
)

const ID = "logmod"

const ErrMissingProvider = errStr("meter provider not set")

type errStr string

func (e errStr) Error() string { return string(e) }

type Provider struct {
	provider *log.LoggerProvider
	done     chan struct{}
	opts     []Opt
}

// New creates log provider module with sane defaults if no options are provided.
// Default options are: WithEnv and WithSlog .
// For instructions on how to integrate log provider with logger of your choice,
// see the list of ready made bridge libraries:
// https://opentelemetry.io/ecosystem/registry/?language=go&component=log-bridge
func New(opts ...Opt) *Provider {
	if len(opts) == 0 {
		opts = []Opt{WithEnv(), WithSlog()}
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

	global.SetLoggerProvider(p.provider)
	fmt.Println("init done")
	return nil
}

func (p *Provider) Run() error {
	<-p.done
	return nil
}

func (p *Provider) Stop() error {
	close(p.done)
	flushErr := p.provider.ForceFlush(context.Background())
	shutdowErr := p.provider.Shutdown(context.Background())
	return errors.Join(flushErr, shutdowErr)
}

func (p *Provider) ID() string { return ID }

type Opt func(*Provider) error

// WithProvider sets the underlying trace provider for module.
func WithProvider(exp *log.LoggerProvider) Opt {
	return WithProviderFn(func() (*log.LoggerProvider, error) {
		return exp, nil
	})
}

// WithProviderFn sets the underlying trace provider for module using given function.
func WithProviderFn(fn func() (*log.LoggerProvider, error)) Opt {
	return func(p *Provider) error {
		prov, err := fn()
		if err != nil {
			return err
		}
		p.provider = prov
		return nil
	}
}

// WithHTTP creates log provider with batch processor using http exporter from OTEL_* env configs.
// Env variables: https://pkg.go.dev/go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp
func WithHTTP() Opt {
	return func(p *Provider) error {
		exp, err := otlploghttp.New(context.Background())
		if err != nil {
			return fmt.Errorf("failed to create http exporter: %w", err)
		}
		p.provider = log.NewLoggerProvider(log.WithProcessor(log.NewBatchProcessor(exp)))
		return nil
	}
}

// WithGRPC creates log provider with batch processor using grpc exporter from OTEL_* env configs.
// Env variables: https://pkg.go.dev/go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc
func WithGRPC() Opt {
	return func(p *Provider) error {
		exp, err := otlploggrpc.New(context.Background())
		if err != nil {
			return fmt.Errorf("failed to create grpc exporter: %w", err)
		}
		p.provider = log.NewLoggerProvider(log.WithProcessor(log.NewBatchProcessor(exp)))
		return nil
	}
}

// WithStdout creates log provider with stdout exporter.
func WithStdout(opt ...stdoutlog.Option) Opt {
	return func(p *Provider) error {
		exp, err := stdoutlog.New(opt...)
		if err != nil {
			return fmt.Errorf("failed to create stdout exporter: %w", err)
		}
		p.provider = log.NewLoggerProvider(log.WithProcessor(log.NewBatchProcessor(exp)))
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

// WithSlog sets slog's default logger to otelslog logger.
// By default logger will be named "default" or taken from OTEL_SERVICE_NAME env variable if present.
func WithSlog() Opt {
	return func(p *Provider) error {
		l := otelslog.NewLogger(cmp.Or(os.Getenv("OTEL_SERVICE_NAME"), "default"), otelslog.WithLoggerProvider(global.GetLoggerProvider()))
		slog.SetDefault(l)
		return nil
	}
}
