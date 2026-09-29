package app

import (
	"context"
	"errors"
	"github.com/italo/go-api-gateway/internal/admin"
	"github.com/italo/go-api-gateway/internal/config"
	"github.com/italo/go-api-gateway/internal/gateway"
	"github.com/italo/go-api-gateway/internal/health"
	"github.com/italo/go-api-gateway/internal/observability"
	"github.com/prometheus/client_golang/prometheus"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func Serve(ctx context.Context, cfg *config.Config, logger *slog.Logger) error {
	reg := prometheus.NewRegistry()
	metrics := observability.NewMetrics(reg)
	g, e := gateway.New(cfg, logger, metrics)
	if e != nil {
		return e
	}
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()
	checker := health.New(cfg.Routes, g.Registry())
	checker.Start(ctx)
	start := time.Now()
	public := &http.Server{Addr: cfg.Server.Address, Handler: g, ReadHeaderTimeout: cfg.Server.ReadHeaderTimeout, IdleTimeout: cfg.Server.IdleTimeout, MaxHeaderBytes: cfg.Server.MaxHeaderBytes}
	adm := &http.Server{Addr: cfg.Admin.Address, Handler: admin.Handler(g, reg, cfg.Admin.EnablePprof, start), ReadHeaderTimeout: cfg.Server.ReadHeaderTimeout, IdleTimeout: cfg.Server.IdleTimeout, MaxHeaderBytes: cfg.Server.MaxHeaderBytes}
	errc := make(chan error, 2)
	go func() { errc <- public.ListenAndServe() }()
	go func() { errc <- adm.ListenAndServe() }()
	select {
	case <-ctx.Done():
	case e := <-errc:
		if !errors.Is(e, http.ErrServerClosed) {
			return e
		}
	}
	timeout, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()
	e1 := public.Shutdown(timeout)
	e2 := adm.Shutdown(timeout)
	checker.Wait()
	g.CloseIdleConnections()
	return errors.Join(e1, e2)
}
