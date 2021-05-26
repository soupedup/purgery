package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jsierles/clvr/internal/cache"
	"github.com/jsierles/clvr/internal/env"
	"github.com/jsierles/clvr/internal/purge"
)

func init() {
	log.SetOutput(os.Stderr)
	log.SetPrefix("clvr ")
	log.SetFlags(log.LUTC | log.Ldate | log.Ltime | log.Lmicroseconds | log.Lmsgprefix)
}

func main() {
	allocID, err := env.AllocID()
	if err != nil {
		log.Fatalf("failed fetching allocID: %v", err)
	}

	redisURL, err := env.RedisURL()
	if err != nil {
		log.Fatalf("failed fetching redisURL: %v", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	cache, err := cache.Dial(ctx, allocID, redisURL)
	if err != nil {
		log.Fatalf("failed dialing cache: %v", err)
	}

	for {
		if err = ctx.Err(); err != nil {
			log.Printf("context canceled; bailing ...")

			break
		}

		const sleepFor = time.Millisecond << 6

		var checkpoint, url string
		switch checkpoint, url, err = cache.Next(ctx); {
		case err != nil:
			log.Fatalf("failed fetching url to purge: %v", err)
		case url == "":
			time.Sleep(sleepFor) // nothing to purge
			continue
		}

		if err = purge.URL(ctx, url); err != nil {
			log.Printf("failed purging %q: %v", url, err)

			time.Sleep(sleepFor)
			continue // retry after sleeping
		}
		log.Printf("purged %q; saving checkpoint ...", url)

		if err = cache.Store(ctx, checkpoint); err != nil {
			log.Fatalf("failed storing checkpoint: %q", err)
		}
	}
}
