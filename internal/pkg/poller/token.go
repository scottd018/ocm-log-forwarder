package poller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/scottd018/ocm-log-forwarder/internal/pkg/processor"
	"github.com/scottd018/ocm-log-forwarder/internal/pkg/utils"
	v1 "k8s.io/api/core/v1"
)

const (
	defaultExpiryMinutes int64  = 10
	defaultGrantType     string = "refresh_token"
)

type token struct {
	BearerToken string
	ExpireTime  *time.Time
	Endpoint    string
}

func (token *token) Refresh(proc *processor.Processor) error {
	expireTime := time.Now().Add(time.Duration(defaultExpiryMinutes * time.Minute.Nanoseconds()))

	proc.Log.Infof("retrieving cluster secret: cluster=[%s], secret=[%s]", proc.Config.ClusterID, proc.Config.SecretName)
	secret, err := utils.GetKubernetesSecret(proc.KubeClient, proc.Context, proc.Config.SecretName, proc.Config.SecretNamespace)
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
