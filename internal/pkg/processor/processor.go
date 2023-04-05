package processor

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/apsdehal/go-logger"
	"golang.org/x/net/context"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/scottd018/ocm-log-forwarder/internal/pkg/config"
	"github.com/scottd018/ocm-log-forwarder/internal/pkg/utils"
)

type Processor struct {
	Log          *logger.Logger
	Config       *config.Config
	Context      context.Context
	ResponseData []byte
	KubeClient   *kubernetes.Clientset
}

func NewProcessor(cfg *config.Config) (*Processor, error) {
	// create a logger using the backend string as the designator
	log, err := logger.New(cfg.Backend, 1, os.Stdout)
	if err != nil {
		return &Processor{}, fmt.Errorf("unable to setup logger - %w", err)
	}

	if cfg.Debug {
		log.SetLogLevel(logger.DebugLevel)
	}

	// create the processor
	processor := &Processor{
		Log:     log,
		Config:  cfg,
		Context: context.Background(),
	}

	// create the kubernetes client and store it on the processor
	client, err := newKubeClient(*processor)
	if err != nil {
		return &Processor{}, fmt.Errorf("error creating kubernetes client - %w", err)
	}
	processor.KubeClient = client

	return processor, nil
}

func newKubeClient(proc Processor) (*kubernetes.Clientset, error) {
	proc.Log.InfoF("initializing kubernetes cluster config: cluster=[%s]", proc.Config.ClusterID)
	config, err := rest.InClusterConfig()
	if err == nil {
		// create the clientset for the config
		client, err := kubernetes.NewForConfig(config)
		if err != nil {
			return &kubernetes.Clientset{}, fmt.Errorf("unable to create kubernetes client from in cluster - %w", err)
		}

		return client, nil
	}

	proc.Log.WarningF("unable to initialize cluster config: cluster=[%s], attempting file initialization", proc.Config.ClusterID)

	kubeConfig := kubeConfigPath()

	proc.Log.InfoF("initializing kubernetes file config: cluster=[%s], file=[%s]", proc.Config.ClusterID, kubeConfig)
	config, err = clientcmd.BuildConfigFromFlags("", kubeConfig)
	if err == nil {
		// create the clientset for the config
		client, err := kubernetes.NewForConfig(config)
		if err != nil {
			return &kubernetes.Clientset{}, fmt.Errorf("unable to create kubernetes client from kubeconfig - %w", err)
		}

		return client, nil
	}

	return &kubernetes.Clientset{}, fmt.Errorf("unable to create kubernetes client - %w", err)
}

func homeDir() string {
	return utils.FromEnvironment("HOME", "~")
}

func kubeConfigPath() string {
	return utils.FromEnvironment("KUBECONFIG", filepath.Join(homeDir(), ".kube", "config"))
}
