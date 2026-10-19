package volume_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.eoscloud.io/go-eoscloud/eoscloud/compute/model"
	"go.eoscloud.io/go-eoscloud/eoscloud/compute/volume"
	"go.eoscloud.io/go-eoscloud/pkg/api"
)

var volumeDetailJSON = map[string]any{
	"id":       "vol-1",
	"name":     "data",
	"block_id": "b-1",
	"capacity": 10,
}

func newVolumeTestServer(handler http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(handler)
}

func newVolumeClient(srvURL string) volume.APIClient {
	apiClient := api.NewClient(api.WithBaseURL(srvURL))
	return volume.NewClient(apiClient, "compute", "/{version}/realm-1/{kind}/")
}

func errorHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"type":       "error",
			"request_id": "req-1",
			"errors":     []map[string]any{{"path": "/", "kind": "not_found", "message": "not found"}},
		})
	}
}

func TestVolumeClient_List(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var gotPath, gotMethod string

		srv := newVolumeTestServer(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"count":   1,
				"results": []map[string]any{{"id": "vol-1", "name": "data", "block_id": "b-1", "capacity": 10}},
			})
		})
		defer srv.Close()

		c := newVolumeClient(srv.URL)
		result, err := c.List(context.Background(), nil)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, "/v1/realm-1/volumes/", gotPath)
		assert.Equal(t, http.MethodGet, gotMethod)
		assert.Equal(t, 1, result.Count)
		require.Len(t, result.Results, 1)
		assert.Equal(t, "vol-1", result.Results[0].ID)
		assert.Equal(t, "data", result.Results[0].Name)
	})
}

func TestVolumeClient_Get(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var gotPath, gotMethod string

		srv := newVolumeTestServer(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(volumeDetailJSON)
		})
		defer srv.Close()

		c := newVolumeClient(srv.URL)
		result, err := c.Get(context.Background(), "vol-1", nil)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, "/v1/realm-1/volumes/vol-1/", gotPath)
		assert.Equal(t, http.MethodGet, gotMethod)
		assert.Equal(t, "vol-1", result.ID)
		assert.Equal(t, "data", result.Name)
	})
}

func TestVolumeClient_Create(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var gotPath, gotMethod string
		var gotBody map[string]any

		srv := newVolumeTestServer(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method
			body, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(body, &gotBody)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(volumeDetailJSON)
		})
		defer srv.Close()

		c := newVolumeClient(srv.URL)
		input := &model.VolumeRequest{
			Block:    "b-1",
			Name:     "data",
			Capacity: 10,
		}
		result, err := c.Create(context.Background(), input)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, "/v1/realm-1/volumes/", gotPath)
		assert.Equal(t, http.MethodPost, gotMethod)
		assert.Equal(t, "b-1", gotBody["block_id"])
		assert.Equal(t, "vol-1", result.ID)
	})
}

func TestVolumeClient_Update(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var gotPath, gotMethod string

		srv := newVolumeTestServer(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(volumeDetailJSON)
		})
		defer srv.Close()

		c := newVolumeClient(srv.URL)
		input := &model.PatchedVolumeRequest{Name: "updated"}
		result, err := c.Update(context.Background(), "vol-1", input)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, "/v1/realm-1/volumes/vol-1/", gotPath)
		assert.Equal(t, http.MethodPut, gotMethod)
	})
}

func TestVolumeClient_UpdatePartial(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var gotPath, gotMethod string

		srv := newVolumeTestServer(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(volumeDetailJSON)
		})
		defer srv.Close()

		c := newVolumeClient(srv.URL)
		input := &model.PatchedVolumeRequest{Name: "patched"}
		result, err := c.UpdatePartial(context.Background(), "vol-1", input)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, "/v1/realm-1/volumes/vol-1/", gotPath)
		assert.Equal(t, http.MethodPatch, gotMethod)
	})
}

func TestVolumeClient_Delete(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var gotPath, gotMethod string

		srv := newVolumeTestServer(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method
			w.WriteHeader(http.StatusNoContent)
		})
		defer srv.Close()

		c := newVolumeClient(srv.URL)
		err := c.Delete(context.Background(), "vol-1")
		require.NoError(t, err)

		assert.Equal(t, "/v1/realm-1/volumes/vol-1/", gotPath)
		assert.Equal(t, http.MethodDelete, gotMethod)
	})
}

func TestVolumeClient_Actions(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var gotPath, gotMethod string

		srv := newVolumeTestServer(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"name":        "volume",
				"description": "Volume actions",
				"renders":     []string{"application/json"},
				"parses":      []string{"application/json"},
				"actions": map[string]any{
					"available":   []string{"attach", "detach"},
					"unavailable": map[string]any{},
				},
			})
		})
		defer srv.Close()

		c := newVolumeClient(srv.URL)
		result, err := c.Actions(context.Background(), "vol-1")
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, "/v1/realm-1/volumes/vol-1/", gotPath)
		assert.Equal(t, http.MethodOptions, gotMethod)
		assert.Equal(t, "volume", result.Name)
	})
}

func TestVolumeClient_Attach(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var gotPath, gotMethod string
		var gotBody map[string]any

		srv := newVolumeTestServer(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method
			body, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(body, &gotBody)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(volumeDetailJSON)
		})
		defer srv.Close()

		c := newVolumeClient(srv.URL)
		id := "123"
		input := &model.VolumeAttachRequest{VM: "vm-1"}
		result, err := c.Attach(context.Background(), &id, input)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, "/v1/realm-1/volumes/123/:attach", gotPath)
		assert.Equal(t, http.MethodPost, gotMethod)
		assert.Equal(t, "vm-1", gotBody["vm_id"])
		assert.Equal(t, "vol-1", result.ID)
	})
}

func TestVolumeClient_Detach(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var gotPath, gotMethod string

		srv := newVolumeTestServer(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(volumeDetailJSON)
		})
		defer srv.Close()

		c := newVolumeClient(srv.URL)
		id := "123"
		result, err := c.Detach(context.Background(), &id)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, "/v1/realm-1/volumes/123/:detach", gotPath)
		assert.Equal(t, http.MethodPost, gotMethod)
		assert.Equal(t, "vol-1", result.ID)
	})
}

func TestVolumeClient_AttachVolume(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var gotPath, gotMethod string
		var gotBody map[string]any

		srv := newVolumeTestServer(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method
			body, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(body, &gotBody)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(volumeDetailJSON)
		})
		defer srv.Close()

		c := newVolumeClient(srv.URL)
		id := "123"
		input := &model.VolumeAttachRequest{VM: "vm-1"}
		result, err := c.AttachVolume(context.Background(), &id, input)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, "/v1/realm-1/volumes/123/:attach", gotPath)
		assert.Equal(t, http.MethodPost, gotMethod)
		assert.Equal(t, "vm-1", gotBody["vm_id"])
		assert.Equal(t, "vol-1", result.ID)
	})
}

func TestVolumeClient_DetachVolume(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var gotPath, gotMethod string

		srv := newVolumeTestServer(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(volumeDetailJSON)
		})
		defer srv.Close()

		c := newVolumeClient(srv.URL)
		id := "123"
		result, err := c.DetachVolume(context.Background(), &id)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, "/v1/realm-1/volumes/123/:detach", gotPath)
		assert.Equal(t, http.MethodPost, gotMethod)
		assert.Equal(t, "vol-1", result.ID)
	})
}

func TestVolumeClient_Grow(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var gotPath, gotMethod string
		var gotBody map[string]any

		srv := newVolumeTestServer(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method
			body, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(body, &gotBody)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusAccepted)
			_ = json.NewEncoder(w).Encode(volumeDetailJSON)
		})
		defer srv.Close()

		c := newVolumeClient(srv.URL)
		id := "123"
		input := &model.GrowVolumeRequest{Capacity: 20}
		result, err := c.Grow(context.Background(), &id, input)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, "/v1/realm-1/volumes/123/:grow", gotPath)
		assert.Equal(t, http.MethodPost, gotMethod)
		assert.InDelta(t, 20, gotBody["capacity"], 0)
		assert.Equal(t, "vol-1", result.ID)
	})

	t.Run("Error", func(t *testing.T) {
		srv := newVolumeTestServer(errorHandler())
		defer srv.Close()

		c := newVolumeClient(srv.URL)
		id := "123"
		input := &model.GrowVolumeRequest{Capacity: 20}
		result, err := c.Grow(context.Background(), &id, input)
		require.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestVolumeClient_GrowVolume(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var gotPath, gotMethod string
		var gotBody map[string]any

		srv := newVolumeTestServer(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method
			body, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(body, &gotBody)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusAccepted)
			_ = json.NewEncoder(w).Encode(volumeDetailJSON)
		})
		defer srv.Close()

		c := newVolumeClient(srv.URL)
		id := "123"
		input := &model.GrowVolumeRequest{Capacity: 20}
		result, err := c.GrowVolume(context.Background(), &id, input)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, "/v1/realm-1/volumes/123/:grow", gotPath)
		assert.Equal(t, http.MethodPost, gotMethod)
		assert.InDelta(t, 20, gotBody["capacity"], 0)
		assert.Equal(t, "vol-1", result.ID)
	})

	t.Run("Error", func(t *testing.T) {
		srv := newVolumeTestServer(errorHandler())
		defer srv.Close()

		c := newVolumeClient(srv.URL)
		id := "123"
		input := &model.GrowVolumeRequest{Capacity: 20}
		result, err := c.GrowVolume(context.Background(), &id, input)
		require.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestVolumeClient_Stats(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var gotPath, gotMethod string

		srv := newVolumeTestServer(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			// The volumes stats endpoint is currently a stub on the server
			// side and returns an empty object.
			_, _ = w.Write([]byte(`{}`))
		})
		defer srv.Close()

		c := newVolumeClient(srv.URL)
		id := "123"
		result, err := c.Stats(context.Background(), &id)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, "/v1/realm-1/volumes/123/:stats", gotPath)
		assert.Equal(t, http.MethodGet, gotMethod)
		assert.Empty(t, *result)
	})
}

func TestVolumeClient_Error(t *testing.T) {
	t.Run("404_ReturnsError", func(t *testing.T) {
		srv := newVolumeTestServer(errorHandler())
		defer srv.Close()

		c := newVolumeClient(srv.URL)
		result, err := c.List(context.Background(), nil)
		require.Error(t, err)
		assert.Nil(t, result)
	})
}
