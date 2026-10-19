package auditlog_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.eoscloud.io/go-eoscloud/eoscloud/auditlog"
	"go.eoscloud.io/go-eoscloud/pkg/api"
)

func TestAuditlogClient_List(t *testing.T) {
	t.Run("Success_path_query_and_results", func(t *testing.T) {
		var gotPath, gotMethod, gotQuery string

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method
			gotQuery = r.URL.RawQuery
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			// Standard "count"/"results" envelope (verified against the live API).
			_ = json.NewEncoder(w).Encode(map[string]any{
				"count": 1,
				"results": []map[string]any{{
					"id":             "evt-1",
					"service":        "compute",
					"resource_id":    "net-123",
					"resource_kind":  "networks",
					"resource_name":  "e2e-net",
					"type":           "create",
					"identity_id":    "user-1",
					"identity_name":  "Jane Doe",
					"identity_email": "jane@example.com",
					"identity_type":  "user",
				}},
			})
		}))
		defer srv.Close()

		apiClient := api.NewClient(api.WithBaseURL(srv.URL))
		c := auditlog.NewClient(apiClient).AuditLog

		filter := map[string][]string{
			"resource_id":   {"net-123"},
			"resource_kind": {"networks"},
			"type":          {"create"},
			"service":       {"compute"},
		}
		result, err := c.List(context.Background(), filter)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, "/v1/auditlogs", gotPath)
		assert.Equal(t, http.MethodGet, gotMethod)
		assert.Contains(t, gotQuery, "resource_id=net-123")
		assert.Contains(t, gotQuery, "resource_kind=networks")
		assert.Contains(t, gotQuery, "type=create")

		assert.Equal(t, 1, result.Count)
		require.Len(t, result.Results, 1)
		e := result.Results[0]
		assert.Equal(t, "evt-1", e.ID)
		assert.Equal(t, "compute", e.Service)
		assert.Equal(t, "net-123", e.ResourceID)
		assert.Equal(t, "networks", e.ResourceKind)
		assert.Equal(t, "create", e.Type)
		assert.Equal(t, "user-1", e.IdentityID)
		assert.Equal(t, "Jane Doe", e.IdentityName)
		assert.Equal(t, "jane@example.com", e.IdentityEmail)
		require.NotNil(t, e.IdentityType)
		assert.Equal(t, auditlog.IdentityTypeUser, *e.IdentityType)
	})

	t.Run("Decodes_client_identity_type", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"count":1,"results":[{"id":"evt-2","identity_id":"cc-1","identity_type":"client"}]}`))
		}))
		defer srv.Close()

		c := auditlog.NewClient(api.NewClient(api.WithBaseURL(srv.URL))).AuditLog
		result, err := c.List(context.Background(), nil)
		require.NoError(t, err)

		require.Len(t, result.Results, 1)
		require.NotNil(t, result.Results[0].IdentityType)
		assert.Equal(t, auditlog.IdentityTypeClient, *result.Results[0].IdentityType)
	})

	t.Run("Get", func(t *testing.T) {
		var gotPath, gotMethod string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id":"evt-1","service":"compute","type":"create"}`))
		}))
		defer srv.Close()

		apiClient := api.NewClient(api.WithBaseURL(srv.URL))
		c := auditlog.NewClient(apiClient).AuditLog

		result, err := c.Get(context.Background(), "evt-1")
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, http.MethodGet, gotMethod)
		assert.Equal(t, "/v1/auditlogs/evt-1", gotPath)
		assert.Equal(t, "evt-1", result.ID)
	})

	t.Run("Error_404", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
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
		c := auditlog.NewClient(apiClient).AuditLog

		result, err := c.List(context.Background(), nil)
		require.Error(t, err)
		assert.Nil(t, result)
	})
}
