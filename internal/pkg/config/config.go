package config

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrMissingEnvironmentVariable = errors.New("unable to find required environment variable")
)

type Config struct {
	ClusterID       string
	Backend         string
	PollerInterval  time.Duration
	SecretName      string
	SecretNamespace string
	SecretFile      string
	TokenFile       string

	Debug bool
}

func Initialize() (*Config, error) {
	// get the cluster id
	clusterName, err := getClusterID()
	if err != nil {
		return &Config{}, fmt.Errorf("unable to get cluster id - %w", err)
	}

	// get the backend
	backend, err := getBackendConfig()
	if err != nil {
		return &Config{}, fmt.Errorf("unable to get backend config - %w", err)
	}

	// get the polling interval
	interval, err := getPollerInterval()
	if err != nil {
		return &Config{}, fmt.Errorf("unable to get poller interval config - %w", err)
	}

	return &Config{
		ClusterID:       clusterName,
		Backend:         backend,
		PollerInterval:  interval,
		SecretName:      getSecretName(),
		SecretNamespace: getSecretNamespace(),
		SecretFile:      getTokenFile(),
		TokenFile:       getTokenFile(),
		Debug:           getDebug(),
	}, nil
}
