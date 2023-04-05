package config

import (
	"context"
	"fmt"
	"os"

	"github.com/scottd018/ocm-log-forwarder/internal/pkg/utils"
	"k8s.io/client-go/kubernetes"
)

const (
	// environment variables
	DefaultEnvironmentBackend                             = "BACKEND_TYPE"
	defaultEnvironmentBackendElasticSearchURL             = "BACKEND_ES_URL"
	defaultEnvironmentBackendElasticSearchAuthType        = "BACKEND_ES_AUTH_TYPE"
	defaultEnvironmentBackendElasticSearchSecretName      = "BACKEND_ES_SECRET_NAME"
	defaultEnvironmentBackendElasticSearchSecretNamespace = "BACKEND_ES_SECRET_NAMESPACE"
	defaultEnvironmentBackendElasticIndex                 = "BACKEND_ES_INDEX"

	// default settings
	DefaultBackendElasticSearch                = "elasticsearch"
	DefaultBackend                             = DefaultBackendElasticSearch
	DefaultBackendAuthTypeBasic                = "basic"
	DefaultBackendElasticSearchAuthType        = DefaultBackendAuthTypeBasic
	DefaultBackendElasticIndex                 = "ocm_service_logs"
	defaultBackendElasticSearchURL             = "http://localhost:9200"
	defaultBackendElasticSearchSecretName      = "elastic-auth"
	defaultBackendElasticSearchSecretNamespace = "elastic-system"
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

func GetElasticSearchAuthTypeBasic(client *kubernetes.Clientset, ctx context.Context) (string, string, error) {
	secretName, secretNamespace := getElasticSearchAuthSecretName(), getElasticSearchAuthSecretNamespace()

	secret, err := utils.GetKubernetesSecret(client, ctx, secretName, secretNamespace)
	if err != nil {
		return "", "", fmt.Errorf("error fetching secret containing elasticsearch auth info - %w", err)
	}
	authData := secret.Data

	// return an error if we do not have exactly one item containing the user/pass
	if len(authData) != 1 {
		return "", "", fmt.Errorf("expect exactly one key value pair in auth data secret; found [%d]", len(authData))
	}

	// return the first key/value pair of the auth data as we only should have 1
	for username, password := range authData {
		if username == "" {
			return "", "", fmt.Errorf("username must not be empty")
		}

		if string(password) == "" {
			return "", "", fmt.Errorf("password must not be empty")
		}

		return username, string(password), nil
	}

	// if we made it this far, something really bad happened
	return "", "", fmt.Errorf("unknown error occurred when retrieving basic auth info for elasticsearch")
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
		return backend, fmt.Errorf("unknown backend type [%s]", backendType)
	}
}

func getElasticSearchAuthSecretName() string {
	return utils.FromEnvironment(defaultEnvironmentBackendElasticSearchSecretName, defaultBackendElasticSearchSecretName)
}

func getElasticSearchAuthSecretNamespace() string {
	return utils.FromEnvironment(defaultEnvironmentBackendElasticSearchSecretNamespace, defaultBackendElasticSearchSecretNamespace)
}
