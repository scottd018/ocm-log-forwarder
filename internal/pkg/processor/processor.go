package processor

import (
	"fmt"
	"os"

	"github.com/apsdehal/go-logger"
	"github.com/scottd018/ocm-log-forwarder/internal/pkg/config"
	"golang.org/x/net/context"
)

type Processor struct {
	Log          *logger.Logger
	Config       *config.Config
	Context      context.Context
	ResponseData map[string]interface{}
}

func NewProcessor(cfg *config.Config) (*Processor, error) {
	// create a logger using the backend string as the designator
	log, err := logger.New(cfg.Backend, 1, os.Stdout)
	if err != nil {
		return &Processor{}, fmt.Errorf("unable to setup logger - %w", err)
	}

	return &Processor{
		Log:     log,
		Config:  cfg,
		Context: context.Background(),
	}, nil
}
