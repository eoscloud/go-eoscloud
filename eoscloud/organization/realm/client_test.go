package realm_test

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
	"go.eoscloud.io/go-eoscloud/eoscloud/organization/realm"
	"go.eoscloud.io/go-eoscloud/pkg/api"
)

func TestNewClient(t *testing.T) {
	c := api.NewClient()
	var _ = realm.NewClient(c, "organization", "/{version}/{kind}/")
}

func TestRealmClient(t *testing.T) {
	t.Run("List", func(t *testing.T) {
		var gotPath, gotMethod string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"count":1,"results":[{"id":"realm-1","name":"test-realm","organization_id":"org-1","status":"active"}]}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := realm.NewClient(c, "organization", "/{version}/{kind}/")

		result, err := client.List(context.Background(), nil)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, http.MethodGet, gotMethod)
		assert.Equal(t, "/v1/realms/", gotPath)
		assert.Equal(t, 1, result.Count)
		require.Len(t, result.Results, 1)
		assert.Equal(t, "realm-1", result.Results[0].ID)
		assert.Equal(t, "test-realm", result.Results[0].Name)
		assert.Equal(t, "org-1", result.Results[0].OrganizationID)
	})

	t.Run("Get", func(t *testing.T) {
		var gotPath, gotMethod string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id":"realm-1","name":"test-realm","organization_id":"org-1","status":"active"}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := realm.NewClient(c, "organization", "/{version}/{kind}/")

		result, err := client.Get(context.Background(), "realm-1", nil)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, http.MethodGet, gotMethod)
		assert.Equal(t, "/v1/realms/realm-1/", gotPath)
		assert.Equal(t, "realm-1", result.ID)
		assert.Equal(t, "test-realm", result.Name)
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
			_, _ = w.Write([]byte(`{"id":"realm-1","name":"new-realm","organization_id":"org-1","subject":"sub-1"}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := realm.NewClient(c, "organization", "/{version}/{kind}/")

		req := &model.RealmRequest{Name: "new-realm", OrganizationID: "org-1"}
		result, err := client.Create(context.Background(), req)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, http.MethodPost, gotMethod)
		assert.Equal(t, "/v1/realms/", gotPath)
		assert.Equal(t, "new-realm", gotBody["name"])
		assert.Equal(t, "org-1", gotBody["organization_id"])
		assert.Equal(t, "realm-1", result.ID)
		assert.Equal(t, "new-realm", result.Name)
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
			_, _ = w.Write([]byte(`{"id":"realm-1","name":"updated-realm"}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := realm.NewClient(c, "organization", "/{version}/{kind}/")

		name := "updated-realm"
		req := &model.PatchedRealmRequest{Name: &name}
		result, err := client.UpdatePartial(context.Background(), "realm-1", req)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, http.MethodPatch, gotMethod)
		assert.Equal(t, "/v1/realms/realm-1/", gotPath)
		assert.Equal(t, "updated-realm", gotBody["name"])
		assert.Equal(t, "realm-1", result.ID)
		assert.Equal(t, "updated-realm", result.Name)
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
		client := realm.NewClient(c, "organization", "/{version}/{kind}/")

		err := client.Delete(context.Background(), "realm-1")
		require.NoError(t, err)

		assert.Equal(t, http.MethodDelete, gotMethod)
		assert.Equal(t, "/v1/realms/realm-1/", gotPath)
	})

	t.Run("Actions", func(t *testing.T) {
		var gotPath, gotMethod string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"name":"realm","description":"","renders":[],"parses":[],"actions":{"available":["delete"],"unavailable":{}}}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := realm.NewClient(c, "organization", "/{version}/{kind}/")

		result, err := client.Actions(context.Background(), "realm-1")
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, http.MethodOptions, gotMethod)
		assert.Equal(t, "/v1/realms/realm-1/", gotPath)
		assert.Equal(t, "realm", result.Name)
		require.Len(t, result.Actions.Available, 1)
		assert.Equal(t, "delete", result.Actions.Available[0])
	})

	t.Run("Suspend", func(t *testing.T) {
		var gotPath, gotMethod string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method
			w.WriteHeader(http.StatusNoContent)
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := realm.NewClient(c, "organization", "/{version}/{kind}/")

		err := client.Suspend(context.Background(), "realm-1")
		require.NoError(t, err)

		assert.Equal(t, http.MethodPost, gotMethod)
		assert.Equal(t, "/v1/realms/realm-1/:suspend", gotPath)
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
		client := realm.NewClient(c, "organization", "/{version}/{kind}/")

		err := client.Restore(context.Background(), "realm-1")
		require.NoError(t, err)

		assert.Equal(t, http.MethodPost, gotMethod)
		assert.Equal(t, "/v1/realms/realm-1/:restore", gotPath)
	})

	t.Run("CancelRestore", func(t *testing.T) {
		var gotPath, gotMethod string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method
			w.WriteHeader(http.StatusNoContent)
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := realm.NewClient(c, "organization", "/{version}/{kind}/")

		err := client.CancelRestore(context.Background(), "realm-1")
		require.NoError(t, err)

		assert.Equal(t, http.MethodPost, gotMethod)
		assert.Equal(t, "/v1/realms/realm-1/:cancel-restore", gotPath)
	})

	t.Run("Stats", func(t *testing.T) {
		var gotPath, gotMethod string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"members_count":4,"client_credentials_count":1}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := realm.NewClient(c, "organization", "/{version}/{kind}/")

		result, err := client.Stats(context.Background(), "realm-1")
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, http.MethodGet, gotMethod)
		assert.Equal(t, "/v1/realms/realm-1/:stats", gotPath)
		assert.Equal(t, float64(4), (*result)["members_count"])
		assert.Equal(t, float64(1), (*result)["client_credentials_count"])
	})

	t.Run("ErrorCase_NotFound", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"detail":"not found"}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := realm.NewClient(c, "organization", "/{version}/{kind}/")

		result, err := client.Get(context.Background(), "nonexistent", nil)
		require.Error(t, err)
		assert.Nil(t, result)
	})
}
