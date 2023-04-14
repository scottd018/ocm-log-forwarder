package elasticsearch

import (
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/olivere/elastic/v7"

	"github.com/scottd018/ocm-log-forwarder/internal/pkg/config"
	"github.com/scottd018/ocm-log-forwarder/internal/pkg/processor"
)

func getAuthTypeBasic(proc *processor.Processor) (*elastic.Client, error) {
	username, password, err := config.GetElasticSearchAuthTypeBasic(proc.KubeClient, proc.Context)
	if err != nil {
		return &elastic.Client{}, fmt.Errorf("unable to configure basic auth type - %w", err)
	}

	tlsConfig, err := getTLSConfig()
	if err != nil {
		return &elastic.Client{}, fmt.Errorf("unable to set tls config - %w", err)
	}

	client, err := elastic.NewClient(
		elastic.SetSniff(false),
		elastic.SetURL(config.GetElasticSearchURL()),
		elastic.SetBasicAuth(username, password),
		elastic.SetHttpClient(
			&http.Client{
				Transport: &http.Transport{
					TLSClientConfig: tlsConfig,
				},
			},
		),
	)

	if err != nil {
		return client, fmt.Errorf("unable to create elasticsearch client - %w", err)
	}

	return client, nil
}

func getTLSConfig() (*tls.Config, error) {
	cert, key, verify := config.GetElasticSearchTLSConnectionInfo()

	// if verify false is explicitly requested, return the connection info
	//nolint: gosec
	if !verify {
		return &tls.Config{InsecureSkipVerify: true, MinVersion: tls.VersionTLS12}, nil
	}

	// load the tls connection info
	connectionCert, err := tls.LoadX509KeyPair(cert, key)
	if err != nil {
		//nolint: gosec
		return &tls.Config{}, fmt.Errorf(
			"unable to load key pair: cert=%s, key=%s - %w",
			cert,
			key,
			err,
		)
	}

	return &tls.Config{Certificates: []tls.Certificate{connectionCert}, MinVersion: tls.VersionTLS12}, nil
}
