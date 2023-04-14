package elasticsearch

import "github.com/scottd018/ocm-log-forwarder/internal/pkg/poller"

type ElasticSearchDocument struct {
	id        string
	ClusterID string `json:"cluster_id"`
	Username  string `json:"username"`
	Severity  string `json:"severity"`
	EventID   string `json:"event_stream_id"`
	CreatedBy string `json:"created_by"`
	Message   string `json:"message"`
	Timestamp string `json:"@timestamp"`
}

// buildDocument builds an ElasticSearch document from a service log message.
func buildDocument(message *poller.ServiceLogMessage) *ElasticSearchDocument {
	return &ElasticSearchDocument{
		id:        message.ID,
		ClusterID: message.ClusterID,
		Username:  message.Username,
		Severity:  message.Severity,
		EventID:   message.EventID,
		CreatedBy: message.CreatedBy,
		Message:   message.Summary,
		Timestamp: message.Timestamp,
	}
}
