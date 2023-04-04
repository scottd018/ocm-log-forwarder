package utils

import "os"

func FromEnvironment(variable string, def string) string {
	varFromEnv := os.Getenv(variable)
	if varFromEnv != "" {
		return varFromEnv
	}

	return def
}
