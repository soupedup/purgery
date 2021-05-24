package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/jsierles/clvr/internal/db"
	"github.com/jsierles/clvr/internal/purge"
)

func init() {
	log.SetOutput(os.Stderr)
	log.SetPrefix("clvr ")
	log.SetFlags(log.LUTC | log.Ldate | log.Ltime)
}

// there's a set of caching servers (we'll call these nodes)
// there's a *source of truth* which contains the latest timestamp each
// domain has been purge.

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	db := dialDB(ctx)

	for {
		if err := ctx.Err(); err != nil {
			log.Print("context canceled; bailing ...")

			break
		}

		if !cycle(ctx, db) {
			break
		}
	}

	log.Printf("exiting ...")
}

func cycle(ctx context.Context, db *db.DB) bool {
	prefixes, err := db.PrefixesToPurge(ctx)
	if err != nil {
		log.Printf("failed getting prefixes: %v", err)

		return false
	}

	var wg sync.WaitGroup
	defer wg.Wait()

	for _, prefix := range prefixes {
		wg.Add(1)

		go func(prefix string) {
			defer wg.Done()

			purge.Node(ctx, prefix, time.Now())
		}(prefix)
	}

	return true
}

func dialDB(ctx context.Context) (database *db.DB) {
	log.Print("dialing db ...")

	var err error
	if database, err = db.Dial(ctx); err != nil {
		log.Fatalf("failed dialing db: %v", err)
	}
	log.Print("dialed db.")

	return
}
