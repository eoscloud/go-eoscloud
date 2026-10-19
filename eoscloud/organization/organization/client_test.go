package organization_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.eoscloud.io/go-eoscloud/eoscloud/organization/model"
	"go.eoscloud.io/go-eoscloud/eoscloud/organization/organization"
	"go.eoscloud.io/go-eoscloud/pkg/api"
)

func TestNewClient(t *testing.T) {
	c := api.NewClient()
	var _ = organization.NewClient(c, "organization", "/{version}/{kind}/")
}

func TestOrganizationClient(t *testing.T) {
	t.Run("List", func(t *testing.T) {
		var gotPath, gotMethod string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"count":1,"results":[{"id":"org-1","name":"test-org","status":"active"}]}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := organization.NewClient(c, "organization", "/{version}/{kind}/")

		result, err := client.List(context.Background(), nil)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, http.MethodGet, gotMethod)
		assert.Equal(t, "/v1/organizations/", gotPath)
		assert.Equal(t, 1, result.Count)
		require.Len(t, result.Results, 1)
		assert.Equal(t, "org-1", result.Results[0].ID)
		assert.Equal(t, "test-org", result.Results[0].Name)
	})

	t.Run("Get", func(t *testing.T) {
		var gotPath, gotMethod string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id":"org-1","name":"test-org","status":"active"}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := organization.NewClient(c, "organization", "/{version}/{kind}/")

		result, err := client.Get(context.Background(), "org-1", nil)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, http.MethodGet, gotMethod)
		assert.Equal(t, "/v1/organizations/org-1/", gotPath)
		assert.Equal(t, "org-1", result.ID)
		assert.Equal(t, "test-org", result.Name)
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
			_, _ = w.Write([]byte(`{"id":"org-1","name":"new-org"}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := organization.NewClient(c, "organization", "/{version}/{kind}/")

		req := &model.OrganizationRequest{Name: "new-org"}
		result, err := client.Create(context.Background(), req)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, http.MethodPost, gotMethod)
		assert.Equal(t, "/v1/organizations/", gotPath)
		assert.Equal(t, "new-org", gotBody["name"])
		assert.Equal(t, "org-1", result.ID)
		assert.Equal(t, "new-org", result.Name)
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
			_, _ = w.Write([]byte(`{"id":"org-1","name":"updated-org"}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := organization.NewClient(c, "organization", "/{version}/{kind}/")

		name := "updated-org"
		req := &model.PatchedOrganizationRequest{Name: &name}
		result, err := client.UpdatePartial(context.Background(), "org-1", req)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, http.MethodPatch, gotMethod)
		assert.Equal(t, "/v1/organizations/org-1/", gotPath)
		assert.Equal(t, "updated-org", gotBody["name"])
		assert.Equal(t, "org-1", result.ID)
		assert.Equal(t, "updated-org", result.Name)
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
		client := organization.NewClient(c, "organization", "/{version}/{kind}/")

		err := client.Delete(context.Background(), "org-1")
		require.NoError(t, err)

		assert.Equal(t, http.MethodDelete, gotMethod)
		assert.Equal(t, "/v1/organizations/org-1/", gotPath)
	})

	t.Run("Actions", func(t *testing.T) {
		var gotPath, gotMethod string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"name":"organization","description":"","renders":[],"parses":[],"actions":{"available":["delete"],"unavailable":{}}}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := organization.NewClient(c, "organization", "/{version}/{kind}/")

		result, err := client.Actions(context.Background(), "org-1")
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, http.MethodOptions, gotMethod)
		assert.Equal(t, "/v1/organizations/org-1/", gotPath)
		assert.Equal(t, "organization", result.Name)
		require.Len(t, result.Actions.Available, 1)
		assert.Equal(t, "delete", result.Actions.Available[0])
	})

	t.Run("Undelete", func(t *testing.T) {
		var gotPath, gotMethod string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method
			w.WriteHeader(http.StatusNoContent)
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := organization.NewClient(c, "organization", "/{version}/{kind}/")

		err := client.Undelete(context.Background(), "org-1")
		require.NoError(t, err)

		assert.Equal(t, http.MethodPost, gotMethod)
		assert.Equal(t, "/v1/organizations/org-1/:undelete", gotPath)
	})

	t.Run("Restore", func(t *testing.T) {
		var gotPath, gotMethod string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method
			w.WriteHeader(http.StatusNoContent)
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := organization.NewClient(c, "organization", "/{version}/{kind}/")

		err := client.Restore(context.Background(), "org-1")
		require.NoError(t, err)

		assert.Equal(t, http.MethodPost, gotMethod)
		assert.Equal(t, "/v1/organizations/org-1/:restore", gotPath)
	})

	t.Run("Stats", func(t *testing.T) {
		var gotPath, gotMethod string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"members_count":3,"realms_count":2}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := organization.NewClient(c, "organization", "/{version}/{kind}/")

		result, err := client.Stats(context.Background(), "org-1")
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, http.MethodGet, gotMethod)
		assert.Equal(t, "/v1/organizations/org-1/:stats", gotPath)
		assert.Equal(t, float64(3), (*result)["members_count"])
		assert.Equal(t, float64(2), (*result)["realms_count"])
	})

	t.Run("ErrorCase_NotFound", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"detail":"not found"}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := organization.NewClient(c, "organization", "/{version}/{kind}/")

		result, err := client.Get(context.Background(), "nonexistent", nil)
		require.Error(t, err)
		assert.Nil(t, result)
	})
}
