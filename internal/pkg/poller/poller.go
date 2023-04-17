package poller

import (
	"fmt"

	sdk "github.com/openshift-online/ocm-sdk-go"
	v1 "github.com/openshift-online/ocm-sdk-go/servicelogs/v1"

	"github.com/scottd018/ocm-log-forwarder/internal/pkg/processor"
)

const (
	defaultPollerRequestSize = 1000
)

type Poller struct {
	Client *sdk.Connection
}

func NewPoller(proc *processor.Processor) (*Poller, error) {
	// retrieve the token
	token, err := NewToken(proc.Config.TokenFile)
	if err != nil {
		return &Poller{}, fmt.Errorf("unable to create poller token - %w", err)
	}

	// set the client
	client, err := sdk.NewConnectionBuilder().
		Tokens(token.RefreshToken).
		Build()
	if err != nil {
		return &Poller{}, fmt.Errorf("unable to build poller connection - %w", err)
	}

	return &Poller{Client: client}, nil
}

func (poller *Poller) Request(proc *processor.Processor) (response Response, err error) {
	if poller.Client == nil {
		return response, fmt.Errorf("missing client from poller object - %w", ErrTokenInvalid)
	}

	// loop through each of the pages and generate a response that stores
	// all of the messages.
	page := 1
	for {
		logResponse, err := poller.RequestPage(proc, page)
		if err != nil {
			return response, err
		}

		// ensure the response was ok
		logs, ok := logResponse.GetItems()
		if !ok {
			return response, fmt.Errorf("unable to retrieve logs from response page [%d]", page)
		}

		// append the items to the response
		response.Logs = append(response.Logs, logs.Slice()...)
		response.Total = logResponse.Total()
		response.Size = logResponse.Size()

		if response.PageCount() == page {
			break
		}

		page++
	}

	return response, nil
}

func (poller *Poller) RequestPage(proc *processor.Processor, pageNum int) (*v1.ClusterLogsListResponse, error) {
	request := poller.Client.ServiceLogs().
		V1().
		ClusterLogs().
		List().
		Search(fmt.Sprintf("cluster_id = '%s'", proc.Config.ClusterID)).
		Size(defaultPollerRequestSize).
		Page(pageNum)

	// send the request
	response, err := request.Send()
	if err != nil {
		return &v1.ClusterLogsListResponse{}, fmt.Errorf("error requesting service logs - %w", err)
	}

	return response, nil
}
