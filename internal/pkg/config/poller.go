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
	defaultIntervalMinutes       = 5
	minPollIntervalMinutes int64 = 1    // 1 minute minimum
	maxPollIntervalMinutes int64 = 1440 // 1 day maximum
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

	// validate the poller interval is within range
	switch {
	case pollerInterval < minPollIntervalMinutes:
		return 0, fmt.Errorf("poller interval [%v] less than minimum allowed [%v]", pollerInterval, minPollIntervalMinutes)
	case pollerInterval > maxPollIntervalMinutes:
		return 0, fmt.Errorf("poller interval [%v] greater than maximum allowed [%v]", pollerInterval, maxPollIntervalMinutes)
	default:
		return time.Duration(pollerInterval * time.Minute.Nanoseconds()), nil
	}
}
