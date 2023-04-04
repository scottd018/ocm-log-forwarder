package backend

import (
	"errors"
	"fmt"

	"github.com/scottd018/ocm-log-forwarder/internal/pkg/config"
	"github.com/scottd018/ocm-log-forwarder/internal/pkg/processor"
)

var (
	ErrUnknownBackend = errors.New("unknown backend type")
)

type Backend interface {
	Send(*processor.Processor) error
	Initialize() error
	String() string
}

// Documents stores an array of Documents.
type Documents struct {
	Items []Document `json:"items"`
}

// Document stores the important information from the service log to
// be shipped to the backend.
type Document struct {
	ClusterID string `json:"cluster_id"`
	Username  string `json:"username"`
	Severity  string `json:"severity"`
	Timestamp string `json:"timestamp"`
	Summary   string `json:"summary"`
}

func Initialize(backend Backend) error {
	switch obj := backend.(type) {
	case *ElasticSearch:
		return obj.Initialize()
	default:
		return fmt.Errorf("backend [%T] - %w", obj, ErrUnknownBackend)
	}
}

func FromConfig(cfg *config.Config) (Backend, error) {
	var backend Backend

	switch cfg.Backend {
	case config.DefaultBackendElasticSearch:
		backend = &ElasticSearch{}
	default:
		return backend, fmt.Errorf("backend from environment [%s=%s] - %w", config.DefaultEnvironmentBackend, cfg.Backend, ErrUnknownBackend)
	}

	// initialize the backend from the environment
	if err := backend.Initialize(); err != nil {
		return backend, fmt.Errorf("unable to initialize backend - %w", err)
	}

	return backend, nil
}
