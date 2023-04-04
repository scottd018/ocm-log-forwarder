package config

import (
	"fmt"
	"os"
)

const (
	// environment variables
	defaultEnvironmentBackend = "BACKEND_TYPE"

	// default settings
	DefaultBackendElasticSearch = "elasticsearch"
	DefaultBackend              = DefaultBackendElasticSearch
)

func getBackendConfig() (string, error) {
	var backend string

	// get the backend
	switch backendType := os.Getenv(defaultEnvironmentBackend); {
	case backendType == "":
		return DefaultBackend, nil
	case backendType == DefaultBackendElasticSearch:
		return DefaultBackendElasticSearch, nil
	default:
		return backend, fmt.Errorf("unknown backend type [%s]", backendType)
	}
}
