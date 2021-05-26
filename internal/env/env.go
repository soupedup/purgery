// Package env implements environment-related functionality.
package env

import (
	"fmt"
	"os"
	"strings"
)

// RedisURL reports the value of the $REDIS_URL environment variable or an
// error in case it's either undefined or empty.
func RedisURL() (string, error) {
	return fetch("REDIS_URL")
}

// AllocID reports the value of the $FLY_ALLOC_ID environment variable or an
// error in case it's either undefined or empty.
func AllocID() (string, error) {
	return fetch("FLY_ALLOC_ID")
}

// PurgeAddr reports the value of the $PURGE_ADDR environment variable or an
// error in case it's either undefined or empty.
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
