package clientcredential_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.eoscloud.io/go-eoscloud/eoscloud/organization/clientcredential"
	"go.eoscloud.io/go-eoscloud/eoscloud/organization/model"
	"go.eoscloud.io/go-eoscloud/pkg/api"
)

func TestNewClient(t *testing.T) {
	c := api.NewClient()
	var _ = clientcredential.NewClient(c, "organization", "/{version}/{kind}/")
}

func TestClientCredentialClient(t *testing.T) {
	t.Run("List", func(t *testing.T) {
		var gotPath, gotMethod string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"count":1,"results":[{"id":"cc-1","name":"test-cc","assignment_id":"org-1","assignment_type":"organization"}]}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := clientcredential.NewClient(c, "organization", "/{version}/{kind}/")

		result, err := client.List(context.Background(), nil)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, http.MethodGet, gotMethod)
		assert.Equal(t, "/v1/client-credentials/", gotPath)
		assert.Equal(t, 1, result.Count)
		require.Len(t, result.Results, 1)
		assert.Equal(t, "cc-1", result.Results[0].ID)
		assert.Equal(t, "org-1", result.Results[0].AssignmentID)
	})

	t.Run("Get", func(t *testing.T) {
		var gotPath, gotMethod string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id":"cc-1","name":"test-cc","assignment_id":"org-1","assignment_type":"organization"}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := clientcredential.NewClient(c, "organization", "/{version}/{kind}/")

		result, err := client.Get(context.Background(), "cc-1", nil)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, http.MethodGet, gotMethod)
		assert.Equal(t, "/v1/client-credentials/cc-1/", gotPath)
		assert.Equal(t, "cc-1", result.ID)
	})

	t.Run("Create", func(t *testing.T) {
		var gotPath, gotMethod string
		var gotBody map[string]any
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method
			b, _ := io.ReadAll(r.Body)
			_ = r.Body.Close()
			_ = json.Unmarshal(b, &gotBody)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id":"cc-1","name":"new-cc","assignment_id":"org-1","assignment_type":"organization","secret":"s3cr3t"}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := clientcredential.NewClient(c, "organization", "/{version}/{kind}/")

		req := &model.ClientCredentialRequest{Name: "new-cc", AssignmentID: "org-1", AssignmentType: "organization"}
		result, err := client.Create(context.Background(), req)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, http.MethodPost, gotMethod)
		assert.Equal(t, "/v1/client-credentials/", gotPath)
		assert.Equal(t, "new-cc", gotBody["name"])
		assert.Equal(t, "org-1", gotBody["assignment_id"])
		assert.Equal(t, "cc-1", result.ID)
		require.NotNil(t, result.Secret)
		assert.Equal(t, "s3cr3t", *result.Secret)
	})

	t.Run("UpdatePartial", func(t *testing.T) {
		var gotPath, gotMethod string
		var gotBody map[string]any
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method
			b, _ := io.ReadAll(r.Body)
			_ = r.Body.Close()
			_ = json.Unmarshal(b, &gotBody)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id":"cc-1","name":"updated-cc"}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := clientcredential.NewClient(c, "organization", "/{version}/{kind}/")

		name := "updated-cc"
		req := &model.PatchedClientCredentialRequest{Name: &name}
		result, err := client.UpdatePartial(context.Background(), "cc-1", req)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, http.MethodPatch, gotMethod)
		assert.Equal(t, "/v1/client-credentials/cc-1/", gotPath)
		assert.Equal(t, "updated-cc", gotBody["name"])
		assert.Equal(t, "cc-1", result.ID)
	})

	t.Run("Delete", func(t *testing.T) {
		var gotPath, gotMethod string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method
			w.WriteHeader(http.StatusNoContent)
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := clientcredential.NewClient(c, "organization", "/{version}/{kind}/")

		err := client.Delete(context.Background(), "cc-1")
		require.NoError(t, err)

		assert.Equal(t, http.MethodDelete, gotMethod)
		assert.Equal(t, "/v1/client-credentials/cc-1/", gotPath)
	})

	t.Run("Actions", func(t *testing.T) {
		var gotPath, gotMethod string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"name":"client-credential","description":"","renders":[],"parses":[],"actions":{"available":["delete"],"unavailable":{}}}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := clientcredential.NewClient(c, "organization", "/{version}/{kind}/")

		result, err := client.Actions(context.Background(), "cc-1")
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, http.MethodOptions, gotMethod)
		assert.Equal(t, "/v1/client-credentials/cc-1/", gotPath)
		assert.Equal(t, "client-credential", result.Name)
		require.Len(t, result.Actions.Available, 1)
		assert.Equal(t, "delete", result.Actions.Available[0])
	})

	t.Run("ErrorCase_NotFound", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"detail":"not found"}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := clientcredential.NewClient(c, "organization", "/{version}/{kind}/")

		result, err := client.Get(context.Background(), "nonexistent", nil)
		require.Error(t, err)
		assert.Nil(t, result)
	})
}
