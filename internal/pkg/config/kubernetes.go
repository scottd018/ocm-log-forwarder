package config

import (
	"github.com/scottd018/ocm-log-forwarder/internal/pkg/utils"
)

// NOTE: we are not storing credentials rather pointers to credentials here so
// we do not need to lint this.
//
//nolint:gosec
const (
	defaultEnvironmentSecretFile      = "OCM_SECRET_FILE"
	defaultEnvironmentSecretName      = "OCM_SECRET_NAME"
	defaultEnvironmentSecretNamespace = "OCM_SECRET_NAMESPACE"

	defaultSecretFile      = "~/.ocm.json"
	defaultSecretName      = "ocm-token"
	defaultSecretNamespace = "ocm-log-forwarder"
)

func getSecretFile() string {
	return utils.FromEnvironment(defaultEnvironmentSecretFile, defaultSecretFile)
}

func getSecretName() string {
	return utils.FromEnvironment(defaultEnvironmentSecretName, defaultSecretName)
}

func getSecretNamespace() string {
	return utils.FromEnvironment(defaultEnvironmentSecretNamespace, defaultSecretNamespace)
}
