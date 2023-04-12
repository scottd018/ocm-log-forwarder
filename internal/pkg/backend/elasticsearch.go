package backend

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"

	"github.com/olivere/elastic/v7"

	"github.com/scottd018/ocm-log-forwarder/internal/pkg/config"
	"github.com/scottd018/ocm-log-forwarder/internal/pkg/processor"
)

const (
	elasticSearchDocType = "_doc"
)

type ElasticSearch struct {
	Client  *elastic.Client
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
	EventID   string `json:"event_stream_id"`
}

func (es *ElasticSearch) Initialize(proc *processor.Processor) error {
	var client *elastic.Client

	// create the client based on the authentication type
	switch authType := config.GetElasticSearchAuthType(); {
	case authType == config.DefaultBackendAuthTypeBasic:
		username, password, err := config.GetElasticSearchAuthTypeBasic(proc.KubeClient, proc.Context)
		if err != nil {
			return fmt.Errorf("unable to configure basic auth type - %w", err)
		}

		client, err = elastic.NewClient(
			elastic.SetSniff(false),
			elastic.SetURL(config.GetElasticSearchURL()),
			elastic.SetBasicAuth(username, password),
			elastic.SetHttpClient(
				&http.Client{
					Transport: &http.Transport{
						TLSClientConfig: &tls.Config{
							InsecureSkipVerify: true,
						},
					},
				},
			),
		)

		if err != nil {
			return fmt.Errorf("unable to create elasticsearch client - %w", err)
		}
	default:
		return fmt.Errorf("auth type [%s] - %w", authType, config.ErrBackendAuthUnknown)
	}

	// store the client and request on the elasticsearch object
	es.Client = client
	es.Request = &ElasticSearchRequest{}

	return nil
}

func (es *ElasticSearch) Send(proc *processor.Processor) error {
	// serialize data to documents
	documents := Documents{}
	if err := json.Unmarshal(proc.ResponseData, &documents); err != nil {
		return fmt.Errorf("unable to unmarshal response data to documents object - %w", err)
	}

	index := config.GetElasticSearchIndex()

	bulkRequest := es.Client.Bulk().Index(index).Type(elasticSearchDocType)

	var documentCount int

	for i := range documents.Items {
		// create the elasticsearch document from the generic backend document
		esDoc := ElasticSearchDocument{
			ClusterID: documents.Items[i].ClusterID,
			Username:  documents.Items[i].Username,
			Severity:  documents.Items[i].Severity,
			Message:   documents.Items[i].Summary,
			Timestamp: documents.Items[i].Timestamp,
			EventID:   documents.Items[i].EventID,
		}

		// skip processing if this has been sent
		if es.Request.Sent(&esDoc) {
			continue
		}

		// serialize document as json
		esBody, err := json.Marshal(esDoc)
		if err != nil {
			proc.Log.ErrorF("error serializing json for message [%s] - %s", esDoc.Message, err)

			continue
		}

		// add the document to the bulk request
		proc.Log.Infof(
			"adding document to elasticsearch bulk request: cluster=%s, event_stream_id=%s, index=%s",
			proc.Config.ClusterID,
			esDoc.EventID,
			index,
		)
		bulkRequest.Add(elastic.NewBulkIndexRequest().Id(esDoc.EventID).Doc(esBody))

		// if we have made it this far, we are successful and can append the document to the list
		// of documents and increment the document count to notify the logger how many
		// documents we are sending as part of the request
		documentCount++

		es.Request.Documents = append(es.Request.Documents, esDoc)
	}

	// return if we have no documents to send as part of the request
	if documentCount == 0 {
		return nil
	}

	// send the bulk request
	proc.Log.Infof("sending [%d] documents to elasticsearch: cluster=%s, index=%s", documentCount, proc.Config.ClusterID, index)
	response, err := bulkRequest.Do(proc.Context)
	if err != nil {
		return fmt.Errorf("error sending bulk request to elasticsearch - %s", err)
	}

	// check for failures in the responses
	if response.Errors {
		for _, item := range response.Items {
			for _, result := range item {
				if result.Status != http.StatusOK {
					proc.Log.ErrorF(
						"error in elasticsearch request: status_code=%d, message=%s",
						result.Status,
						result.Result,
					)
				}
			}
		}
	}

	return nil
}

func (req *ElasticSearchRequest) Sent(document *ElasticSearchDocument) bool {
	for i := range req.Documents {
		// skip adding a document to the request if we have already added it
		if reflect.DeepEqual(req.Documents[i], document) {
			return true
		}
	}

	return false
}

func (es *ElasticSearch) String() string {
	return config.DefaultBackendElasticSearch
}
