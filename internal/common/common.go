// Package common implements functionality consumed by other packages.
package common

import "net/url"

// AppName denotes the app's name.
const AppName = "purgery"

// The set of exit codes this application uses.
const (
	_ = iota + 2

	// ECLoadConfig is returned when the application is unable to load its
	// configuration from the environment or its configuration is invalid.
	ECLoadConfig

	// ECDialCache is returned when the application is unable to dial its cache.
	ECDialCache

	// ECBind is returned when the application's embedded REST server fails to
	// bind on the configured address.
	ECBind
)

// IsValidURL reports whether the given URL is a valid one.
func IsValidURL(rawurl string) bool {
	if rawurl == "" {
		return false
	}

	url, err := url.Parse(rawurl)
	return err == nil &&
		url.Scheme == "http"
}
