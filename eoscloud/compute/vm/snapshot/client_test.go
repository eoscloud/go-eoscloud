package snapshot_test

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

	"go.eoscloud.io/go-eoscloud/eoscloud/compute/model"
	"go.eoscloud.io/go-eoscloud/eoscloud/compute/vm/snapshot"
	"go.eoscloud.io/go-eoscloud/pkg/api"
)

func TestNewClient(t *testing.T) {
	c := api.NewClient()
	var _ = snapshot.NewClient(c, "compute", "/{version}/realm-1/{kind}/")
}

func TestSnapshotClient(t *testing.T) {
	t.Run("List", func(t *testing.T) {
		var gotPath string
		var gotMethod string

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"count":1,"results":[{"id":"snap-1","vm_id":"vm-1","name":"snap","description":"test snapshot"}]}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := snapshot.NewClient(c, "compute", "/{version}/realm-1/{kind}/")

		result, err := client.List(context.Background(), nil)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, http.MethodGet, gotMethod)
		assert.Equal(t, "/v1/realm-1/snapshots/", gotPath)
		assert.Equal(t, 1, result.Count)
		require.Len(t, result.Results, 1)
		assert.Equal(t, "snap-1", result.Results[0].ID)
		assert.Equal(t, "vm-1", result.Results[0].VM)
		assert.Equal(t, "snap", result.Results[0].Name)
		assert.Equal(t, "test snapshot", result.Results[0].Description)
	})

	t.Run("Get", func(t *testing.T) {
		var gotPath string
		var gotMethod string

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id":"snap-1","vm_id":"vm-1","name":"snap","description":"test snapshot"}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := snapshot.NewClient(c, "compute", "/{version}/realm-1/{kind}/")

		result, err := client.Get(context.Background(), "snap-1", nil)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, http.MethodGet, gotMethod)
		assert.Equal(t, "/v1/realm-1/snapshots/snap-1/", gotPath)
		assert.Equal(t, "snap-1", result.ID)
		assert.Equal(t, "vm-1", result.VM)
		assert.Equal(t, "snap", result.Name)
		assert.Equal(t, "test snapshot", result.Description)
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
			_, _ = w.Write([]byte(`{"id":"snap-1","vm_id":"vm-1","name":"snap","description":"test snapshot"}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := snapshot.NewClient(c, "compute", "/{version}/realm-1/{kind}/")

		req := &model.SnapshotRequest{VM: "vm-1", Name: "snap", Description: "test snapshot"}
		result, err := client.Create(context.Background(), req)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, http.MethodPost, gotMethod)
		assert.Equal(t, "/v1/realm-1/snapshots/", gotPath)
		assert.Equal(t, "snap-1", result.ID)
		assert.Equal(t, "test snapshot", result.Description)
	})

	t.Run("Update", func(t *testing.T) {
		var gotPath string
		var gotMethod string

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id":"snap-1","vm_id":"vm-1","name":"updated","description":"updated desc"}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := snapshot.NewClient(c, "compute", "/{version}/realm-1/{kind}/")

		input := &model.PatchedSnapshotRequest{Name: "updated", Description: "updated desc"}
		result, err := client.Update(context.Background(), "snap-1", input)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, http.MethodPut, gotMethod)
		assert.Equal(t, "/v1/realm-1/snapshots/snap-1/", gotPath)
		assert.Equal(t, "snap-1", result.ID)
		assert.Equal(t, "updated", result.Name)
		assert.Equal(t, "updated desc", result.Description)
	})

	t.Run("UpdatePartial", func(t *testing.T) {
		var gotPath string
		var gotMethod string

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id":"snap-1","vm_id":"vm-1","name":"snap","description":"patched desc"}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := snapshot.NewClient(c, "compute", "/{version}/realm-1/{kind}/")

		input := &model.PatchedSnapshotRequest{Description: "patched desc"}
		result, err := client.UpdatePartial(context.Background(), "snap-1", input)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, http.MethodPatch, gotMethod)
		assert.Equal(t, "/v1/realm-1/snapshots/snap-1/", gotPath)
		assert.Equal(t, "snap-1", result.ID)
		assert.Equal(t, "patched desc", result.Description)
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
		client := snapshot.NewClient(c, "compute", "/{version}/realm-1/{kind}/")

		err := client.Delete(context.Background(), "snap-1")
		require.NoError(t, err)

		assert.Equal(t, http.MethodDelete, gotMethod)
		assert.Equal(t, "/v1/realm-1/snapshots/snap-1/", gotPath)
	})

	t.Run("Actions", func(t *testing.T) {
		var gotPath string
		var gotMethod string

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"name":"snapshot","description":"","renders":[],"parses":[],"actions":{"available":["delete"],"unavailable":{}}}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := snapshot.NewClient(c, "compute", "/{version}/realm-1/{kind}/")

		result, err := client.Actions(context.Background(), "snap-1")
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, http.MethodOptions, gotMethod)
		assert.Equal(t, "/v1/realm-1/snapshots/snap-1/", gotPath)
		assert.Equal(t, "snapshot", result.Name)
	})

	t.Run("Restore", func(t *testing.T) {
		var gotPath string
		var gotMethod string

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusAccepted)
			_, _ = w.Write([]byte(`{"id":"snap-1","vm_id":"vm-1","name":"snap","description":"test snapshot"}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := snapshot.NewClient(c, "compute", "/{version}/realm-1/{kind}/")

		id := "123"
		result, err := client.Restore(context.Background(), &id)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, http.MethodPost, gotMethod)
		assert.True(t, strings.Contains(gotPath, ":restore"), "path should contain :restore, got: %s", gotPath)
		assert.Equal(t, "/v1/realm-1/snapshots/123/:restore", gotPath)
		assert.Equal(t, "snap-1", result.ID)
	})

	t.Run("Stats", func(t *testing.T) {
		var gotPath string
		var gotMethod string

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method

			// The snapshots stats endpoint is currently a stub on the server
			// side and returns an empty object.
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := snapshot.NewClient(c, "compute", "/{version}/realm-1/{kind}/")

		id := "snap-1"
		result, err := client.Stats(context.Background(), &id)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, http.MethodGet, gotMethod)
		assert.Equal(t, "/v1/realm-1/snapshots/snap-1/:stats", gotPath)
		assert.Empty(t, *result)
	})

	t.Run("ErrorCase_NotFound", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"detail":"not found"}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := snapshot.NewClient(c, "compute", "/{version}/realm-1/{kind}/")

		result, err := client.Get(context.Background(), "nonexistent", nil)
		require.Error(t, err)
		assert.Nil(t, result)
	})
}
