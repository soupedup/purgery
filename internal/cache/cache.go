// Package cache implements integrations with Redis.
package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/azazeal/exit"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"

	"github.com/soupedup/purgery/internal/common"
	"github.com/soupedup/purgery/internal/env"
	"github.com/soupedup/purgery/internal/log"
)

const (
	keyspace = "purgery:"
	stream   = keyspace + "purge"

	healthy uintptr = 1
)

var errDial = exit.Wrap(common.ECDialCache,
	errors.New("cache: dial"))

// Dial initializes and returns a new Cache client.
func Dial(ctx context.Context, logger *zap.Logger, cfg *env.Config) (*Cache, error) {
	logger.Info("dialing cache ...")

	if err := cfg.Redis.Ping(ctx).Err(); err != nil {
		_ = cfg.Redis.Close()

		logger.Error("failed dialing cache.",
			zap.Error(err))

		return nil, errDial
	}

	logger.Debug("cache dialed.")

	return &Cache{
		client:    cfg.Redis,
		purgeryID: cfg.PurgeryID,
	}, nil
}

// Cache wraps the functionality of our redis client.
type Cache struct {
	client    *redis.Client
	purgeryID string
}

// Ping pings the Redis instance the Cache is configured to connect to.
func (c *Cache) Ping(ctx context.Context, logger *zap.Logger) bool {
	logger.Info("pinging redis ...")

	if err := c.client.Ping(ctx).Err(); err != nil {
		logger.Error("failed pinging redis.",
			zap.Error(err))

		return false
	}

	logger.Debug("pinged.")

	return true
}

// Close implements io.Closer for Cache.
func (c *Cache) Close() error {
	return c.client.Close()
}

var checkpointScript = redis.NewScript(`
	local at = redis.call('TIME')
	local ms = at[2] - (at[2] % 1000)
	ms = (at[1] * 1000) + (ms / 1000)

	local cp = ms .. "-0"

	if not redis.call("SET", KEYS[2], cp, "EX", 86400, "NX") then
		-- key existed; read what's in it and use it as the checkpoint
		cp = redis.call("GET", KEYS[2])
	end

	return cp
`)

func (c *Cache) checkpointKey() string {
	return fmt.Sprintf("%scheckpoints:%s", keyspace, c.purgeryID)
}

func (c *Cache) checkpoint(ctx context.Context, logger *zap.Logger) string {
	keys := []string{
		stream,
		c.checkpointKey(),
	}

	logger.Debug("fetching checkpoint ...")

	cp, err := checkpointScript.Run(ctx, c.client, keys).Text()
	if err != nil {
		logger.Warn("failed loading checkpoint.",
			zap.Error(err))

		return ""
	}

	logger.Debug("checkpoint fetched.",
		log.Checkpoint(cp))

	return cp
}

// Next returns the next url to be purged or empty strings in case such
// a URL does not exist yet.
func (c *Cache) Next(ctx context.Context, logger *zap.Logger) (cp, url string, ok bool) {
	if cp = c.checkpoint(ctx, logger); cp == "" {
		return
	}

	logger.Debug("xreading ...")

	cmd := c.client.XRead(ctx, &redis.XReadArgs{
		Count:   1,
		Block:   time.Second,
		Streams: []string{stream, cp},
	})

	switch err := cmd.Err(); err {
	default:
		logger.Warn("failed xreading.",
			zap.Error(err))
	case redis.Nil:
		ok = true

		logger.Debug("nothing xread.")
	case nil:
		ok = true

		msg := cmd.Val()[0].Messages[0]
		cp = msg.ID
		url = msg.Values["url"].(string)

		logger.Info("xread.",
			log.URL(url),
			log.Checkpoint(cp),
		)
	}

	return
}

// Store saves the given value as the Cache's checkpoint.
func (c *Cache) Store(ctx context.Context, logger *zap.Logger, checkpoint string) bool {
	logger = logger.With(log.Checkpoint(checkpoint))
	logger.Info("storing checkpoint ...")

	if err := c.client.Set(ctx, c.checkpointKey(), checkpoint, time.Minute).Err(); err != nil {

		logger.Error("failed storing checkpoint.", zap.Error(err))

		return false
	}

	logger.Debug("checkpoint stored.")

	return true
}

// Purge enqueues a purge request for the given URL.
func (c *Cache) Purge(ctx context.Context, logger *zap.Logger, url string) (err error) {
	// TODO(azazeal): find client which implements XMIN
	err = c.client.XAdd(ctx, &redis.XAddArgs{
		Stream: stream,
		ID:     "*",
		Values: []string{"url", url},
	}).Err()

	return
}
