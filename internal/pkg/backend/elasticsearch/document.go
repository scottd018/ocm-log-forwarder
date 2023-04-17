package elasticsearch

import v1 "github.com/openshift-online/ocm-sdk-go/servicelogs/v1"

// ElasticSearchDocument represents the final document that gets sent
// to ElasticSearch.  These are the fields that show in in the index
// as represented by the json tags.
type ElasticSearchDocument struct {
	id          string
	ClusterID   string `json:"cluster_id"`
	ExternalID  string `json:"external_id"`
	Username    string `json:"username"`
	Severity    string `json:"severity"`
	ServiceName string `json:"service_name"`
	EventID     string `json:"event_stream_id"`
	Message     string `json:"message"`
	Timestamp   string `json:"@timestamp"`
}

// buildDocument builds an ElasticSearch document from a service log message.
func buildDocument(log *v1.LogEntry) *ElasticSearchDocument {
	return &ElasticSearchDocument{
		id:          log.ID(),
		ClusterID:   log.ClusterID(),
		ExternalID:  log.ClusterUUID(),
		Username:    log.Username(),
		Severity:    string(log.Severity()),
		EventID:     log.EventStreamID(),
		ServiceName: log.ServiceName(),
		Message:     log.Summary(),
		Timestamp:   log.Timestamp().String(),
	}
}
