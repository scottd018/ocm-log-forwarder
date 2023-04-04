package backend

import (
	"errors"
	"fmt"

	"github.com/scottd018/ocm-log-forwarder/internal/pkg/config"
)

const (
	envBackendType = "BACKEND_TYPE"
)

var (
	ErrUnknownBackend = errors.New("unknown backend type")
)

type Backend interface {
	Send() error
	Initialize() error
	String() string
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
		return backend, fmt.Errorf("backend from environment [%s=%s] - %w", envBackendType, cfg.Backend, ErrUnknownBackend)
	}

	// initialize the backend from the environment
	if err := backend.Initialize(); err != nil {
		return backend, fmt.Errorf("unable to initialize backend - %w", err)
	}

	return backend, nil
}
