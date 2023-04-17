package processor

import (
	"fmt"
	"path/filepath"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/context"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/scottd018/ocm-log-forwarder/internal/pkg/config"
	"github.com/scottd018/ocm-log-forwarder/internal/pkg/utils"
)

type Processor struct {
	Config       *config.Config
	Context      context.Context
	ResponseData []byte
	KubeClient   *kubernetes.Clientset
}

func NewProcessor(cfg *config.Config) (*Processor, error) {
	// setup the logger
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if cfg.Debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	// create the processor
	processor := &Processor{
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
	proc.Log(log.Info().Str("cluster", proc.Config.ClusterID), "initializing kubernetes cluster config")
	cfg, err := rest.InClusterConfig()

	if err == nil {
		// create the clientset for the config
		client, clusterErr := kubernetes.NewForConfig(cfg)
		if clusterErr != nil {
			return &kubernetes.Clientset{}, fmt.Errorf("unable to create kubernetes client from in cluster - %w", clusterErr)
		}

		return client, nil
	}

	proc.Log(log.Warn().Str("cluster", proc.Config.ClusterID), "unable to initialize in-cluster config; attempting file initialization")

	kubeConfig := kubeConfigPath()

	proc.Log(log.Info().Str("cluster", proc.Config.ClusterID), "initializing kubernetes file config")
	cfg, err = clientcmd.BuildConfigFromFlags("", kubeConfig)
	if err == nil {
		// create the clientset for the config
		client, fileErr := kubernetes.NewForConfig(cfg)
		if fileErr != nil {
			return &kubernetes.Clientset{}, fmt.Errorf("unable to create kubernetes client from kubeconfig - %w", fileErr)
		}

		return client, nil
	}

	return &kubernetes.Clientset{}, fmt.Errorf("unable to create kubernetes client - %w", err)
}

func (proc Processor) Log(event *zerolog.Event, message string) {
	event.Str("source", "controller").Msg(message)
}

func homeDir() string {
	return utils.FromEnvironment("HOME", "~")
}

func kubeConfigPath() string {
	return utils.FromEnvironment("KUBECONFIG", filepath.Join(homeDir(), ".kube", "config"))
}
