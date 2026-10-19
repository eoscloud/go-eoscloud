package role_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.eoscloud.io/go-eoscloud/eoscloud/organization/role"
	"go.eoscloud.io/go-eoscloud/pkg/api"
)

func TestNewClient(t *testing.T) {
	c := api.NewClient()
	var _ = role.NewClient(c, "organization", "/{version}/{kind}/")
}

func TestRoleClient(t *testing.T) {
	t.Run("List", func(t *testing.T) {
		var gotPath, gotMethod string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"count":1,"previous":null,"next":null,"results":[{"role":"organization:admin","title":"Admin","service":"organization","scope":"organization"}]}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := role.NewClient(c, "organization", "/{version}/{kind}/")

		result, err := client.List(context.Background(), nil)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, http.MethodGet, gotMethod)
		assert.Equal(t, "/v1/roles/", gotPath)
		assert.Equal(t, 1, result.Count)
		require.Len(t, result.Results, 1)
		assert.Equal(t, "organization:admin", result.Results[0].Role)
		assert.Equal(t, "Admin", result.Results[0].Title)
		assert.Equal(t, "organization", result.Results[0].Service)
		assert.Equal(t, "organization", result.Results[0].Scope)
	})

	t.Run("List_ScopeFilter", func(t *testing.T) {
		var gotQuery string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotQuery = r.URL.RawQuery
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"count":0,"previous":null,"next":null,"results":[]}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := role.NewClient(c, "organization", "/{version}/{kind}/")

		_, err := client.List(context.Background(), url.Values{"scope": {"realm"}})
		require.NoError(t, err)

		assert.Contains(t, gotQuery, "scope=realm")
	})

	t.Run("ErrorCase_ServerError", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"detail":"boom"}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := role.NewClient(c, "organization", "/{version}/{kind}/")

		result, err := client.List(context.Background(), nil)
		require.Error(t, err)
		assert.Nil(t, result)
	})
}
