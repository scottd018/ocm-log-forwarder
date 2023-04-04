package backend

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"

	elasticsearch "github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"

	"github.com/scottd018/ocm-log-forwarder/internal/pkg/config"
	"github.com/scottd018/ocm-log-forwarder/internal/pkg/processor"
)

type ElasticSearch struct {
	Client  *elasticsearch.Client
	Request *ElasticSearchRequest
}

type ElasticSearchRequest struct {
	Documents []ElasticSearchDocument
}

type ElasticSearchDocument struct {
	ClusterID string `json:"cluster_id"`
	Username  string `json:"username"`
	Severity  string `json:"severity"`
	Message   string `json:"message"`
	Timestamp string `json:"@timestamp"`
}

func (es *ElasticSearch) Send(proc *processor.Processor) error {
	// serialize data to documents
	documents := Documents{}

	// collect errors
	var errors []error

	if err := json.Unmarshal(proc.ResponseData, &documents); err != nil {
		return fmt.Errorf("unable to unmarshal response data to documents object - %w", err)
	}

	for i := range documents.Items {
		// create the elasticsearch document from the generic backend document
		esDoc := ElasticSearchDocument{
			ClusterID: documents.Items[i].ClusterID,
			Username:  documents.Items[i].Username,
			Severity:  documents.Items[i].Severity,
			Message:   documents.Items[i].Summary,
			Timestamp: documents.Items[i].Timestamp,
		}

		// skip processing if this has been sent
		if es.Request.Sent(esDoc) {
			continue
		}

		// serialize document as json
		esBody, err := json.Marshal(esDoc)
		if err != nil {
			errors = append(errors, fmt.Errorf("unable to serialize document - %w", err))

			continue
		}

		// create an Elasticsearch request to index the document
		// TODO: index input
		request := esapi.IndexRequest{
			Index: "my_index",
			Body:  bytes.NewReader(esBody),
		}

		// send the request to elasticsearch
		response, err := request.Do(proc.Context, es.Client)
		if err != nil {
			errors = append(errors, fmt.Errorf("error in elasticsearch request - %w", err))

			continue
		}
		defer response.Body.Close()

		// check status code
		if response.IsError() {
			errors = append(errors, fmt.Errorf(
				"error in elasticsearch request with status code [%d] and message [%s] - %w",
				response.StatusCode,
				response.Status(),
				err),
			)

			continue
		}

		// if we have made it this far, we are successful and can append the document to the list
		// of documents
		es.Request.Documents = append(es.Request.Documents, esDoc)
	}

	// if we found any errors, return them
	if len(errors) > 0 {
		return fmt.Errorf("errors: %v", errors)
	}

	return nil
}

func (es *ElasticSearch) Initialize() error {
	client, err := elasticsearch.NewDefaultClient()
	if err != nil {
		return fmt.Errorf("unable to initialize the elasticsearch client - %w", err)
	}
	es.Client = client
	es.Request = &ElasticSearchRequest{}

	return nil
}

func (es *ElasticSearch) String() string {
	return config.DefaultBackendElasticSearch
}

func (req *ElasticSearchRequest) Sent(document ElasticSearchDocument) bool {
	for i := range req.Documents {
		// skip adding a document to the request if we have already added it
		if reflect.DeepEqual(req.Documents[i], document) {
			return true
		}
	}

	req.Documents = append(req.Documents, document)

	return false
}
