package config

import (
	"fmt"
	"os"
)

const (
	defaultEnvironmentClusterID = "OCM_CLUSTER_ID"
)

func getClusterID() (string, error) {
	clusterID := os.Getenv(defaultEnvironmentClusterID)
	if clusterID == "" {
		return "", fmt.Errorf("missing [%s] - %w", defaultEnvironmentClusterID, ErrMissingEnvironmentVariable)
	}

	return clusterID, nil
}
