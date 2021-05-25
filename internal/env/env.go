// Package env implements environment-related functionality.
package env

import (
	"fmt"
	"os"
	"strings"
)

func RedisURL() (string, error) {
	return fetch("REDIS_URL")
}

func AllocID() (string, error) {
	return fetch("FLY_ALLOC_ID")
}

func PurgeAddr() (string, error) {
	return fetch("PURGE_ADDR")
}

type errUndefinedOrBlank string

func (err errUndefinedOrBlank) Error() string {
	return fmt.Sprintf("env: $%s is undefined or empty", string(err))
}

func fetch(key string) (val string, err error) {
	if val = strings.TrimSpace(os.Getenv(key)); val == "" {
		err = errUndefinedOrBlank(key)
	}

	return
}
