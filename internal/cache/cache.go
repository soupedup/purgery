// Package cache implements integrations with Redis.
package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/go-redis/redis/v8"
)

const ns = "clvr"

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
			nextKeys: []string{
				fmt.Sprintf("%s:purge", ns),
				fmt.Sprintf("%s:checkpoints:%s", ns, purgerID),
			},
		}
	}

	return
}

// Cache wraps the functionality of our redis client.
type Cache struct {
	client   *redis.Client
	purgerID string
	nextKeys []string
}

// Close implements io.Closer for Cache.
func (c *Cache) Close() error {
	return c.client.Close()
}

var nextScript = redis.NewScript(`
	local cp = "0-0"

	if redis.call("SET", KEYS[2], cp, "EX", 60, "NX") ~= true then
		-- key existed; read what's in it and use it as the checkpoint
		cp = redis.call("GET", KEYS[2])
	end

	-- NOTE: we cannot block from a lua script
	local ret = redis.call("XREAD", "COUNT", 1, "STREAMS", KEYS[1], cp)
	if ret ~= false then
		ret = ret[1][2][1]

		local obj = { id = ret[1], url = ret[2][2] }
		ret = cjson.encode(obj)
	end
	return ret
`)

// Next returns the next URL that must be purged.
func (c *Cache) Next(ctx context.Context) (cp, url string, err error) {
	var res string
	switch res, err = nextScript.Run(ctx, c.client, c.nextKeys).Text(); err {
	case nil:
		break
	case redis.Nil:
		err = nil

		return
	default:
		return
	}

	log.Print(res)

	var tuple struct {
		ID  string `json:"id"`
		URL string `json:"url"`
	}

	dec := json.NewDecoder(strings.NewReader(res))
	if err = dec.Decode(&tuple); err == nil {
		cp = tuple.ID
		url = tuple.URL
	}

	return
}
