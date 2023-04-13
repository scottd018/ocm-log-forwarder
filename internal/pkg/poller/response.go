package poller

import (
	"encoding/json"
	"fmt"
)

// Response stores an array of ResponseItems.  It represents
// a response from OCM.
type Response struct {
	Messages []*ServiceLogMessage `json:"items"`

	Page  int `json:"page,omitempty"`
	Size  int `json:"size,omitempty"`
	Total int `json:"total,omitempty"`
}

// ServiceLogMessage encompasses an individual service log message.
type ServiceLogMessage struct {
	ID        string `json:"id"`
	Summary   string `json:"summary"`
	Timestamp string `json:"timestamp"`
	ClusterID string `json:"cluster_id"`
	Username  string `json:"username"`
	Severity  string `json:"severity"`
	EventID   string `json:"event_stream_id"`
	CreatedBy string `json:"created_by"`
}

// PageCount returns the total number of pages in the response.
func (response *Response) PageCount() int {
	pages := response.Total / response.Size

	// if there are any leftover pages, add it as a final page
	if (response.Total % response.Size) > 0 {
		pages += 1
	}

	return pages
}

// Data returns the response data as a string format which is useful for logging.
func (response *Response) Data() (string, error) {
	jsonBytes, err := json.Marshal(response)
	if err != nil {
		return "", fmt.Errorf("unable to generate json from response - %w", err)
	}

	return string(jsonBytes), nil
}
