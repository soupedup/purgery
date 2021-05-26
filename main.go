package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/soupedup/purgery/internal/cache"
	"github.com/soupedup/purgery/internal/env"
	"github.com/soupedup/purgery/internal/purge"
)

func init() {
	log.SetOutput(os.Stderr)
	log.SetPrefix("purgery ")
	log.SetFlags(log.LUTC | log.Ldate | log.Ltime | log.Lmicroseconds | log.Lmsgprefix)
}

func main() {
	addr, err := env.PurgeAddr()
	if err != nil {
		log.Fatalf("failed fetching purgeAddr: %v", err)
	}

	log.Print("")
	log.Print("")
	log.Print("")
	log.Printf("PURGE_ADDR: %q", addr)
	log.Print("")
	log.Print("")
	log.Print("")

	allocID, err := env.PurgerID()
	if err != nil {
		log.Fatalf("failed fetching purgerID: %v", err)
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
		const errorSleep = time.Millisecond << 6
		if err != nil {
			time.Sleep(errorSleep)
		}
		if err = ctx.Err(); err != nil {
			log.Printf("context canceled; bailing ...")

			break
		}

		var checkpoint, url string
		switch checkpoint, url, err = cache.Next(ctx); {
		case err != nil:
			log.Printf("failed fetching url to purge: %v", err)

			continue
		case url == "":
			continue
		}

		if err = purge.URL(ctx, url); err != nil {
			log.Printf("failed purging %q: %v", url, err)

			continue
		}
		log.Printf("purged %q; saving checkpoint ...", url)

		if err = cache.Store(ctx, checkpoint); err != nil {
			log.Printf("failed storing checkpoint: %q", err)
		}
	}
}
