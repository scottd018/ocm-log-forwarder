package poller

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/scottd018/ocm-log-forwarder/internal/pkg/processor"
	"github.com/scottd018/ocm-log-forwarder/internal/pkg/utils"
)

var (
	ErrTokenInvalid    = errors.New("invalid token")
	ErrResponseInvalid = errors.New("invalid http response")
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

type TokenData struct {
	RefreshToken string `json:"refresh_token"`
	URL          string `json:"url"`
	TokenURL     string `json:"token_url"`
	ClientID     string `json:"client_id"`
	AccessToken  string `json:"access_token"`
}

type TokenRefreshData struct {
	BearerToken string `json:"access_token"`
}

func (token *token) Refresh(proc *processor.Processor) error {
	expireTime := time.Now().Add(time.Duration(defaultExpiryMinutes * time.Minute.Nanoseconds()))

	// retrieve the kubernetes secret containing the ocm token from the cluster
	proc.Log.Infof("retrieving cluster secret: cluster=[%s], secret=[%s]", proc.Config.ClusterID, proc.Config.SecretName)
	secret, err := utils.GetKubernetesSecret(proc.KubeClient, proc.Context, proc.Config.SecretName, proc.Config.SecretNamespace)
	if err != nil {
		return fmt.Errorf("unable to retrieve kubernetes secret - %w", err)
	}

	// retrieve the token data from the secret
	tokenData, err := getTokenData(secret.Data[proc.Config.ClusterID])
	if err != nil {
		return fmt.Errorf("unable to retrieve token data from secret [%s] - %w", secret.GetName(), err)
	}

	// refresh the token
	proc.Log.Infof("retrieving bearer token: cluster=[%s]", proc.Config.ClusterID)
	if err := token.getBearerToken(proc, tokenData); err != nil {
		return fmt.Errorf("unable to retrieve bearer token - %w", err)
	}

	// reset the expire time for the refresh token
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

func (token *token) getBearerToken(proc *processor.Processor, tokenData TokenData) error {
	// set the endpoint for this token
	token.Endpoint = tokenData.URL

	// create payload
	payload := url.Values{}
	payload.Set("grant_type", defaultGrantType)
	payload.Set("refresh_token", tokenData.RefreshToken)

	// create the request
	request, err := http.NewRequest("POST", tokenData.TokenURL, bytes.NewBufferString(payload.Encode()))
	if err != nil {
		return fmt.Errorf("unable to create http request - %w", err)
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.SetBasicAuth(tokenData.ClientID, tokenData.AccessToken)

	// send the request
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("error in http request - %w", err)
	}
	defer response.Body.Close()

	// check the response code
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf(
			"invalid response code [%d]; message [%s] - %w",
			response.StatusCode,
			response.Status,
			ErrResponseInvalid,
		)
	}

	// retrieve the access token from the response
	responseData, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("unable to read response body - %w", err)
	}

	tokenRefreshData := &TokenRefreshData{}
	if err := json.Unmarshal(responseData, tokenRefreshData); err != nil {
		return fmt.Errorf("unable to read response data - %w", err)
	}

	if tokenRefreshData.BearerToken == "" {
		return fmt.Errorf("unable to find bearer token in http response - %w", ErrResponseInvalid)
	}

	token.BearerToken = tokenRefreshData.BearerToken

	return nil
}

func getTokenData(tokenBytes []byte) (TokenData, error) {
	tokenData := TokenData{}

	if len(tokenBytes) == 0 {
		return tokenData, fmt.Errorf("missing token data - %w", ErrTokenInvalid)
	}

	// serialize token as json
	if err := json.Unmarshal(tokenBytes, &tokenData); err != nil {
		return tokenData, fmt.Errorf("unable to serialize token - %w", ErrTokenInvalid)
	}

	return tokenData, nil
}
