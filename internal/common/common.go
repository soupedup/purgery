// Package common implements functionality consumed by other packages.
package common

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
)
