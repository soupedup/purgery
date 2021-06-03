package main

import (
	"context"
	"os"

	"os/signal"
	"syscall"
	"time"

	"github.com/soupedup/purgery/internal/cache"
	"github.com/soupedup/purgery/internal/env"
	"github.com/soupedup/purgery/internal/log"
	"github.com/soupedup/purgery/internal/purge"
	"go.uber.org/zap"
)

const (
	_ = iota + 1
	ecLoadConfig
	ecDialCache
)

func main() {
	logger := log.New("purgery")

	cfg := env.LoadConfig(logger)
	if cfg == nil {
		os.Exit(ecLoadConfig)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	cache := cache.Dial(ctx, logger, cfg)
	if cache == nil {
		os.Exit(ecDialCache)
	}
	defer closeCache(logger, cache)

	purge := purge.New(cfg.VarnishAddr)

	for ok := true; ; ok = tick(ctx, logger, cache, purge) {
		// after each error sleep for a bit
		if !ok {
			const errorSleep = time.Millisecond << 6
			time.Sleep(errorSleep)
		}

		// bail if the context is no longer valid
		if err := ctx.Err(); err != nil {
			logger.Warn("context canceled; bailing ...", zap.Error(err))

			break
		}
	}
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

func tick(ctx context.Context, logger *zap.Logger, cache *cache.Cache, purge purge.Func) (ok bool) {
	var checkpoint, url string
	if checkpoint, url, ok = cache.Next(ctx, logger); ok && url != "" {
		ok = purge(ctx, logger, url) &&
			cache.Store(ctx, logger, checkpoint)
	}

	return
}
