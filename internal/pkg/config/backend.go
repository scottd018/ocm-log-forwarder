package config

import (
	"context"
	"errors"
	"fmt"
	"os"

	"k8s.io/client-go/kubernetes"

	"github.com/scottd018/ocm-log-forwarder/internal/pkg/utils"
)

var (
	ErrBackendUnknown             = errors.New("backend type is unknown")
	ErrBackendAuthUnknown         = errors.New("backend auth type is unknown")
	ErrBackendAuthUnkownError     = errors.New("an unknown error occurred retrieving basic auth credentials")
	ErrBackendAuthSecretFormat    = errors.New("invalid secret format")
	ErrBackendAuthMissingUsername = errors.New("unable to find username")
	ErrBackendAuthMissingPassword = errors.New("unable to find password")
)

// NOTE: we are not storing credentials rather pointers to credentials here so
// we do not need to lint this.
//
//nolint:gosec
const (
	// Default Environment Variables.
	DefaultEnvironmentBackend                             = "BACKEND_TYPE"
	defaultEnvironmentBackendElasticSearchURL             = "BACKEND_ES_URL"
	defaultEnvironmentBackendElasticSearchAuthType        = "BACKEND_ES_AUTH_TYPE"
	defaultEnvironmentBackendElasticSearchSecretName      = "BACKEND_ES_SECRET_NAME"
	defaultEnvironmentBackendElasticSearchSecretNamespace = "BACKEND_ES_SECRET_NAMESPACE"
	defaultEnvironmentBackendElasticIndex                 = "BACKEND_ES_INDEX"
	DefaultEnvironmentBackendElasticTLSCertificate        = "BACKEND_ES_CERT"
	DefaultEnvironmentBackendElasticTLSKey                = "BACKEND_ES_KEY"
	DefaultEnvironmentBackendElasticTLSVerify             = "BACKEND_ES_TLS_VERIFY"

	// Default Settings for Environment Variables.
	DefaultBackendElasticSearch                = "elasticsearch"
	DefaultBackend                             = DefaultBackendElasticSearch
	DefaultBackendAuthTypeBasic                = "basic"
	DefaultBackendElasticSearchAuthType        = DefaultBackendAuthTypeBasic
	DefaultBackendElasticIndex                 = "ocm_service_logs"
	defaultBackendElasticSearchURL             = "http://localhost:9200"
	defaultBackendElasticSearchSecretName      = "elastic-auth"
	defaultBackendElasticSearchSecretNamespace = "ocm-log-forwarder"
	DefaultBackendElasticTLSCertificate        = "/etc/pki/tls.crt"
	DefaultBackendElasticTLSKey                = "/etc/pki/tls.key"
	DefaultBackendElasticTLSVerify             = "true"
)

func GetElasticSearchIndex() string {
	return utils.FromEnvironment(defaultEnvironmentBackendElasticIndex, DefaultBackendElasticIndex)
}

func GetElasticSearchURL() string {
	return utils.FromEnvironment(defaultEnvironmentBackendElasticSearchURL, defaultBackendElasticSearchURL)
}

func GetElasticSearchAuthType() string {
	return utils.FromEnvironment(defaultEnvironmentBackendElasticSearchAuthType, DefaultBackendElasticSearchAuthType)
}

func GetElasticSearchTLSConnectionInfo() (tlsCert, tlsKey string, tlsVerify bool) {
	return utils.FromEnvironment(DefaultEnvironmentBackendElasticTLSCertificate, DefaultBackendElasticTLSCertificate),
		utils.FromEnvironment(DefaultEnvironmentBackendElasticTLSKey, DefaultBackendElasticTLSKey),
		utils.BoolFromString(utils.FromEnvironment(DefaultEnvironmentBackendElasticTLSVerify, DefaultBackendElasticTLSVerify))
}

func GetElasticSearchAuthTypeBasic(client *kubernetes.Clientset, ctx context.Context) (username, password string, err error) {
	secretName, secretNamespace := getElasticSearchAuthSecretName(), getElasticSearchAuthSecretNamespace()

	secret, err := utils.GetKubernetesSecret(client, ctx, secretName, secretNamespace)
	if err != nil {
		return "", "", fmt.Errorf("error fetching secret containing elasticsearch auth info - %w", err)
	}

	authData := secret.Data

	// return an error if we do not have exactly one item containing the user/pass
	if len(authData) != 1 {
		return "", "", fmt.Errorf(
			"expect exactly one key value pair in auth data secret; found [%d] - %w",
			len(authData),
			ErrBackendAuthSecretFormat,
		)
	}

	// return the first key/value pair of the auth data as we only should have 1
	for username, password := range authData {
		if username == "" {
			return "", "", fmt.Errorf(
				"error retrieving username from secret [%s] - %w",
				secretName,
				ErrBackendAuthMissingUsername,
			)
		}

		if len(password) == 0 {
			return "", "", fmt.Errorf(
				"error retrieving password from secret [%s] - %w",
				secretName,
				ErrBackendAuthMissingPassword,
			)
		}

		return username, string(password), nil
	}

	// if we made it this far, something really bad happened
	return "", "", ErrBackendAuthUnkownError
}

func getBackendConfig() (string, error) {
	var backend string

	// get the backend
	switch backendType := os.Getenv(DefaultEnvironmentBackend); {
	case backendType == "":
		return DefaultBackend, nil
	case backendType == DefaultBackendElasticSearch:
		return DefaultBackendElasticSearch, nil
	default:
		return backend, fmt.Errorf("backend type [%s] - %w", backendType, ErrBackendUnknown)
	}
}

func getElasticSearchAuthSecretName() string {
	return utils.FromEnvironment(
		defaultEnvironmentBackendElasticSearchSecretName,
		defaultBackendElasticSearchSecretName,
	)
}

func getElasticSearchAuthSecretNamespace() string {
	return utils.FromEnvironment(
		defaultEnvironmentBackendElasticSearchSecretNamespace,
		defaultBackendElasticSearchSecretNamespace,
	)
}
