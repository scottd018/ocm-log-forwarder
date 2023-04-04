package config

import (
	"fmt"
	"os"
)

const (
	defaultEnvironmentClusterId = "OCM_CLUSTER_ID"
)

func getClusterId() (string, error) {
	clusterId := os.Getenv(defaultEnvironmentClusterId)
	if clusterId == "" {
		return "", fmt.Errorf("missing required environment variable [%s]", defaultEnvironmentClusterId)
	}

	return clusterId, nil
}
