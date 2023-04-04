package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

const (
	// environment variables
	defaultEnvironmentIntervalMinutes = "OCM_POLL_INTERVAL_MINUTES"

	// default settings
	defaultIntervalMinutes = 5
)

func getPollerInterval() (time.Duration, error) {
	pollerIntervalMinutes := os.Getenv(defaultEnvironmentIntervalMinutes)
	if pollerIntervalMinutes == "" {
		return defaultIntervalMinutes, nil
	}

	pollerInterval, err := strconv.ParseInt(pollerIntervalMinutes, 10, 64)
	if err != nil {
		return 0, fmt.Errorf(
			"unable to convert environment variable [%s=%s] to int64 value - %w",
			defaultEnvironmentIntervalMinutes,
			pollerIntervalMinutes,
			err,
		)
	}

	return time.Duration(pollerInterval * time.Minute.Nanoseconds()), nil
}
