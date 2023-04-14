package poller

import (
	"encoding/json"
	"fmt"
	"io"
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

func (poller *Poller) Poll(proc *processor.Processor) (responseData Response, err error) {
	if poller.Token == nil {
		return responseData, fmt.Errorf("missing token from poller object - %w", ErrTokenInvalid)
	}

	// refresh the token if it is invalid
	if !poller.Token.Valid() {
		proc.Log.Infof("refreshing token: cluster=[%s]", proc.Config.ClusterID)

		if err = poller.Token.Refresh(proc); err != nil {
			return responseData, fmt.Errorf("unable to refresh token - %w", err)
		}
	}

	// call the endpoint
	proc.Log.Infof("retrieving service logs: cluster=[%s]", proc.Config.ClusterID)
	if responseData, err = poller.Request(proc); err != nil {
		return responseData, fmt.Errorf("unable to retrieve service logs - %w", err)
	}

	return responseData, nil
}

func (poller *Poller) Request(proc *processor.Processor) (responseData Response, err error) {
	page := 1

	// loop through each of the pages and generate a response that stores
	// all of the messages.
	for {
		response, err := poller.Call(proc, page)
		if err != nil {
			return responseData, err
		}

		// append the items to the response
		responseData.Messages = append(responseData.Messages, response.Messages...)

		if response.PageCount() == page {
			break
		}

		page++
	}

	return responseData, nil
}

func (poller *Poller) Call(proc *processor.Processor, pageNum int) (responseData Response, err error) {
	// create payload
	payload := url.Values{}
	payload.Set("size", "1000")
	payload.Set("page", fmt.Sprintf("%d", pageNum))
	payload.Set("search", fmt.Sprintf("cluster_id = '%s'", proc.Config.ClusterID))

	// create the url object with the base url and set the params
	requestURL, err := url.Parse(serviceLogPath(poller.Token))
	if err != nil {
		return responseData, fmt.Errorf("unable to create base url object - %w", err)
	}

	requestURL.RawQuery = payload.Encode()

	// create the request
	request, err := http.NewRequest("GET", requestURL.String(), http.NoBody)
	if err != nil {
		return responseData, fmt.Errorf("unable to create http request - %w", err)
	}

	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", poller.Token.BearerToken))

	// send the request
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return responseData, fmt.Errorf("error in http request - %w", err)
	}
	defer response.Body.Close()

	// check the response code
	if response.StatusCode != http.StatusOK {
		return responseData, fmt.Errorf(
			"invalid response code [%d]; message [%s] - %w",
			response.StatusCode,
			response.Status,
			ErrResponseInvalid,
		)
	}

	// unmarshal the response
	responseReader := io.NopCloser(response.Body)
	if err := json.NewDecoder(responseReader).Decode(&responseData); err != nil {
		return responseData, fmt.Errorf("unable to read response body - %w", err)
	}

	return responseData, nil
}

func serviceLogPath(token *token) string {
	return fmt.Sprintf("%s/api/service_logs/v1/cluster_logs", token.Endpoint)
}
