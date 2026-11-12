package eoscloud

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.eoscloud.io/go-eoscloud/pkg/api"
)

func TestNewClient_AllowsComputeConsoleByDefault(t *testing.T) {
	for name, tc := range map[string]struct {
		opts []api.Option
		want []string
	}{
		"default": {
			want: []string{"compute.eoscloud.io", "compute-console.eoscloud.io"},
		},
		"custom base URL": {
			opts: []api.Option{api.WithBaseURL("https://{service}.eoscloud.dev")},
			want: []string{"compute.eoscloud.dev", "compute-console.eoscloud.dev"},
		},
		"additional subdomain": {
			opts: []api.Option{api.WithWebSocketSubdomains("terminal")},
			want: []string{"compute.eoscloud.io", "compute-console.eoscloud.io", "terminal.eoscloud.io"},
		},
	} {
		t.Run(name, func(t *testing.T) {
			c := NewClient(tc.opts...)
			hosts, err := c.apiClient.(api.WebSocketHostProvider).WebSocketHosts("compute")
			require.NoError(t, err)
			assert.Equal(t, tc.want, hosts)
		})
	}
}
