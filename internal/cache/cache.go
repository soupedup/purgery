// Package cache implements integrations with Redis.
package cache

import (
	"errors"
	"fmt"

	"github.com/azazeal/exit"
	"github.com/gomodule/redigo/redis"
	"go.uber.org/zap"

	"github.com/soupedup/purgery/internal/common"
	"github.com/soupedup/purgery/internal/log"
)

const (
	keyspace = common.AppName + ":"
	stream   = keyspace + "purge"
)

var errDial = exit.Wrap(common.ECDialCache,
	errors.New("cache: dial"))

// New initializes and returns a new Cache, for the given ID, which works on the
// given Conn.
func New(purgeryID string, pool *redis.Pool) *Cache {
	return &Cache{
		redis:     pool,
		purgeryID: purgeryID,
	}
}

// Cache wraps the functionality of our redis client.
type Cache struct {
	redis     *redis.Pool
	purgeryID string
}

// Ping pings the Redis instance the Cache is configured to connect to.
func (c *Cache) Ping(logger *zap.Logger) bool {
	conn := c.redis.Get()
	defer conn.Close()

	logger.Info("pinging redis ...")

	switch got, err := redis.String(conn.Do("PING", common.AppName)); {
	case err != nil:
		logger.Error("failed pinging redis.",
			zap.Error(err))

		return false
	case common.AppName != got:
		logger.Warn("read unexpected ping response.",
			zap.String("exp", common.AppName),
			zap.String("got", got))

		return false
	default:
		logger.Debug("pinged.")

		return true
	}
}

// Close closes the underlying Cache connection pool.
func (c *Cache) Close() error {
	return c.redis.Close()
}

var checkpointScript = redis.NewScript(2, `
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

func (c *Cache) checkpoint(logger *zap.Logger, conn redis.Conn) string {
	logger.Debug("fetching checkpoint ...")

	cp, err := redis.String(checkpointScript.Do(conn, stream, c.checkpointKey()))
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
func (c *Cache) Next(logger *zap.Logger) (cp, url string, ok bool) {
	conn := c.redis.Get()
	defer conn.Close()

	if cp = c.checkpoint(logger, conn); cp == "" {
		return
	}

	logger.Debug("xreading ...")

	ret, err := redis.Values(conn.Do("XREAD",
		"COUNT", 1,
		"BLOCK", 1000,
		"STREAMS", stream,
		cp,
	))

	switch err {
	default:
		logger.Warn("failed xreading.",
			zap.Error(err))
	case redis.ErrNil:
		ok = true

		logger.Debug("nothing xread.")
	case nil:
		ok = true

		msg := ret[0].([]interface{})[1].([]interface{})[0].([]interface{})
		cp = string(msg[0].([]byte))

		pairs := msg[1].([]interface{})
		url = string(pairs[1].([]byte))

		logger.Info("xread.",
			log.URL(url),
			log.Checkpoint(cp),
		)
	}

	return
}

// Store saves the given value as the Cache's checkpoint.
func (c *Cache) Store(logger *zap.Logger, checkpoint string) bool {
	conn := c.redis.Get()
	defer conn.Close()

	logger = logger.With(log.Checkpoint(checkpoint))
	logger.Info("storing checkpoint ...")

	res, err := redis.String(conn.Do("SET",
		c.checkpointKey(), checkpoint,
		"EX", 60,
	))

	switch {
	case err != nil:
		logger.Error("failed storing checkpoint.",
			zap.Error(err))

		return false
	case res != "OK":
		logger.Warn("wrong redis SET response.",
			zap.String("response", res))

		return false
	default:
		logger.Debug("checkpoint stored.")

		return true
	}
}

// EnqueuePurgeRequest enqueues a purge request for the given URL.
func (c *Cache) EnqueuePurgeRequest(logger *zap.Logger, url string) bool {
	conn := c.redis.Get()
	defer conn.Close()

	logger = logger.With(log.URL(url))
	logger.Info("enqueueing purge request ...")

	id, err := redis.String(conn.Do("XADD",
		stream,
		"MINID", "~", "0-0",
		"*",
		"url", url,
	))

	if err != nil {
		logger.Error("failed enqueueing purge request.",
			zap.Error(err))

		return false
	}

	logger.Debug("enqueued purge request.", zap.String("id", id))

	return true
}
