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

// PurgerID reports the value of the $PURGER_ID environment variable or an
// error in case it's either undefined or empty.
func PurgerID() (string, error) {
	return fetch("PURGER_ID")
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
