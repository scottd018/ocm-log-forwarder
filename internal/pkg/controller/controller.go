package controller

import (
	"fmt"
	"time"

	"github.com/scottd018/ocm-log-forwarder/internal/pkg/backend"
	"github.com/scottd018/ocm-log-forwarder/internal/pkg/config"
	"github.com/scottd018/ocm-log-forwarder/internal/pkg/poller"
	"github.com/scottd018/ocm-log-forwarder/internal/pkg/processor"
)

const (
	minPollIntervalMinutes int64 = 1    // 1 minute minimum
	maxPollIntervalMinutes int64 = 1440 // 1 day maximum
)

type Controller struct {
	Config    *config.Config
	Backend   backend.Backend
	Poller    *poller.Poller
	Processor *processor.Processor
}

func NewController(cfg *config.Config) (*Controller, error) {
	// initialize the processor
	processor, err := processor.NewProcessor(cfg)
	if err != nil {
		return &Controller{}, fmt.Errorf("unable to initialize processor - %w", err)
	}

	// initialize the backend
	processor.Log.Infof("initializing backend: cluster=[%s], type=[%s]", cfg.ClusterID, cfg.Backend)
	backend, err := backend.FromConfig(cfg)
	if err != nil {
		return &Controller{}, fmt.Errorf("unable to initialize backend - %w", err)
	}

	// create the poller
	processor.Log.Infof("initializing poller: cluster=[%s], interval=[%v minutes]", cfg.ClusterID, cfg.PollerInterval.Minutes())
	poller, err := poller.NewPoller(processor)
	if err != nil {
		return &Controller{}, fmt.Errorf("unable to initialize poller - %w", err)
	}

	return &Controller{
		Config:    cfg,
		Backend:   backend,
		Poller:    poller,
		Processor: processor,
	}, nil
}

func (controller *Controller) Run() error {
	// create a channel to signal the task to run its loop
	loopSignal := make(chan poller.Poller)

	// create a channel to send errors
	errorSignal := make(chan error)

	// start the go routine
	controller.Processor.Log.InfoF("starting main program loop")
	go controller.Loop(loopSignal, errorSignal)

	// run immediately
	loopSignal <- *controller.Poller

	// run the task for each poll interval
	ticker := time.NewTicker(controller.Config.PollerInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			loopSignal <- *controller.Poller
		case err := <-errorSignal:
			// log and return the error if we received one
			controller.Processor.Log.Error(err.Error())

			return err
		default:
			break
		}
	}
}

func (controller *Controller) Loop(loopSignal <-chan poller.Poller, errorSignal chan<- error) {
	for {
		select {
		case <-loopSignal:
			controller.Processor.Log.InfoF("polling openshift cluster manager: cluster=[%s]", controller.Processor.Config.ClusterID)

			if err := controller.Poller.Poll(controller.Processor); err != nil {
				errorSignal <- err

				return
			}

			controller.Processor.Log.Infof("response data: %s", controller.Processor.ResponseData)

			if err := controller.Backend.Send(controller.Processor); err != nil {
				errorSignal <- err

				return
			}
		}
	}
}
