package config

import (
	"github.com/scottd018/ocm-log-forwarder/internal/pkg/utils"
)

// NOTE: we are not storing credentials rather pointers to credentials here so
// we do not need to lint this.
//
//nolint:gosec
const (
	defaultEnvironmentSecretName      = "OCM_SECRET_NAME"
	defaultEnvironmentSecretNamespace = "OCM_SECRET_NAMESPACE"

	defaultSecretName      = "ocm-token"
	defaultSecretNamespace = "ocm-log-forwarder"
)

func getSecretName() string {
	return utils.FromEnvironment(defaultEnvironmentSecretName, defaultSecretName)
}

func getSecretNamespace() string {
	return utils.FromEnvironment(defaultEnvironmentSecretNamespace, defaultSecretNamespace)
}
