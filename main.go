package main

import (
	"context"
	"os/signal"
	"syscall"
	"time"

	"github.com/azazeal/exit"
	"go.uber.org/zap"

	"github.com/soupedup/purgery/internal/cache"
	"github.com/soupedup/purgery/internal/env"
	"github.com/soupedup/purgery/internal/log"
	"github.com/soupedup/purgery/internal/purge"
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

	var c *cache.Cache
	if c, err = cache.Dial(ctx, logger, cfg); err != nil {
		return
	}
	defer closeCache(logger, c)

	purge := purge.New(cfg.VarnishAddr)

	for ok := true; ; ok = tick(ctx, logger, c, purge) {
		// after each error sleep for a bit
		if !ok {
			const errorSleep = time.Millisecond << 6
			time.Sleep(errorSleep)
		}

		// bail if the context is no longer valid
		if err = ctx.Err(); err != nil {
			logger.Warn("context canceled; bailing ...", zap.Error(err))

			break
		}
	}

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

func tick(ctx context.Context, logger *zap.Logger, cache *cache.Cache, purge purge.Func) (ok bool) {
	var checkpoint, url string
	if checkpoint, url, ok = cache.Next(ctx, logger); ok && url != "" {
		ok = purge(ctx, logger, url) &&
			cache.Store(ctx, logger, checkpoint)
	}

	return
}
