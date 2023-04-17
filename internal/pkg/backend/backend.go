package backend

import (
	"fmt"

	"github.com/scottd018/ocm-log-forwarder/internal/pkg/backend/elasticsearch"
	"github.com/scottd018/ocm-log-forwarder/internal/pkg/backend/stdout"
	"github.com/scottd018/ocm-log-forwarder/internal/pkg/config"
	"github.com/scottd018/ocm-log-forwarder/internal/pkg/poller"
	"github.com/scottd018/ocm-log-forwarder/internal/pkg/processor"
)

type Backend interface {
	Send(*processor.Processor, *poller.Response) error
	Initialize(*processor.Processor) error
	String() string
}

func Initialize(proc *processor.Processor) (Backend, error) {
	var backend Backend

	switch proc.Config.Backend {
	case config.DefaultBackendElasticSearch:
		backend = &elasticsearch.ElasticSearch{}
	case config.DefaultBackendStdOut:
		backend = &stdout.StdOut{}
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
		return backend, fmt.Errorf("unable to initialize elasticsearch backend - %w", err)
	}

	return backend, nil
}
