// Package httpmod provides http server as module.
package httpmod

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"
)

const ID = "httpmod"

type Opt func(s *Server) error

type Server struct {
	srv             *http.Server
	ln              net.Listener
	shutdownTimeout time.Duration
	opts            []Opt
	url             string
}

// New creates Server with given options.
func New(opts ...Opt) *Server {
	return &Server{
		opts: opts,
	}
}

// Init starts net.Listener after applying all options.
// Options are applied in same order as they were provided.
func (s *Server) Init() error {
	s.srv = &http.Server{ReadHeaderTimeout: time.Second * 10}
	s.shutdownTimeout = time.Minute
	for _, opt := range s.opts {
		if err := opt(s); err != nil {
			return fmt.Errorf("failed to apply option: %w", err)
		}
	}

	ln, err := net.Listen("tcp", s.srv.Addr)
	if err != nil {
		return fmt.Errorf("failed to init listener: %w", err)
	}

	s.ln = ln

	if s.srv.TLSConfig == nil {
		s.url = "http://" + s.ln.Addr().String()
	} else {
		s.url = "https://" + s.ln.Addr().String()
	}
	return nil
}

// URL returns server's URL and can be called after initialization.
func (s *Server) URL() string {
	return s.url
}

// Run starts serving http request and can be called after initialization.
func (s *Server) Run() error {
	err := s.srv.Serve(s.ln)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// Stop calls shutdown for server.
func (s *Server) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
	defer cancel()
	return s.srv.Shutdown(ctx)
}

func (s *Server) ID() string { return ID }

// WithServer sets http.Server for module.
func WithServer(srv *http.Server) Opt {
	return WithServerFn(func() (*http.Server, error) {
		return srv, nil
	})
}

// WithServerFn sets http.Server using value returned from fn.
func WithServerFn(fn func() (*http.Server, error)) Opt {
	return func(s *Server) error {
		srv, err := fn()
		if err != nil {
			return err
		}
		s.srv = srv
		return nil
	}
}

// WithAddr sets http.Server.Addr.
func WithAddr(addr string) Opt {
	return WithAddrFn(func() (string, error) {
		return addr, nil
	})
}

// WithAddrFn sets http.Server.Addr using value returned from fn.
func WithAddrFn(fn func() (string, error)) Opt {
	return func(s *Server) error {
		addr, err := fn()
		if err != nil {
			return err
		}
		s.srv.Addr = addr
		return nil
	}
}

// WithHandler sets http.Server.Handler.
func WithHandler(h http.Handler) Opt {
	return WithHandlerFn(func() (http.Handler, error) {
		return h, nil
	})
}

// WithHandlerFn sets http.Server.Handler using value returned from fn.
func WithHandlerFn(fn func() (http.Handler, error)) Opt {
	return func(s *Server) error {
		h, err := fn()
		if err != nil {
			return err
		}
		s.srv.Handler = h
		return nil
	}
}

// WithShutdownTimeout sets timeout for graceful shutdown.
func WithShutdownTimeout(d time.Duration) Opt {
	return func(s *Server) error {
		s.shutdownTimeout = d
		return nil
	}
}
