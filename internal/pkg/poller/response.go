package poller

import (
	"encoding/json"
	"fmt"

	v1 "github.com/openshift-online/ocm-sdk-go/servicelogs/v1"
)

// Response stores an array of ResponseItems.  It represents
// a response from OCM.
type Response struct {
	Logs []*v1.LogEntry

	Size  int `json:"size,omitempty"`
	Total int `json:"total,omitempty"`
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
