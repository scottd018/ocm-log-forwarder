package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/scottd018/ocm-log-forwarder/internal/pkg/utils"
)

var (
	ErrPollerIntervalRange = errors.New("poller interval out of range")
)

const (
	// Default Environment Variables.
	defaultEnvironmentIntervalMinutes = "OCM_POLL_INTERVAL_MINUTES"
	defaultEnvironmentTokenFile       = "OCM_TOKEN_FILE"

	// Default Settings for Environment Variables.
	defaultIntervalMinutes              = 5
	defaultTokenFile                    = "/tmp/ocm.json"
	defaultMinPollIntervalMinutes int64 = 1    // 1 minute minimum
	defaultMaxPollIntervalMinutes int64 = 1440 // 1 day maximum
)

func getPollerInterval() (time.Duration, error) {
	pollerIntervalMinutes := os.Getenv(defaultEnvironmentIntervalMinutes)
	if pollerIntervalMinutes == "" {
		return (defaultIntervalMinutes * time.Minute), nil
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
	case pollerInterval < defaultMinPollIntervalMinutes:
		return 0, fmt.Errorf(
			"poller interval [%v] less than minimum allowed [%v] - %w",
			pollerInterval,
			defaultMinPollIntervalMinutes,
			ErrPollerIntervalRange,
		)
	case pollerInterval > defaultMaxPollIntervalMinutes:
		return 0, fmt.Errorf(
			"poller interval [%v] greater than maximum allowed [%v] - %w",
			pollerInterval,
			defaultMaxPollIntervalMinutes,
			ErrPollerIntervalRange,
		)
	default:
		return time.Duration(pollerInterval * time.Minute.Nanoseconds()), nil
	}
}

func getTokenFile() string {
	return utils.FromEnvironment(defaultEnvironmentTokenFile, defaultTokenFile)
}
