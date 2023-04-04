package config

import (
	"github.com/scottd018/ocm-log-forwarder/internal/pkg/utils"
)

const (
	defaultEnvironmentSecretName      = "OCM_SECRET_NAME"
	defaultEnvironmentSecretNamespace = "OCM_SECRET_NAMESPACE"

	defaultSecretName      = "ocm-token"
	defaultSecretNamespace = "elastic-system"
)

func getSecretName() string {
	return utils.FromEnvironment(defaultEnvironmentSecretName, defaultSecretName)
}

func getSecretNamespace() string {
	return utils.FromEnvironment(defaultEnvironmentSecretNamespace, defaultSecretNamespace)
}
