package bucket_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.eoscloud.io/go-eoscloud/eoscloud/object_storage/model"
	"go.eoscloud.io/go-eoscloud/eoscloud/object_storage/product/bucket"
	"go.eoscloud.io/go-eoscloud/pkg/api"
)

func TestBucketProductClient_List(t *testing.T) {
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
					"_self":               "/v1/bucket-products/prod-1/",
					"id":                  "prod-1",
					"name":                "small",
					"unit":                "gigabytehours",
					"min_quota_bytes":     1073741824,
					"max_quota_bytes":     5497558138880,
					"default_quota_bytes": 10737418240,
					"_actions":            []string{},
				}},
			})
		}))
		defer srv.Close()

		apiClient := api.NewClient(api.WithBaseURL(srv.URL))
		c := bucket.NewClient(apiClient, "object-storage", "/{version}/{kind}/")

		result, err := c.List(context.Background(), nil)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, "/v1/bucket-products/", gotPath)
		assert.Equal(t, http.MethodGet, gotMethod)
		assert.Equal(t, 1, result.Count)
		require.Len(t, result.Results, 1)
		assert.Equal(t, "prod-1", result.Results[0].ID)
		assert.Equal(t, "small", result.Results[0].Name)
		assert.Equal(t, "/v1/bucket-products/prod-1/", result.Results[0].Self)
		assert.Equal(t, model.UnitGigabyteHours, result.Results[0].Unit)
		assert.Equal(t, int64(1073741824), result.Results[0].MinQuotaBytes)
		assert.Equal(t, int64(5497558138880), result.Results[0].MaxQuotaBytes)
		assert.Equal(t, int64(10737418240), result.Results[0].DefaultQuotaBytes)
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
		c := bucket.NewClient(apiClient, "object-storage", "/{version}/{kind}/")

		result, err := c.List(context.Background(), nil)
		require.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestBucketProductClient_Get(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var gotPath, gotMethod string

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":              "prod-1",
				"name":            "small",
				"description":     "small product",
				"unit":            "gigabytehours",
				"max_quota_bytes": 5497558138880,
			})
		}))
		defer srv.Close()

		apiClient := api.NewClient(api.WithBaseURL(srv.URL))
		c := bucket.NewClient(apiClient, "object-storage", "/{version}/{kind}/")

		result, err := c.Get(context.Background(), "prod-1", nil)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, "/v1/bucket-products/prod-1/", gotPath)
		assert.Equal(t, http.MethodGet, gotMethod)
		assert.Equal(t, "prod-1", result.ID)
		assert.Equal(t, "small", result.Name)
		assert.Equal(t, "small product", result.Description)
		assert.Equal(t, model.UnitGigabyteHours, result.Unit)
		assert.Equal(t, int64(5497558138880), result.MaxQuotaBytes)
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
		c := bucket.NewClient(apiClient, "object-storage", "/{version}/{kind}/")

		result, err := c.Get(context.Background(), "prod-missing", nil)
		require.Error(t, err)
		assert.Nil(t, result)
	})
}
