package elasticsearch

import (
	"crypto/tls"
	"os"
	"testing"

	"github.com/scottd018/ocm-log-forwarder/internal/pkg/config"
)

//nolint:paralleltest
func Test_getTLSConfig(t *testing.T) {
	tests := []struct {
		name    string
		want    *tls.Config
		wantErr bool
		env     map[string]string
	}{
		{
			name:    "ensure ssl verify false takes precedence if requested",
			wantErr: false,
			env: map[string]string{
				config.DefaultEnvironmentBackendElasticTLSVerify: "false",
			},
		},
		{
			name:    "ensure a bad cert configuration returns an error",
			wantErr: true,
			env: map[string]string{
				config.DefaultEnvironmentBackendElasticTLSVerify:      "true",
				config.DefaultEnvironmentBackendElasticTLSCertificate: "/etc/pki/exist.crt",
				config.DefaultEnvironmentBackendElasticTLSKey:         "/etc/pki/exist.key",
			},
		},
		{
			name:    "ensure a good cert configuration returns as expected",
			wantErr: false,
			env: map[string]string{
				config.DefaultEnvironmentBackendElasticTLSVerify:      "true",
				config.DefaultEnvironmentBackendElasticTLSCertificate: "../../../../test/certs/fake-tls.crt",
				config.DefaultEnvironmentBackendElasticTLSKey:         "../../../../test/certs/fake-tls.key",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.env {
				os.Setenv(k, v)
			}

			_, err := getTLSConfig()
			if (err != nil) != tt.wantErr {
				t.Errorf("getTLSConfig() error = %v, wantErr %v", err, tt.wantErr)

				return
			}
		})
	}
}
