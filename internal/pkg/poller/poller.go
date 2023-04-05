package poller

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/scottd018/ocm-log-forwarder/internal/pkg/processor"
)

type Poller struct {
	Token *token
}

func NewPoller(proc *processor.Processor) (*Poller, error) {
	return &Poller{Token: &token{}}, nil
}

func (poller *Poller) Poll(proc *processor.Processor) error {
	if poller.Token == nil {
		return fmt.Errorf("missing token from poller object")
	}

	// refresh the token if it is invalid
	if !poller.Token.Valid() {
		proc.Log.Infof("refreshing token: cluster=[%s]", proc.Config.ClusterID)
		if err := poller.Token.Refresh(proc); err != nil {
			return fmt.Errorf("unable to refresh token - %w", err)
		}
	}

	// call the endpoint
	proc.Log.Infof("retrieving service logs: cluster=[%s]", proc.Config.ClusterID)
	if err := poller.Call(proc); err != nil {
		return fmt.Errorf("unable to retrieve service logs - %w", err)
	}

	return nil
}

func (poller *Poller) Call(proc *processor.Processor) error {
	// create payload
	payload := url.Values{}
	payload.Set("size", "1000")
	payload.Set("search", fmt.Sprintf("cluster_id = '%s'", proc.Config.ClusterID))

	// create the url object with the base url and set the params
	requestURL, err := url.Parse(serviceLogPath(poller.Token))
	if err != nil {
		return fmt.Errorf("unable to create base url object - %w", err)
	}
	requestURL.RawQuery = payload.Encode()

	// create the request
	request, err := http.NewRequest("GET", requestURL.String(), nil)
	if err != nil {
		return fmt.Errorf("unable to create http request - %w", err)
	}
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", poller.Token.BearerToken))

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

	proc.ResponseData = responseData

	return nil
}

func serviceLogPath(token *token) string {
	return fmt.Sprintf("%s/api/service_logs/v1/cluster_logs", token.Endpoint)
}
