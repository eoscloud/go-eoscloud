package compute_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.eoscloud.io/go-eoscloud/eoscloud/compute"
	"go.eoscloud.io/go-eoscloud/pkg/api"
)

func TestNewClient(t *testing.T) {
	c := api.NewClient()
	_ = compute.NewClient(c, "/{version}/{kind}/", "/{version}/realm-1/{kind}/")
}

func TestClient_GetRealmStatistics(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var gotPath, gotMethod string

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"vms_total_count":5,"volumes_total_count":10,"networks_total_count":8}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := compute.NewClient(c, "/{version}/{kind}/", "/{version}/realm-1/{kind}/")

		result, err := client.Stats(context.Background())
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, http.MethodGet, gotMethod)
		assert.Equal(t, "/v1/realm-1/:stats", gotPath)
		assert.Equal(t, float64(5), (*result)["vms_total_count"])
		assert.Equal(t, float64(10), (*result)["volumes_total_count"])
		assert.Equal(t, float64(8), (*result)["networks_total_count"])
	})

	t.Run("ErrorCase_NotFound", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"detail":"not found"}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := compute.NewClient(c, "/{version}/{kind}/", "/{version}/realm-1/{kind}/")

		result, err := client.Stats(context.Background())
		require.Error(t, err)
		assert.Nil(t, result)
	})
}
