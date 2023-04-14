package elasticsearch

import (
	"fmt"

	"github.com/olivere/elastic/v7"
	"github.com/scottd018/ocm-log-forwarder/internal/pkg/processor"
)

type ElasticSearchRequest struct {
	Index     string
	Documents []*ElasticSearchDocument
	Bulk      *elastic.BulkService
}

func (req *ElasticSearchRequest) BatchSend(proc *processor.Processor) (*elastic.BulkResponse, error) {
	// return if we have no documents to send as part of the request
	if req.Bulk.NumberOfActions() < 1 {
		return nil, nil
	}

	// send the bulk request
	proc.Log.Infof(
		"sending [%d] documents to elasticsearch: cluster=%s, index=%s",
		req.Bulk.NumberOfActions(),
		proc.Config.ClusterID,
		req.Index,
	)
	bulkResponse, err := req.Bulk.Do(proc.Context)
	if err != nil {
		return bulkResponse, fmt.Errorf("error sending bulk request to elasticsearch - %w", err)
	}

	return bulkResponse, nil
}
