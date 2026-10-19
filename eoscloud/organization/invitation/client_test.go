package invitation_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.eoscloud.io/go-eoscloud/eoscloud/organization/invitation"
	"go.eoscloud.io/go-eoscloud/eoscloud/organization/model"
	"go.eoscloud.io/go-eoscloud/pkg/api"
)

func TestNewClient(t *testing.T) {
	c := api.NewClient()
	var _ = invitation.NewClient(c, "organization", "/{version}/{kind}/")
}

func TestInvitationClient(t *testing.T) {
	t.Run("List", func(t *testing.T) {
		var gotPath, gotMethod string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"count":1,"results":[{"id":"inv-1","email":"user@example.com","organization_id":"org-1","status":"pending"}]}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := invitation.NewClient(c, "organization", "/{version}/{kind}/")

		result, err := client.List(context.Background(), nil)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, http.MethodGet, gotMethod)
		assert.Equal(t, "/v1/invitations/", gotPath)
		assert.Equal(t, 1, result.Count)
		require.Len(t, result.Results, 1)
		assert.Equal(t, "inv-1", result.Results[0].ID)
		require.NotNil(t, result.Results[0].Email)
		assert.Equal(t, "user@example.com", *result.Results[0].Email)
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
			_, _ = w.Write([]byte(`{"id":"inv-1","email":"user@example.com","organization_id":"org-1","roles":["member"]}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := invitation.NewClient(c, "organization", "/{version}/{kind}/")

		req := &model.InvitationRequest{Email: "user@example.com", OrganizationID: "org-1", Roles: []string{"member"}}
		result, err := client.Create(context.Background(), req)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, http.MethodPost, gotMethod)
		assert.Equal(t, "/v1/invitations/", gotPath)
		assert.Equal(t, "user@example.com", gotBody["email"])
		assert.Equal(t, "org-1", gotBody["organization_id"])
		assert.Equal(t, "inv-1", result.ID)
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
		client := invitation.NewClient(c, "organization", "/{version}/{kind}/")

		err := client.Delete(context.Background(), "inv-1")
		require.NoError(t, err)

		assert.Equal(t, http.MethodDelete, gotMethod)
		assert.Equal(t, "/v1/invitations/inv-1/", gotPath)
	})

	t.Run("Actions", func(t *testing.T) {
		var gotPath, gotMethod string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"name":"invitation","description":"","renders":[],"parses":[],"actions":{"available":["create"],"unavailable":{}}}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := invitation.NewClient(c, "organization", "/{version}/{kind}/")

		result, err := client.Actions(context.Background(), "inv-1")
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, http.MethodOptions, gotMethod)
		assert.Equal(t, "/v1/invitations/inv-1/", gotPath)
		assert.Equal(t, "invitation", result.Name)
		require.Len(t, result.Actions.Available, 1)
		assert.Equal(t, "create", result.Actions.Available[0])
	})

	t.Run("ErrorCase_NotFound", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"detail":"not found"}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := invitation.NewClient(c, "organization", "/{version}/{kind}/")

		err := client.Delete(context.Background(), "nonexistent")
		require.Error(t, err)
	})
}
