// Package cache implements integrations with Redis.
package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

const (
	keyspace = "purgery:"
	stream   = keyspace + "purge"
)

func Dial(ctx context.Context, purgerID, url string) (c *Cache, err error) {
	var opts *redis.Options
	if opts, err = redis.ParseURL(url); err != nil {
		return
	}
	client := redis.NewClient(opts)

	if err = client.Ping(ctx).Err(); err == nil {
		c = &Cache{
			client:   client,
			purgerID: purgerID,
		}
	}

	return
}

// Cache wraps the functionality of our redis client.
type Cache struct {
	client   *redis.Client
	purgerID string
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
	return fmt.Sprintf("%scheckpoints:%s", keyspace, c.purgerID)
}

func (c *Cache) checkpoint(ctx context.Context) (string, error) {
	keys := []string{
		stream,
		c.checkpointKey(),
	}

	return checkpointScript.Run(ctx, c.client, keys).Text()
}

// Next returns the next url to be purge or empty strings in case such
// a URL does not exist yet.
func (c *Cache) Next(ctx context.Context) (cp, url string, err error) {
	if cp, err = c.checkpoint(ctx); err != nil {
		return
	}

	cmd := c.client.XRead(ctx, &redis.XReadArgs{
		Count:   1,
		Block:   time.Second,
		Streams: []string{stream, cp},
	})

	switch cmd.Err() {
	case redis.Nil:
		cp = ""
		err = nil
	case nil:
		msg := cmd.Val()[0].Messages[0]

		cp = msg.ID
		url = msg.Values["url"].(string)
	}

	return
}

// Store saves the given value as the Cache's checkpoint.
func (c *Cache) Store(ctx context.Context, checkpoint string) error {
	return c.client.Set(ctx, c.checkpointKey(), checkpoint, time.Minute).Err()
}
