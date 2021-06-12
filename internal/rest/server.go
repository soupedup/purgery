// Package rest implements the embedded REST server.
package rest

import (
	"context"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/azazeal/exit"
	"go.uber.org/zap"

	"github.com/soupedup/purgery/internal/cache"
	"github.com/soupedup/purgery/internal/common"
)

// Bind binds a TCP listener on the given address and returns a reference
// to it.
func Bind(logger *zap.Logger, on string) (l net.Listener, err error) {
	logger.Info("binding ...", zap.String("on", on))

	switch l, err = net.Listen("tcp", on); err {
	default:
		err = exit.Wrap(common.ECBind, err)

		logger.Error("failed binding.", zap.Error(err))
	case nil:
		logger.Debug("bound.")
	}

	return
}

// Serve takes ownership of l and starts serving HTTP requests from clients it
// accepts on it via h until l encounters a terminal error or ctx has been
// terminated.
//
// When Serve returns l will be closed. Contrary to similar functions of the
// http package, Serve reports nil instead of http.ErrServerClosed.
func Serve(ctx context.Context, logger *zap.Logger, l net.Listener, c *cache.Cache) (err error) {
	srv := &http.Server{
		ReadHeaderTimeout: time.Second << 3,
		IdleTimeout:       time.Minute,
		MaxHeaderBytes:    1 << 12,
		ErrorLog:          zap.NewStdLog(logger),
		Handler:           newHandler(logger, c),
	}

	var wg sync.WaitGroup
	defer wg.Wait()

	served := make(chan struct{}) // closed when Serve return

	wg.Add(1)
	go func() {
		defer wg.Done()

		wait(ctx, logger, served, srv)
	}()

	if err = srv.Serve(l); err == http.ErrServerClosed {
		err = nil
	}
	close(served)

	return err
}

func wait(ctx context.Context, logger *zap.Logger, served <-chan struct{}, srv *http.Server) {
	select {
	case <-ctx.Done():
		const wait = time.Minute >> 1
		ctx, cancel := context.WithTimeout(context.Background(), wait)
		defer cancel()

		logger.Warn("context cancelled; shutting down ...")

		_ = srv.Shutdown(ctx)
	case <-served:
		logger.Warn("stopped serving!")
	}
}
