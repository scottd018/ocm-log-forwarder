package backend

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"time"

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
	EventID   string `json:"event_stream_id"`
}

func (es *ElasticSearch) Initialize(proc *processor.Processor) error {
	es.Request = &ElasticSearchRequest{}

	var esConfig elasticsearch.Config

	// set the authentication parameters
	switch authType := config.GetElasticSearchAuthType(); {
	case authType == config.DefaultBackendAuthTypeBasic:
		username, password, err := config.GetElasticSearchAuthTypeBasic(proc.KubeClient, proc.Context)
		if err != nil {
			return fmt.Errorf("unable to configure basic auth type - %w", err)
		}

		// create the basic auth connection config
		esConfig = elasticsearch.Config{
			Addresses: []string{config.GetElasticSearchURL()},
			Username:  username,
			Password:  password,
			Transport: &http.Transport{
				MaxIdleConnsPerHost:   10,
				ResponseHeaderTimeout: time.Second,
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		}
	default:
		return fmt.Errorf("unknown auth type [%s]", authType)
	}

	// create the elasticsearch client
	client, err := elasticsearch.NewClient(esConfig)
	if err != nil {
		fmt.Println("Error creating Elasticsearch client:", err)
		return fmt.Errorf("unable to configure elasticsearch client - %w", err)
	}
	es.Client = client

	return nil
}

func (es *ElasticSearch) Send(proc *processor.Processor) error {
	// serialize data to documents
	documents := Documents{}

	index := config.GetElasticSearchIndex()

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
			EventID:   documents.Items[i].EventID,
		}

		// skip processing if this has been sent
		if es.Request.Sent(esDoc) {
			continue
		}

		// serialize document as json
		esBody, err := json.Marshal(esDoc)
		if err != nil {
			proc.Log.ErrorF("error serializing json for message [%s] - %s", esDoc.Message, err)

			continue
		}

		// create an Elasticsearch request to index the document
		request := esapi.IndexRequest{
			Index: index,
			Body:  bytes.NewReader(esBody),
		}

		// send the request to elasticsearch
		proc.Log.Infof("sending items to elasticsearch: cluster=%s, event_stream_id=%s, index=%s", proc.Config.ClusterID, esDoc.EventID, index)
		response, err := request.Do(proc.Context, es.Client)
		if err != nil {
			proc.Log.ErrorF("error sending request to elasticsearch for message [%s] - %s", esDoc.Message, err)

			continue
		}
		defer response.Body.Close()

		// check status code
		if response.IsError() {
			proc.Log.ErrorF(
				"error in elasticsearch request with status code [%d] and message [%s] - %w",
				response.StatusCode,
				response.Status(),
				err,
			)

			continue
		}

		// if we have made it this far, we are successful and can append the document to the list
		// of documents
		es.Request.Documents = append(es.Request.Documents, esDoc)
	}

	return nil
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

func (es *ElasticSearch) String() string {
	return config.DefaultBackendElasticSearch
}
