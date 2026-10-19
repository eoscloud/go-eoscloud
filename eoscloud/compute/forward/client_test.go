package forward_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.eoscloud.io/go-eoscloud/eoscloud/compute/forward"
	"go.eoscloud.io/go-eoscloud/eoscloud/compute/model"
	"go.eoscloud.io/go-eoscloud/pkg/api"
)

func TestNewClient(t *testing.T) {
	c := api.NewClient()
	var _ = forward.NewClient(c, "compute", "/{version}/realm-1/{kind}/")
}

func TestForwardClient(t *testing.T) {
	t.Run("List", func(t *testing.T) {
		var gotPath string
		var gotMethod string

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"count":1,"results":[{"id":"fwd-1","network_id":"net-1","target":{"type":"ip","ip":"10.0.0.1"}}]}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := forward.NewClient(c, "compute", "/{version}/realm-1/{kind}/")

		result, err := client.List(context.Background(), nil)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, http.MethodGet, gotMethod)
		assert.Equal(t, "/v1/realm-1/forwards/", gotPath)
		assert.Equal(t, 1, result.Count)
		require.Len(t, result.Results, 1)
		assert.Equal(t, "fwd-1", result.Results[0].ID)
	})

	t.Run("Get", func(t *testing.T) {
		var gotPath string
		var gotMethod string

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id":"fwd-1","network_id":"net-1","target":{"type":"ip","ip":"10.0.0.1"}}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := forward.NewClient(c, "compute", "/{version}/realm-1/{kind}/")

		result, err := client.Get(context.Background(), "fwd-1", nil)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, http.MethodGet, gotMethod)
		assert.Equal(t, "/v1/realm-1/forwards/fwd-1/", gotPath)
		assert.Equal(t, "fwd-1", result.ID)
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
			_, _ = w.Write([]byte(`{"id":"fwd-1","network_id":"net-1","target":{"type":"ip","ip":"10.0.0.1"}}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := forward.NewClient(c, "compute", "/{version}/realm-1/{kind}/")

		req := &model.PortForwardRequest{Network: "net-1", Target: model.NewPortForwardTarget(model.NewPortForwardTargetIP("10.0.0.1"))}
		result, err := client.Create(context.Background(), req)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, http.MethodPost, gotMethod)
		assert.Equal(t, "/v1/realm-1/forwards/", gotPath)
		assert.Equal(t, "net-1", gotBody["network_id"])
		assert.Equal(t, map[string]any{"type": "ip", "ip": "10.0.0.1"}, gotBody["target"])
		assert.Equal(t, "fwd-1", result.ID)
	})

	t.Run("Update", func(t *testing.T) {
		var gotPath string
		var gotMethod string

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id":"fwd-1","network_id":"net-1","target":{"type":"ip","ip":"10.0.0.2"}}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := forward.NewClient(c, "compute", "/{version}/realm-1/{kind}/")

		req := &model.PortForwardRequest{Network: "net-1", ListenIP: "l-1", Target: model.NewPortForwardTarget(model.NewPortForwardTargetVM("vm-2"))}
		result, err := client.Update(context.Background(), "fwd-1", req)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, http.MethodPut, gotMethod)
		assert.Equal(t, "/v1/realm-1/forwards/fwd-1/", gotPath)
		assert.Equal(t, "fwd-1", result.ID)
	})

	t.Run("UpdatePartial", func(t *testing.T) {
		var gotPath string
		var gotMethod string

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id":"fwd-1","network_id":"net-1","target":{"type":"vm","id":"vm-3"}}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := forward.NewClient(c, "compute", "/{version}/realm-1/{kind}/")

		req := &model.PatchedPortForwardRequest{Target: model.NewPortForwardTarget(model.NewPortForwardTargetVM("vm-3"))}
		result, err := client.UpdatePartial(context.Background(), "fwd-1", req)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, http.MethodPatch, gotMethod)
		assert.Equal(t, "/v1/realm-1/forwards/fwd-1/", gotPath)
		assert.Equal(t, "fwd-1", result.ID)
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
		client := forward.NewClient(c, "compute", "/{version}/realm-1/{kind}/")

		err := client.Delete(context.Background(), "fwd-1")
		require.NoError(t, err)

		assert.Equal(t, http.MethodDelete, gotMethod)
		assert.Equal(t, "/v1/realm-1/forwards/fwd-1/", gotPath)
	})

	t.Run("Actions", func(t *testing.T) {
		var gotPath string
		var gotMethod string

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"name":"acl","description":"","renders":[],"parses":[],"actions":{"available":["delete"],"unavailable":{}}}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := forward.NewClient(c, "compute", "/{version}/realm-1/{kind}/")

		result, err := client.Actions(context.Background(), "fwd-1")
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, http.MethodOptions, gotMethod)
		assert.Equal(t, "/v1/realm-1/forwards/fwd-1/", gotPath)
		assert.Equal(t, "acl", result.Name)
		require.Len(t, result.Actions.Available, 1)
		assert.Equal(t, "delete", result.Actions.Available[0])
	})

	t.Run("Stats", func(t *testing.T) {
		var gotPath string
		var gotMethod string

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method

			// The forwards stats endpoint is currently a stub on the server
			// side and returns an empty object.
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := forward.NewClient(c, "compute", "/{version}/realm-1/{kind}/")

		id := "fwd-1"
		result, err := client.Stats(context.Background(), &id)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, http.MethodGet, gotMethod)
		assert.Equal(t, "/v1/realm-1/forwards/fwd-1/:stats", gotPath)
		assert.Empty(t, *result)
	})

	t.Run("ErrorCase_NotFound", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"detail":"not found"}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := forward.NewClient(c, "compute", "/{version}/realm-1/{kind}/")

		result, err := client.Get(context.Background(), "nonexistent", nil)
		require.Error(t, err)
		assert.Nil(t, result)
	})
}
