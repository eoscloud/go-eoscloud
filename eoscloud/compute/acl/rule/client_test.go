package rule_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.eoscloud.io/go-eoscloud/eoscloud/compute/acl/rule"
	"go.eoscloud.io/go-eoscloud/eoscloud/compute/model"
	"go.eoscloud.io/go-eoscloud/pkg/api"
)

func TestNewClient(t *testing.T) {
	c := api.NewClient()
	var _ = rule.NewClient(c, "compute", "/{version}/realm-1/{kind}/")
}

func TestACLRuleClient(t *testing.T) {
	t.Run("List", func(t *testing.T) {
		var gotPath string
		var gotMethod string

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"count":1,"results":[{"id":"rule-1","acl_id":"acl-1"}]}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := rule.NewClient(c, "compute", "/{version}/realm-1/{kind}/")

		result, err := client.List(context.Background(), nil)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, http.MethodGet, gotMethod)
		assert.Equal(t, "/v1/realm-1/acl-rules/", gotPath)
		assert.Equal(t, 1, result.Count)
		require.Len(t, result.Results, 1)
		assert.Equal(t, "rule-1", result.Results[0].ID)
		assert.Equal(t, "acl-1", result.Results[0].ACL)
	})

	t.Run("Get", func(t *testing.T) {
		var gotPath string
		var gotMethod string

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id":"rule-1","acl_id":"acl-1"}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := rule.NewClient(c, "compute", "/{version}/realm-1/{kind}/")

		result, err := client.Get(context.Background(), "rule-1", nil)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, http.MethodGet, gotMethod)
		assert.Equal(t, "/v1/realm-1/acl-rules/rule-1/", gotPath)
		assert.Equal(t, "rule-1", result.ID)
		assert.Equal(t, "acl-1", result.ACL)
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
			_, _ = w.Write([]byte(`{"id":"rule-1","acl_id":"acl-1"}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := rule.NewClient(c, "compute", "/{version}/realm-1/{kind}/")

		src := []model.ACLSelectorValue{
			model.NewACLSelector(model.NewACLSelectorIP("10.0.0.0/24")),
			model.NewACLSelector(model.NewACLSelectorGroup("grp-1")),
		}
		req := &model.ACLRuleRequest{ACL: "acl-1", Sources: &src}
		result, err := client.Create(context.Background(), req)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, http.MethodPost, gotMethod)
		assert.Equal(t, "/v1/realm-1/acl-rules/", gotPath)
		assert.Equal(t, "acl-1", gotBody["acl_id"])
		assert.Equal(t, []any{
			map[string]any{"type": "ip", "ip": "10.0.0.0/24"},
			map[string]any{"type": "group", "id": "grp-1"},
		}, gotBody["sources"])
		assert.Equal(t, "rule-1", result.ID)
		assert.Equal(t, "acl-1", result.ACL)
	})

	t.Run("Update", func(t *testing.T) {
		var gotPath string
		var gotMethod string

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id":"rule-1","acl_id":"acl-1"}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := rule.NewClient(c, "compute", "/{version}/realm-1/{kind}/")

		req := &model.ACLRuleRequest{ACL: "acl-1"}
		result, err := client.Update(context.Background(), "rule-1", req)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, http.MethodPut, gotMethod)
		assert.Equal(t, "/v1/realm-1/acl-rules/rule-1/", gotPath)
		assert.Equal(t, "rule-1", result.ID)
	})

	t.Run("UpdatePartial", func(t *testing.T) {
		var gotPath string
		var gotMethod string

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id":"rule-1","acl_id":"acl-1"}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := rule.NewClient(c, "compute", "/{version}/realm-1/{kind}/")

		req := &model.PatchedACLRuleRequest{Name: "n"}
		result, err := client.UpdatePartial(context.Background(), "rule-1", req)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, http.MethodPatch, gotMethod)
		assert.Equal(t, "/v1/realm-1/acl-rules/rule-1/", gotPath)
		assert.Equal(t, "rule-1", result.ID)
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
		client := rule.NewClient(c, "compute", "/{version}/realm-1/{kind}/")

		err := client.Delete(context.Background(), "rule-1")
		require.NoError(t, err)

		assert.Equal(t, http.MethodDelete, gotMethod)
		assert.Equal(t, "/v1/realm-1/acl-rules/rule-1/", gotPath)
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
		client := rule.NewClient(c, "compute", "/{version}/realm-1/{kind}/")

		result, err := client.Actions(context.Background(), "rule-1")
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, http.MethodOptions, gotMethod)
		assert.Equal(t, "/v1/realm-1/acl-rules/rule-1/", gotPath)
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

			// The acl-rules stats endpoint is currently a stub on the server
			// side and returns an empty object.
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := rule.NewClient(c, "compute", "/{version}/realm-1/{kind}/")

		id := "rule-1"
		result, err := client.Stats(context.Background(), &id)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, http.MethodGet, gotMethod)
		assert.Equal(t, "/v1/realm-1/acl-rules/rule-1/:stats", gotPath)
		assert.Empty(t, *result)
	})

	t.Run("ErrorCase_NotFound", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"detail":"not found"}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := rule.NewClient(c, "compute", "/{version}/realm-1/{kind}/")

		result, err := client.Get(context.Background(), "nonexistent", nil)
		require.Error(t, err)
		assert.Nil(t, result)
	})
}
