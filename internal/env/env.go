// Package env implements environment-related functionality.
package env

import (
	"os"
	"strings"

	"go.uber.org/zap"
)

// Config wraps
type Config struct {
	// ID holds the value of the PURGERY_ID environment value.
	PurgeryID string

	// Redis holds the value of the REDIS_URL environment variable.
	RedisURL string

	// VarnishAddr holds the value of the VARNISH_ADDR environment value.
	VarnishAddr string
}

func LoadConfig(logger *zap.Logger) *Config {
	logger.Info("loading configuration from the environment ...")

	var cfg Config
	l1 := fetch(logger, &cfg.PurgeryID, "PURGERY_ID")
	l2 := fetch(logger, &cfg.RedisURL, "REDIS_URL")
	l3 := fetch(logger, &cfg.VarnishAddr, "VARNISH_ADDR")

	if l1 && l2 && l3 {
		return &cfg
	}

	return nil
}

func fetch(logger *zap.Logger, into *string, key string) (ok bool) {
	*into = strings.TrimSpace(os.Getenv(key))

	if ok = *into != ""; !ok {
		logger.Error("a required environment variable is undefined or empty.",
			zap.String("var", key))
	}

	return
}
