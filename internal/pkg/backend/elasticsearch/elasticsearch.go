package elasticsearch

import (
	"fmt"

	"github.com/olivere/elastic/v7"
	v1 "github.com/openshift-online/ocm-sdk-go/servicelogs/v1"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/scottd018/ocm-log-forwarder/internal/pkg/config"
	"github.com/scottd018/ocm-log-forwarder/internal/pkg/poller"
	"github.com/scottd018/ocm-log-forwarder/internal/pkg/processor"
)

const (
	elasticSearchBatchSize = 100
)

type ElasticSearch struct {
	Client          *elastic.Client
	Documents       []ElasticSearchDocument
	SentDocumentIDs []string
}

func (es *ElasticSearch) Initialize(proc *processor.Processor) (err error) {
	var client *elastic.Client

	// create the client based on the authentication type
	switch authType := config.GetElasticSearchAuthType(); {
	case authType == config.DefaultBackendAuthTypeBasic:
		client, err = getAuthTypeBasic(proc)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("auth type [%s] - %w", authType, config.ErrBackendAuthUnknown)
	}

	// store the client on the elasticsearch object
	es.Client = client

	return nil
}

func (es *ElasticSearch) Send(proc *processor.Processor, response *poller.Response) error {
	var batchCount int

	documents := es.UnsentDocuments(response.Logs)

	documentCount := len(documents)

	// return if there are no unsent documents to send
	if documentCount == 0 {
		return nil
	}

	// we want to do this serially so we do not overwhelm the elasticsearch api
	for i := 0; i < documentCount; i += elasticSearchBatchSize {
		// get the lastDocument item for this batch
		lastDocument := i + elasticSearchBatchSize

		// reset the last item to the last document in the list
		// if it becomes out of range
		if lastDocument > documentCount {
			lastDocument = documentCount
		}

		// create a request for this batch
		documentBatch := documents[i:lastDocument]

		response, err := es.BuildRequest(proc, documentBatch).BatchSend(proc)
		if err != nil {
			es.Log(log.Err(err), fmt.Sprintf("batch number [%d] failed to send", batchCount))
		}

		// append the batch count and handle the response
		batchCount++
		es.handleResponse(proc, response)
	}

	return nil
}

// BuildRequest builds an ElasticSearchRequest object from a set of documents.
func (es *ElasticSearch) BuildRequest(proc *processor.Processor, documents []*ElasticSearchDocument) *ElasticSearchRequest {
	docChan := make(chan *ElasticSearchDocument, len(documents))

	request := &ElasticSearchRequest{
		Index:     config.GetElasticSearchIndex(),
		Bulk:      es.Client.Bulk().Index(config.GetElasticSearchIndex()),
		Documents: make([]*ElasticSearchDocument, len(documents)),
	}

	// build the request for the individual document and add it to
	// the batch send routine
	for i := range documents {
		go func(this *ElasticSearchDocument) {
			docChan <- this
		}(documents[i])
	}

	defer close(docChan)

	// if we find one that matches, return
	for i := 0; i < len(documents); i++ {
		document := <-docChan

		// add the document to the bulk request
		es.Log(log.Info().Str(
			"cluster", proc.Config.ClusterID).Str("id", document.id).Str("index", request.Index),
			"adding document to elasticsearch bulk request",
		)
		es.Log(log.Debug().Str("document", fmt.Sprintf("%+v", document)), "debugging document")
		request.Bulk.Add(elastic.NewBulkIndexRequest().Id(document.id).Doc(document))
	}

	return request
}

// UnsentDocuments builds an array of ElasticSearch documents from an array of service log
// messages.
func (es *ElasticSearch) UnsentDocuments(logs []*v1.LogEntry) []*ElasticSearchDocument {
	documents := []*ElasticSearchDocument{}

	for i := range logs {
		document := buildDocument(logs[i])

		if es.HasSent(document) {
			continue
		}

		documents = append(documents, document)
	}

	return documents
}

func (es *ElasticSearch) HasSent(newDocument *ElasticSearchDocument) bool {
	// return immediately if we have not sent any documents
	if len(es.SentDocumentIDs) < 1 {
		return false
	}

	for i := range es.SentDocumentIDs {
		if newDocument.id == es.SentDocumentIDs[i] {
			return true
		}
	}

	return false
}

func (es *ElasticSearch) String() string {
	return config.DefaultBackendElasticSearch
}

func (es *ElasticSearch) Log(event *zerolog.Event, message string) {
	event.Str("source", fmt.Sprintf("%s-backend", es.String())).Msg(message)
}

// handleResponse handles the response for an elasticsearch request.  It stores successful
// items on the object and logs any unsuccessful or updated items.
func (es *ElasticSearch) handleResponse(proc *processor.Processor, response *elastic.BulkResponse) {
	// check for failures in the responses and log
	if response.Errors {
		for _, failed := range response.Failed() {
			es.Log(
				log.Error().
					Str("index", failed.Index).
					Str("message_id", failed.Id).
					Int("status_code", failed.Status).
					Str("result", failed.Result),
				"error in elasticsearch request")
		}
	}

	// check for creates in the responses and log
	if len(response.Created()) > 0 {
		for _, created := range response.Created() {
			es.Log(log.Info().Str("message_id", created.Id), "created elasticsearch id")
		}
	}

	// check for updates in the responses and log
	if len(response.Updated()) > 0 {
		for _, updated := range response.Updated() {
			es.Log(log.Info().Str("message_id", updated.Id), "updated elasticsearch id")
		}
	}

	// check for successes and log (debug only)
	if len(response.Succeeded()) > 0 {
		for _, succeeded := range response.Succeeded() {
			es.SentDocumentIDs = append(es.SentDocumentIDs, succeeded.Id)

			es.Log(log.Debug().Str("message_id", succeeded.Id), "succeeded elasticsearch id")
		}
	}
}
