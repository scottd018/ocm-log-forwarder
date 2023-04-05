package utils

import (
	"os"
	"strings"
)

func FromEnvironment(variable string, def string) string {
	varFromEnv := os.Getenv(variable)
	if varFromEnv != "" {
		return varFromEnv
	}

	return def
}

func BoolFromString(str string) bool {
	lower := strings.ToLower(str)

	return map[string]bool{
		"yes":  true,
		"y":    true,
		"true": true,
		"1":    true,
		"t":    true,
	}[lower]
}
