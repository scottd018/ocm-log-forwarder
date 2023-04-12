package backend

import (
	"fmt"

	"github.com/scottd018/ocm-log-forwarder/internal/pkg/config"
	"github.com/scottd018/ocm-log-forwarder/internal/pkg/processor"
)

type Backend interface {
	Send(*processor.Processor) error
	Initialize(*processor.Processor) error
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
	EventID   string `json:"event_stream_id"`
}

func Initialize(backend Backend, proc *processor.Processor) error {
	switch obj := backend.(type) {
	case *ElasticSearch:
		return obj.Initialize(proc)
	default:
		return fmt.Errorf("backend [%T] - %w", obj, config.ErrBackendUnknown)
	}
}

func FromConfig(proc *processor.Processor) (Backend, error) {
	var backend Backend

	switch proc.Config.Backend {
	case config.DefaultBackendElasticSearch:
		backend = &ElasticSearch{}
	default:
		return backend, fmt.Errorf(
			"backend from environment [%s=%s] - %w",
			config.DefaultEnvironmentBackend,
			proc.Config.Backend,
			config.ErrBackendUnknown,
		)
	}

	// initialize the backend from the environment
	if err := backend.Initialize(proc); err != nil {
		return backend, fmt.Errorf("unable to initialize backend - %w", err)
	}

	return backend, nil
}
