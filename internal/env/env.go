// Package env implements environment-related functionality.
package env

import (
	"errors"
	"os"
	"strings"
	"time"

	"github.com/azazeal/exit"
	"github.com/gomodule/redigo/redis"
	"go.uber.org/zap"

	"github.com/soupedup/purgery/internal/common"
	"github.com/soupedup/purgery/internal/safe"
)

// Config wraps
type Config struct {
	// Addr holds the value of the ADDR environment variable.
	Addr string

	// APIKey holds the value of the API_KEY environment variable.
	APIKey string

	// PurgeryID holds the value of the PURGERY_ID environment value.
	PurgeryID string

	// Redis holds a reference to the Redis connection pool.
	Redis *redis.Pool

	// VarnishAddr holds the value of the VARNISH_ADDR environment value.
	VarnishAddr string
}

var redisDialOpts = []redis.DialOption{
	redis.DialConnectTimeout(5 * time.Second),
	redis.DialReadTimeout(3 * time.Second),
	redis.DialWriteTimeout(3 * time.Second),
}

func (cfg *Config) dialRedis(logger *zap.Logger, url string) bool {
	logger.Info("dialing redis ...")

	conn, err := redis.DialURL(url, redisDialOpts...)
	if err != nil {
		logger.Error("failed dialing redis.",
			zap.Error(err))

		return false
	}
	_ = conn.Close()

	cfg.Redis = &redis.Pool{
		MaxIdle:     5,
		IdleTimeout: 10 * time.Minute,
		Wait:        true,
		Dial: func() (redis.Conn, error) {
			return redis.DialURL(url, redisDialOpts...)
		},
	}

	logger.Debug("redis dialed.")

	return true
}

func (cfg *Config) setAPIKey(logger *zap.Logger, key string) bool {
	// we have to make sure that the length of the key doesn't exceed
	if l := len(key); l > safe.MaxCompareLen {
		logger.Error("the API key is too long.",
			zap.Int("length", l),
			zap.Int("max", safe.MaxCompareLen))

		return false
	}

	cfg.APIKey = key

	return true
}

var errLoadConfig = exit.Wrap(common.ECLoadConfig,
	errors.New("env: failed loading configuration"))

func LoadConfig(logger *zap.Logger) (*Config, error) {
	logger.Info("loading configuration from the environment ...")

	var (
		cfg      Config
		redisURL string
		apiKey   string
	)

	ok := []bool{
		fetch(logger, &cfg.Addr, "ADDR"),

		fetch(logger, &apiKey, "API_KEY") &&
			cfg.setAPIKey(logger, apiKey),

		fetch(logger, &cfg.PurgeryID, "PURGERY_ID"),

		fetch(logger, &redisURL, "REDIS_URL") &&
			cfg.dialRedis(logger, redisURL),

		fetch(logger, &cfg.VarnishAddr, "VARNISH_ADDR"),
	}

	for _, ok := range ok {
		if !ok {
			return nil, errLoadConfig
		}
	}

	return &cfg, nil
}

func fetch(logger *zap.Logger, into *string, key string) (ok bool) {
	*into = strings.TrimSpace(os.Getenv(key))

	if ok = *into != ""; !ok {
		logger.Error("a required environment variable is undefined or empty.",
			zap.String("var", key))
	}

	return
}
