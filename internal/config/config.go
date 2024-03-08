package config

import "os"

// DBConfig is config with sensitive data, needed for working with db.
type DBConfig struct {
	Host         string
	MigrationURL string
}

// NewDBConfig returns DBConfig with sensitive data, needed for working with db.
func NewDBConfig() *DBConfig {
	return &DBConfig{
		Host:         os.Getenv("SQLITE_HOST"),
		MigrationURL: os.Getenv("MIGRATION_URL"),
	}
}

// HTTPConfig is config with sensitive data, needed for rest API.
type HTTPConfig struct {
	Host string
}

// NewHTTPConfig returns HTTPConfig with sensitive data, needed for rest API.
func NewHTTPConfig() *HTTPConfig {
	return &HTTPConfig{
		Host: os.Getenv("HTTP_HOST"),
	}
}
