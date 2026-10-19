package district_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.eoscloud.io/go-eoscloud/eoscloud/object_storage/district"
	"go.eoscloud.io/go-eoscloud/pkg/api"
)

func TestDistrictClient_List(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var gotPath, gotMethod string

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"count": 1,
				"results": []map[string]any{{
					"id":                "district-1",
					"region_id":         "region-1",
					"slug":              "at-vie-1",
					"block_storage_url": "https://block-storage.example.com",
				}},
			})
		}))
		defer srv.Close()

		apiClient := api.NewClient(api.WithBaseURL(srv.URL))
		c := district.NewClient(apiClient, "object-storage", "/{version}/{kind}/")

		result, err := c.List(context.Background(), nil)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, "/v1/districts/", gotPath)
		assert.Equal(t, http.MethodGet, gotMethod)
		assert.Equal(t, 1, result.Count)
		require.Len(t, result.Results, 1)
		assert.Equal(t, "district-1", result.Results[0].ID)
		assert.Equal(t, "region-1", result.Results[0].RegionID)
		assert.Equal(t, "at-vie-1", result.Results[0].Slug)
		assert.Equal(t, "https://block-storage.example.com", result.Results[0].BlockStorageURL)
	})

	t.Run("Error_404", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"type":       "error",
				"request_id": "req-1",
				"errors":     []map[string]any{{"path": "/", "kind": "not_found", "message": "not found"}},
			})
		}))
		defer srv.Close()

		apiClient := api.NewClient(api.WithBaseURL(srv.URL))
		c := district.NewClient(apiClient, "object-storage", "/{version}/{kind}/")

		result, err := c.List(context.Background(), nil)
		require.Error(t, err)
		assert.Nil(t, result)
	})
}
