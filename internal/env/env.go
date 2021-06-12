// Package env implements environment-related functionality.
package env

import (
	"errors"
	"os"
	"strings"

	"github.com/azazeal/exit"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"

	"github.com/soupedup/purgery/internal/common"
)

// Config wraps
type Config struct {
	// PurgeryID holds the value of the PURGERY_ID environment value.
	PurgeryID string

	// Redis holds a reference to the Redis client.
	Redis *redis.Client

	// VarnishAddr holds the value of the VARNISH_ADDR environment value.
	VarnishAddr string
}

func (cfg *Config) parseRedisURL(logger *zap.Logger, url string) bool {
	opt, err := redis.ParseURL(url)
	if err != nil {
		logger.Error("failed parsing redis url.",
			zap.Error(err))

		return false
	}

	cfg.Redis = redis.NewClient(opt)

	return true
}

var errLoadConfig = exit.Wrap(common.ECLoadConfig,
	errors.New("env: failed loading configuration"))

func LoadConfig(logger *zap.Logger) (*Config, error) {
	logger.Info("loading configuration from the environment ...")

	var cfg Config
	l1 := fetch(logger, &cfg.PurgeryID, "PURGERY_ID")

	var redisURL string
	l2 := fetch(logger, &redisURL, "REDIS_URL") &&
		cfg.parseRedisURL(logger, redisURL)
	l3 := fetch(logger, &cfg.VarnishAddr, "VARNISH_ADDR")

	if l1 && l2 && l3 {
		return &cfg, nil
	}

	return nil, errLoadConfig
}

func fetch(logger *zap.Logger, into *string, key string) (ok bool) {
	*into = strings.TrimSpace(os.Getenv(key))

	if ok = *into != ""; !ok {
		logger.Error("a required environment variable is undefined or empty.",
			zap.String("var", key))
	}

	return
}
