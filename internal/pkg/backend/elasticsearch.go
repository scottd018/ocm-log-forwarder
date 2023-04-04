package backend

import (
	"fmt"

	elasticsearch "github.com/elastic/go-elasticsearch/v8"
)

const (
	envBackendTypeElasticSearch = "elasticsearch"
)

type ElasticSearch struct {
	Client *elasticsearch.Client
}

func (es *ElasticSearch) Send() error {
	return nil
}

func (es *ElasticSearch) Initialize() error {
	client, err := elasticsearch.NewDefaultClient()
	if err != nil {
		return fmt.Errorf("unable to initialize the elasticsearch client - %w", err)
	}
	es.Client = client

	return nil
}

func (es *ElasticSearch) String() string {
	return envBackendTypeElasticSearch
}
