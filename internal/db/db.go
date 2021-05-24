package db

import "context"

// Dial initializes and returns a connection to the database.
func Dial(_ context.Context) (*DB, error) {
	return &DB{}, nil
}

type DB struct{}

// PrefixesToPurge returns a slice of prefixes which must be purged.
func (db *DB) PrefixesToPurge(ctx context.Context) ([]string, error) {
	return nil, nil
}
