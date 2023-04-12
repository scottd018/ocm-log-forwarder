package main

import (
	"fmt"
	"os"

	"github.com/scottd018/ocm-log-forwarder/internal/pkg/config"
	"github.com/scottd018/ocm-log-forwarder/internal/pkg/controller"
)

func main() {
	// pre-validate and store the config
	cfg, err := config.Initialize()
	if err != nil {
		panic(fmt.Errorf("unable to initialize config - %w", err))
	}

	// create the ctrl
	ctrl, err := controller.NewController(cfg)
	if err != nil {
		panic(fmt.Errorf("unable to initialize controller - %w", err))
	}

	// run the controller
	if err := ctrl.Run(); err != nil {
		ctrl.Processor.Log.ErrorF("error in control loop - %s", err)

		os.Exit(1)
	}
}
