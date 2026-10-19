package ip_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.eoscloud.io/go-eoscloud/eoscloud/compute/ip"
	"go.eoscloud.io/go-eoscloud/eoscloud/compute/model"
	"go.eoscloud.io/go-eoscloud/pkg/api"
)

func TestNewClient(t *testing.T) {
	c := api.NewClient()
	var _ = ip.NewClient(c, "compute", "/{version}/realm-1/{kind}/")
}

func TestPublicIPClient(t *testing.T) {
	t.Run("List", func(t *testing.T) {
		var gotPath string
		var gotMethod string

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"count":1,"results":[{"id":"ip-1","ip_family":4,"network_id":"net-1"}]}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := ip.NewClient(c, "compute", "/{version}/realm-1/{kind}/")

		result, err := client.List(context.Background(), nil)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, http.MethodGet, gotMethod)
		assert.Equal(t, "/v1/realm-1/public-ips/", gotPath)
		assert.Equal(t, 1, result.Count)
		require.Len(t, result.Results, 1)
		assert.Equal(t, "ip-1", result.Results[0].ID)
	})

	t.Run("Get", func(t *testing.T) {
		var gotPath string
		var gotMethod string

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id":"ip-1","ip_family":4,"network_id":"net-1"}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := ip.NewClient(c, "compute", "/{version}/realm-1/{kind}/")

		result, err := client.Get(context.Background(), "ip-1", nil)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, http.MethodGet, gotMethod)
		assert.Equal(t, "/v1/realm-1/public-ips/ip-1/", gotPath)
		assert.Equal(t, "ip-1", result.ID)
	})

	t.Run("Create", func(t *testing.T) {
		var gotPath string
		var gotMethod string
		var gotBody map[string]any

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method
			b, _ := io.ReadAll(r.Body)
			_ = r.Body.Close()
			_ = json.Unmarshal(b, &gotBody)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id":"ip-1","ip_family":4,"network_id":"net-1"}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := ip.NewClient(c, "compute", "/{version}/realm-1/{kind}/")

		req := &model.PublicIPRequest{IPFamily: 4, Network: "net-1"}
		result, err := client.Create(context.Background(), req)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, http.MethodPost, gotMethod)
		assert.Equal(t, "/v1/realm-1/public-ips/", gotPath)
		assert.Equal(t, "ip-1", result.ID)
	})

	t.Run("Update", func(t *testing.T) {
		var gotPath string
		var gotMethod string

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id":"ip-1","ip_family":4,"network_id":"net-1","description":"updated"}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := ip.NewClient(c, "compute", "/{version}/realm-1/{kind}/")

		req := &model.PatchedPublicIPRequest{Description: "updated"}
		result, err := client.Update(context.Background(), "ip-1", req)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, http.MethodPut, gotMethod)
		assert.Equal(t, "/v1/realm-1/public-ips/ip-1/", gotPath)
		assert.Equal(t, "ip-1", result.ID)
		assert.Equal(t, "updated", result.Description)
	})

	t.Run("UpdatePartial", func(t *testing.T) {
		var gotPath string
		var gotMethod string

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id":"ip-1","ip_family":4,"network_id":"net-1","description":"patched"}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := ip.NewClient(c, "compute", "/{version}/realm-1/{kind}/")

		req := &model.PatchedPublicIPRequest{Description: "patched"}
		result, err := client.UpdatePartial(context.Background(), "ip-1", req)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, http.MethodPatch, gotMethod)
		assert.Equal(t, "/v1/realm-1/public-ips/ip-1/", gotPath)
		assert.Equal(t, "ip-1", result.ID)
		assert.Equal(t, "patched", result.Description)
	})

	t.Run("Delete", func(t *testing.T) {
		var gotPath string
		var gotMethod string

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method
			w.WriteHeader(http.StatusNoContent)
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := ip.NewClient(c, "compute", "/{version}/realm-1/{kind}/")

		err := client.Delete(context.Background(), "ip-1")
		require.NoError(t, err)

		assert.Equal(t, http.MethodDelete, gotMethod)
		assert.Equal(t, "/v1/realm-1/public-ips/ip-1/", gotPath)
	})

	t.Run("Actions", func(t *testing.T) {
		var gotPath string
		var gotMethod string

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"name":"ip","description":"","renders":[],"parses":[],"actions":{"available":["delete"],"unavailable":{}}}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := ip.NewClient(c, "compute", "/{version}/realm-1/{kind}/")

		result, err := client.Actions(context.Background(), "ip-1")
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, http.MethodOptions, gotMethod)
		assert.Equal(t, "/v1/realm-1/public-ips/ip-1/", gotPath)
		assert.Equal(t, "ip", result.Name)
		require.Len(t, result.Actions.Available, 1)
		assert.Equal(t, "delete", result.Actions.Available[0])
	})

	t.Run("Attach", func(t *testing.T) {
		var gotPath string
		var gotMethod string

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id":"ip-1","ip_family":4,"network_id":"net-1"}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := ip.NewClient(c, "compute", "/{version}/realm-1/{kind}/")

		id := "123"
		input := &model.PublicIPAttachRequest{VM: "vm-1"}
		result, err := client.Attach(context.Background(), &id, input)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, http.MethodPost, gotMethod)
		assert.True(t, strings.Contains(gotPath, ":attach"), "path should contain :attach, got: %s", gotPath)
		assert.Equal(t, "/v1/realm-1/public-ips/123/:attach", gotPath)
		assert.Equal(t, "ip-1", result.ID)
	})

	t.Run("Detach", func(t *testing.T) {
		var gotPath string
		var gotMethod string

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id":"ip-1","ip_family":4,"network_id":"net-1"}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := ip.NewClient(c, "compute", "/{version}/realm-1/{kind}/")

		id := "123"
		result, err := client.Detach(context.Background(), &id)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, http.MethodPost, gotMethod)
		assert.True(t, strings.Contains(gotPath, ":detach"), "path should contain :detach, got: %s", gotPath)
		assert.Equal(t, "/v1/realm-1/public-ips/123/:detach", gotPath)
		assert.Equal(t, "ip-1", result.ID)
	})

	t.Run("AttachVM", func(t *testing.T) {
		var gotPath string
		var gotMethod string

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id":"ip-1","ip_family":4,"network_id":"net-1"}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := ip.NewClient(c, "compute", "/{version}/realm-1/{kind}/")

		id := "123"
		input := &model.PublicIPAttachRequest{VM: "vm-1"}
		result, err := client.AttachVM(context.Background(), &id, input)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, http.MethodPost, gotMethod)
		assert.Equal(t, "/v1/realm-1/public-ips/123/:attach-vm", gotPath)
		assert.Equal(t, "ip-1", result.ID)
	})

	t.Run("DetachVM", func(t *testing.T) {
		var gotPath string
		var gotMethod string

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id":"ip-1","ip_family":4,"network_id":"net-1"}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := ip.NewClient(c, "compute", "/{version}/realm-1/{kind}/")

		id := "123"
		result, err := client.DetachVM(context.Background(), &id)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, http.MethodPost, gotMethod)
		assert.Equal(t, "/v1/realm-1/public-ips/123/:detach-vm", gotPath)
		assert.Equal(t, "ip-1", result.ID)
	})

	t.Run("Stats", func(t *testing.T) {
		var gotPath string
		var gotMethod string

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"public_ip_forwards_total_count":2}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := ip.NewClient(c, "compute", "/{version}/realm-1/{kind}/")

		id := "123"
		result, err := client.Stats(context.Background(), &id)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, http.MethodGet, gotMethod)
		assert.Equal(t, "/v1/realm-1/public-ips/123/:stats", gotPath)
		assert.Equal(t, float64(2), (*result)["public_ip_forwards_total_count"])
	})

	t.Run("ErrorCase_NotFound", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"detail":"not found"}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := ip.NewClient(c, "compute", "/{version}/realm-1/{kind}/")

		result, err := client.Get(context.Background(), "nonexistent", nil)
		require.Error(t, err)
		assert.Nil(t, result)
	})
}
