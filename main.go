package main

import (
	"context"
	"net"
	"os/signal"
	"sync"
	"syscall"

	"github.com/azazeal/exit"
	"go.uber.org/zap"

	"github.com/soupedup/purgery/internal/cache"
	"github.com/soupedup/purgery/internal/env"
	"github.com/soupedup/purgery/internal/log"
	"github.com/soupedup/purgery/internal/purge"
	"github.com/soupedup/purgery/internal/rest"
)

const (
	_ = iota + 1
	ecLoadConfig
	ecDialCache
)

func main() {
	exit.With(run())
}

func run() (err error) {
	logger := log.New("")

	var cfg *env.Config
	if cfg, err = env.LoadConfig(logger); err != nil {
		return
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	cache := cache.New(cfg.PurgeryID, cfg.Redis)
	defer closeCache(logger, cache)

	var l net.Listener
	if l, err = rest.Bind(logger, cfg.Addr); err != nil {
		return
	}
	// we don't need to close the listener as rest.Serve will.

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer cancel()

		purge.New(cfg.VarnishAddr).
			Run(ctx, logger, cache)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer cancel()

		err = rest.Serve(ctx, logger, l, cache, cfg.APIKey)
	}()

	wg.Wait()

	return
}

func closeCache(logger *zap.Logger, cache *cache.Cache) {
	logger.Info("closing cache ...")

	if err := cache.Close(); err != nil {
		logger.Error("failed closing cache.",
			zap.Error(err))

		return
	}

	logger.Debug("cache closed.")
}
