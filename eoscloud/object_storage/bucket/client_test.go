package bucket_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.eoscloud.io/go-eoscloud/eoscloud/object_storage/bucket"
	"go.eoscloud.io/go-eoscloud/eoscloud/object_storage/model"
	"go.eoscloud.io/go-eoscloud/pkg/api"
)

func TestNewClient(t *testing.T) {
	c := api.NewClient()
	var _ = bucket.NewClient(c, "object-storage", "/{version}/realm-1/{kind}/")
}

func TestBucketClient(t *testing.T) {
	t.Run("List", func(t *testing.T) {
		var gotPath string
		var gotMethod string

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"count":1,"results":[{"id":"bucket-1","name":"test-bucket","quota":100,"district_id":"district-1"}]}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := bucket.NewClient(c, "object-storage", "/{version}/realm-1/{kind}/")

		result, err := client.List(context.Background(), nil)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, http.MethodGet, gotMethod)
		assert.Equal(t, "/v1/realm-1/buckets/", gotPath)
		assert.Equal(t, 1, result.Count)
		require.Len(t, result.Results, 1)
		assert.Equal(t, "bucket-1", result.Results[0].ID)
		assert.Equal(t, "test-bucket", result.Results[0].Name)
		assert.Equal(t, "district-1", result.Results[0].DistrictID)
	})

	t.Run("Get", func(t *testing.T) {
		var gotPath string
		var gotMethod string

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id":"bucket-1","name":"test-bucket","quota":100,"district_id":"district-1"}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := bucket.NewClient(c, "object-storage", "/{version}/realm-1/{kind}/")

		result, err := client.Get(context.Background(), "bucket-1", nil)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, http.MethodGet, gotMethod)
		assert.Equal(t, "/v1/realm-1/buckets/bucket-1/", gotPath)
		assert.Equal(t, "bucket-1", result.ID)
		assert.Equal(t, "test-bucket", result.Name)
		assert.Equal(t, "district-1", result.DistrictID)
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
			_, _ = w.Write([]byte(`{"id":"bucket-1","name":"new-bucket","quota":100,"district_id":"district-1"}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := bucket.NewClient(c, "object-storage", "/{version}/realm-1/{kind}/")

		req := &model.BucketRequest{Name: "new-bucket", Quota: 100, DistrictID: "district-1"}
		result, err := client.Create(context.Background(), req)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, http.MethodPost, gotMethod)
		assert.Equal(t, "/v1/realm-1/buckets/", gotPath)
		assert.Equal(t, "new-bucket", gotBody["name"])
		assert.Equal(t, float64(100), gotBody["quota"])
		assert.Equal(t, "district-1", gotBody["district_id"])
		assert.Equal(t, "bucket-1", result.ID)
		assert.Equal(t, "new-bucket", result.Name)
		assert.Equal(t, "district-1", result.DistrictID)
	})

	t.Run("Update", func(t *testing.T) {
		var gotPath string
		var gotMethod string

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id":"bucket-1","description":"updated-bucket","quota":200}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := bucket.NewClient(c, "object-storage", "/{version}/realm-1/{kind}/")

		desc := "updated-bucket"
		req := &model.PatchedBucketRequest{Description: &desc}
		result, err := client.Update(context.Background(), "bucket-1", req)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, http.MethodPut, gotMethod)
		assert.Equal(t, "/v1/realm-1/buckets/bucket-1/", gotPath)
		assert.Equal(t, "bucket-1", result.ID)
		require.NotNil(t, result.Description)
		assert.Equal(t, "updated-bucket", *result.Description)
	})

	t.Run("UpdatePartial", func(t *testing.T) {
		var gotPath string
		var gotMethod string

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id":"bucket-1","name":"test-bucket","quota":300}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := bucket.NewClient(c, "object-storage", "/{version}/realm-1/{kind}/")

		req := &model.PatchedBucketRequest{}
		result, err := client.UpdatePartial(context.Background(), "bucket-1", req)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, http.MethodPatch, gotMethod)
		assert.Equal(t, "/v1/realm-1/buckets/bucket-1/", gotPath)
		assert.Equal(t, "bucket-1", result.ID)
		assert.Equal(t, 300, result.Quota)
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
		client := bucket.NewClient(c, "object-storage", "/{version}/realm-1/{kind}/")

		err := client.Delete(context.Background(), "bucket-1")
		require.NoError(t, err)

		assert.Equal(t, http.MethodDelete, gotMethod)
		assert.Equal(t, "/v1/realm-1/buckets/bucket-1/", gotPath)
	})

	t.Run("Actions", func(t *testing.T) {
		var gotPath string
		var gotMethod string

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"name":"bucket","description":"","renders":[],"parses":[],"actions":{"available":["delete"],"unavailable":{}}}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := bucket.NewClient(c, "object-storage", "/{version}/realm-1/{kind}/")

		result, err := client.Actions(context.Background(), "bucket-1")
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, http.MethodOptions, gotMethod)
		assert.Equal(t, "/v1/realm-1/buckets/bucket-1/", gotPath)
		assert.Equal(t, "bucket", result.Name)
		require.Len(t, result.Actions.Available, 1)
		assert.Equal(t, "delete", result.Actions.Available[0])
	})

	t.Run("RecoverPolicy", func(t *testing.T) {
		var gotPath string
		var gotMethod string

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusAccepted)
			_, _ = w.Write([]byte(`{"id":"bucket-1","name":"test-bucket","quota":100,"district_id":"district-1"}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := bucket.NewClient(c, "object-storage", "/{version}/realm-1/{kind}/")

		id := "bucket-1"
		result, err := client.RecoverPolicy(context.Background(), &id)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, http.MethodPost, gotMethod)
		assert.Equal(t, "/v1/realm-1/buckets/bucket-1/:recover-policy", gotPath)
		assert.Equal(t, "bucket-1", result.ID)
		assert.Equal(t, "test-bucket", result.Name)
		assert.Equal(t, "district-1", result.DistrictID)
	})

	t.Run("Grow", func(t *testing.T) {
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
			w.WriteHeader(http.StatusAccepted)
			_, _ = w.Write([]byte(`{"id":"bucket-1","name":"test-bucket","quota":500,"district_id":"district-1"}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := bucket.NewClient(c, "object-storage", "/{version}/realm-1/{kind}/")

		id := "bucket-1"
		result, err := client.Grow(context.Background(), &id, &model.BucketGrowRequest{Quota: 500})
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, http.MethodPost, gotMethod)
		assert.Equal(t, "/v1/realm-1/buckets/bucket-1/:grow", gotPath)
		assert.Equal(t, float64(500), gotBody["quota"])
		assert.Equal(t, "bucket-1", result.ID)
		assert.Equal(t, 500, result.Quota)
	})

	t.Run("ErrorCase_NotFound", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"detail":"not found"}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := bucket.NewClient(c, "object-storage", "/{version}/realm-1/{kind}/")

		result, err := client.Get(context.Background(), "nonexistent", nil)
		require.Error(t, err)
		assert.Nil(t, result)
	})
}
