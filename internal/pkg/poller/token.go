package poller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"
	"time"

	"github.com/scottd018/ocm-log-forwarder/internal/pkg/processor"
	"github.com/scottd018/ocm-log-forwarder/internal/pkg/utils"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	defaultExpiryMinutes int64  = 10
	defaultGrantType     string = "refresh_token"
)

type token struct {
	BearerToken string
	ExpireTime  *time.Time
	Client      *kubernetes.Clientset
	Endpoint    string
}

func NewToken(proc *processor.Processor) (*token, error) {
	proc.Log.InfoF("initializing kubernetes cluster config: cluster=[%s]", proc.Config.ClusterID)
	config, err := rest.InClusterConfig()
	if err == nil {
		// create the clientset for the config
		client, err := kubernetes.NewForConfig(config)
		if err != nil {
			return &token{}, fmt.Errorf("unable to create kubernetes client from in cluster - %w", err)
		}

		return &token{Client: client}, nil
	}

	proc.Log.WarningF("unable to initialize cluster config: cluster=[%s], attempting file initialization", proc.Config.ClusterID)

	kubeConfig := kubeConfigPath()

	proc.Log.InfoF("initializing kubernetes file config: cluster=[%s], file=[%s]", proc.Config.ClusterID, kubeConfig)
	config, err = clientcmd.BuildConfigFromFlags("", kubeConfig)
	if err == nil {
		// create the clientset for the config
		client, err := kubernetes.NewForConfig(config)
		if err != nil {
			return &token{}, fmt.Errorf("unable to create kubernetes client from kubeconfig - %w", err)
		}

		return &token{Client: client}, nil
	}

	return &token{}, fmt.Errorf("unable to create kubernetes client - %w", err)
}

func (token *token) Refresh(proc *processor.Processor) error {
	expireTime := time.Now().Add(time.Duration(defaultExpiryMinutes * time.Minute.Nanoseconds()))

	proc.Log.Infof("retrieving cluster secret: cluster=[%s], secret=[%s]", proc.Config.ClusterID, proc.Config.SecretName)
	secret, err := token.getSecret(proc)
	if err != nil {
		return fmt.Errorf("unable to retrieve kubernetes secret - %w", err)
	}

	proc.Log.Infof("retrieving bearer token: cluster=[%s]", proc.Config.ClusterID)
	if err := token.getBearerToken(proc, secret); err != nil {
		return fmt.Errorf("unable to retrieve bearer token - %w", err)
	}

	token.ExpireTime = &expireTime

	return nil
}

func (token *token) Valid() bool {
	switch {
	case token.BearerToken == "":
		// if the bearer token is not set, our token is considered invalid
		return false
	case token.ExpireTime == nil:
		// if the expire time is not set, our token is considered invalid
		return false
	default:
		return token.ExpireTime.After(time.Now())
	}
}

func (token *token) getSecret(proc *processor.Processor) (v1.Secret, error) {
	secret, err := token.Client.CoreV1().Secrets(proc.Config.SecretNamespace).Get(proc.Context, proc.Config.SecretName, metav1.GetOptions{})
	if err != nil {
		return v1.Secret{}, fmt.Errorf("unable to retrieve secret [%s/%s] from cluster - %w", proc.Config.SecretNamespace, proc.Config.SecretName, err)
	}

	if secret == nil {
		return v1.Secret{}, fmt.Errorf("unable to retrieve secret [%s/%s] from cluster - %w", proc.Config.SecretNamespace, proc.Config.SecretName, err)
	}

	return *secret, nil
}

func (token *token) getBearerToken(proc *processor.Processor, secret v1.Secret) error {
	ocmToken := secret.Data[proc.Config.ClusterID]
	if len(ocmToken) == 0 {
		return fmt.Errorf("missing token data for cluster [%s]", proc.Config.ClusterID)
	}

	// marshal the ocm token into json
	var tokenMap map[string]interface{}
	if err := json.Unmarshal(ocmToken, &tokenMap); err != nil {
		return fmt.Errorf("unable to marshal token data into json map - %w", err)
	}

	// set the endpoint for this token
	token.Endpoint = tokenMap["url"].(string)

	// create payload
	payload := url.Values{}
	payload.Set("grant_type", defaultGrantType)
	payload.Set("refresh_token", tokenMap[defaultGrantType].(string))

	// create the request
	request, err := http.NewRequest("POST", tokenMap["token_url"].(string), bytes.NewBufferString(payload.Encode()))
	if err != nil {
		return fmt.Errorf("unable to create http request - %w", err)
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.SetBasicAuth(tokenMap["client_id"].(string), tokenMap["access_token"].(string))

	// send the request
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("error in http request - %w", err)
	}
	defer response.Body.Close()

	// check the response code
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("error in http response with code - %d - %s", response.StatusCode, response.Status)
	}

	// retrieve the access token from the response
	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("unable to read response body - %w", err)
	}

	var responseMap map[string]interface{}
	if err := json.Unmarshal(responseData, &responseMap); err != nil {
		return fmt.Errorf("unable to response data into json map - %w", err)
	}

	bearerToken := responseMap["access_token"]
	if bearerToken == "" {
		return fmt.Errorf("unable to find bearer token in http response")
	}

	token.BearerToken = bearerToken.(string)

	return nil
}

func homeDir() string {
	return utils.FromEnvironment("HOME", "~")
}

func kubeConfigPath() string {
	return utils.FromEnvironment("KUBECONFIG", filepath.Join(homeDir(), ".kube", "config"))
}
