package utils

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func GetKubernetesSecret(client *kubernetes.Clientset, ctx context.Context, secretName, secretNamespace string) (v1.Secret, error) {
	secret, err := client.CoreV1().Secrets(secretNamespace).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		return v1.Secret{}, fmt.Errorf("unable to retrieve secret [%s/%s] from cluster - %w", secretNamespace, secretName, err)
	}

	if secret == nil {
		return v1.Secret{}, fmt.Errorf("unable to retrieve secret [%s/%s] from cluster - %w", secretNamespace, secretName, err)
	}

	return *secret, nil
}
