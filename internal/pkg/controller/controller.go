package controller

import (
	"fmt"
	"time"

	"github.com/scottd018/ocm-log-forwarder/internal/pkg/backend"
	"github.com/scottd018/ocm-log-forwarder/internal/pkg/config"
	"github.com/scottd018/ocm-log-forwarder/internal/pkg/poller"
	"github.com/scottd018/ocm-log-forwarder/internal/pkg/processor"
)

type Controller struct {
	Config    *config.Config
	Backend   backend.Backend
	Poller    *poller.Poller
	Processor *processor.Processor
}

func NewController(cfg *config.Config) (*Controller, error) {
	// initialize the processor
	proc, err := processor.NewProcessor(cfg)
	if err != nil {
		return &Controller{}, fmt.Errorf("unable to initialize processor - %w", err)
	}

	// initialize the backend
	proc.Log.Infof("initializing backend: cluster=[%s], type=[%s]", cfg.ClusterID, cfg.Backend)
	logBackend, err := backend.FromConfig(proc)
	if err != nil {
		return &Controller{}, fmt.Errorf("unable to initialize backend - %w", err)
	}

	// create the poller
	proc.Log.Infof("initializing poller: cluster=[%s], interval=[%v minutes]", cfg.ClusterID, cfg.PollerInterval.Minutes())
	ocmPoller, err := poller.NewPoller(proc)
	if err != nil {
		return &Controller{}, fmt.Errorf("unable to initialize poller - %w", err)
	}

	return &Controller{
		Config:    cfg,
		Backend:   logBackend,
		Poller:    ocmPoller,
		Processor: proc,
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
		}
	}
}

func (controller *Controller) Loop(loopSignal <-chan poller.Poller, errorSignal chan<- error) {
	for loop := range loopSignal {
		// poll ocm for service logs
		controller.Processor.Log.InfoF("polling openshift cluster manager: cluster=[%s]", controller.Processor.Config.ClusterID)
		response, err := loop.Poll(controller.Processor)
		if err != nil {
			errorSignal <- err

			return
		}

		// debug the response data
		if controller.Processor.Config.Debug {
			responseData, err := response.Data()
			if err != nil {
				errorSignal <- err
			}

			controller.Processor.Log.DebugF("response data: %+v", responseData)
		}

		// send service logs to the backend
		if err := controller.Backend.Send(controller.Processor, &response); err != nil {
			errorSignal <- err

			return
		}
	}
}
