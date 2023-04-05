package config

import (
	"os"

	"github.com/scottd018/ocm-log-forwarder/internal/pkg/utils"
)

const (
	defaultEnvironmentDebug = "DEBUG"
)

func getDebug() bool {
	return utils.BoolFromString(os.Getenv(defaultEnvironmentDebug))
}
