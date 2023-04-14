package backend

import (
	"fmt"

	"github.com/scottd018/ocm-log-forwarder/internal/pkg/backend/elasticsearch"
	"github.com/scottd018/ocm-log-forwarder/internal/pkg/config"
	"github.com/scottd018/ocm-log-forwarder/internal/pkg/poller"
	"github.com/scottd018/ocm-log-forwarder/internal/pkg/processor"
)

type Backend interface {
	Send(*processor.Processor, *poller.Response) error
	Initialize(*processor.Processor) error
	String() string
}

func Initialize(backend Backend, proc *processor.Processor) error {
	switch obj := backend.(type) {
	case *elasticsearch.ElasticSearch:
		return obj.Initialize(proc)
	default:
		return fmt.Errorf("backend [%T] - %w", obj, config.ErrBackendUnknown)
	}
}

func FromConfig(proc *processor.Processor) (Backend, error) {
	var backend Backend

	switch proc.Config.Backend {
	case config.DefaultBackendElasticSearch:
		backend = &elasticsearch.ElasticSearch{}
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
