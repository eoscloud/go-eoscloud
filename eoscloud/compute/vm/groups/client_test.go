package groups_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.eoscloud.io/go-eoscloud/eoscloud/compute/model"
	"go.eoscloud.io/go-eoscloud/eoscloud/compute/vm/groups"
	"go.eoscloud.io/go-eoscloud/pkg/api"
)

func TestNewClient(t *testing.T) {
	c := api.NewClient()
	var _ = groups.NewClient(c, "compute", "/{version}/realm-1/{kind}/")
}

func TestVMGroupClient(t *testing.T) {
	t.Run("List", func(t *testing.T) {
		var gotPath string
		var gotMethod string

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"count":1,"results":[{"id":"grp-1","name":"group","description":"test group","district_id":"district-1","vm_ids":["vm-1","vm-2"],"status":"ready"}]}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := groups.NewClient(c, "compute", "/{version}/realm-1/{kind}/")

		result, err := client.List(context.Background(), nil)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, http.MethodGet, gotMethod)
		assert.Equal(t, "/v1/realm-1/vm-groups/", gotPath)
		assert.Equal(t, 1, result.Count)
		require.Len(t, result.Results, 1)
		assert.Equal(t, "grp-1", result.Results[0].ID)
		assert.Equal(t, "group", result.Results[0].Name)
		require.NotNil(t, result.Results[0].Description)
		assert.Equal(t, "test group", *result.Results[0].Description)
		assert.Equal(t, []string{"vm-1", "vm-2"}, result.Results[0].VMIds)
		assert.Equal(t, "district-1", result.Results[0].DistrictID)
		assert.Equal(t, model.StatusReady, result.Results[0].Status)
	})

	t.Run("Get", func(t *testing.T) {
		var gotPath string
		var gotMethod string

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id":"grp-1","name":"group","description":"test group","district_id":"district-1","vm_ids":["vm-1"],"status":"ready"}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := groups.NewClient(c, "compute", "/{version}/realm-1/{kind}/")

		result, err := client.Get(context.Background(), "grp-1", nil)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, http.MethodGet, gotMethod)
		assert.Equal(t, "/v1/realm-1/vm-groups/grp-1/", gotPath)
		assert.Equal(t, "grp-1", result.ID)
		assert.Equal(t, "group", result.Name)
		assert.Equal(t, []string{"vm-1"}, result.VMIds)
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
			_, _ = w.Write([]byte(`{"id":"grp-1","name":"group","description":"test group","district_id":"district-1","vm_ids":["vm-1","vm-2"]}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := groups.NewClient(c, "compute", "/{version}/realm-1/{kind}/")

		desc := "test group"
		req := &model.VMGroupRequest{Name: "group", Description: &desc, DistrictID: "district-1"}
		result, err := client.Create(context.Background(), req)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, http.MethodPost, gotMethod)
		assert.Equal(t, "/v1/realm-1/vm-groups/", gotPath)
		assert.Equal(t, "group", gotBody["name"])
		assert.Equal(t, "test group", gotBody["description"])
		assert.Equal(t, "district-1", gotBody["district_id"])
		assert.Equal(t, "grp-1", result.ID)
		assert.Equal(t, []string{"vm-1", "vm-2"}, result.VMIds)
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
			_, _ = w.Write([]byte(`{"id":"grp-1","name":"updated","description":"updated desc","vm_ids":["vm-1"]}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := groups.NewClient(c, "compute", "/{version}/realm-1/{kind}/")

		desc := "updated desc"
		input := &model.PatchedVMGroupRequest{Name: "updated", Description: &desc}
		result, err := client.Update(context.Background(), "grp-1", input)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, http.MethodPut, gotMethod)
		assert.Equal(t, "/v1/realm-1/vm-groups/grp-1/", gotPath)
		assert.Equal(t, "grp-1", result.ID)
		assert.Equal(t, "updated", result.Name)
		require.NotNil(t, result.Description)
		assert.Equal(t, "updated desc", *result.Description)
	})

	t.Run("UpdatePartial", func(t *testing.T) {
		var gotPath string
		var gotMethod string

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id":"grp-1","name":"group","description":"patched desc","vm_ids":["vm-1"]}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := groups.NewClient(c, "compute", "/{version}/realm-1/{kind}/")

		desc := "patched desc"
		input := &model.PatchedVMGroupRequest{Description: &desc}
		result, err := client.UpdatePartial(context.Background(), "grp-1", input)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, http.MethodPatch, gotMethod)
		assert.Equal(t, "/v1/realm-1/vm-groups/grp-1/", gotPath)
		assert.Equal(t, "grp-1", result.ID)
		require.NotNil(t, result.Description)
		assert.Equal(t, "patched desc", *result.Description)
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
		client := groups.NewClient(c, "compute", "/{version}/realm-1/{kind}/")

		err := client.Delete(context.Background(), "grp-1")
		require.NoError(t, err)

		assert.Equal(t, http.MethodDelete, gotMethod)
		assert.Equal(t, "/v1/realm-1/vm-groups/grp-1/", gotPath)
	})

	t.Run("Actions", func(t *testing.T) {
		var gotPath string
		var gotMethod string

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"name":"vmgroup","description":"","renders":[],"parses":[],"actions":{"available":["delete"],"unavailable":{}}}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := groups.NewClient(c, "compute", "/{version}/realm-1/{kind}/")

		result, err := client.Actions(context.Background(), "grp-1")
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, http.MethodOptions, gotMethod)
		assert.Equal(t, "/v1/realm-1/vm-groups/grp-1/", gotPath)
		assert.Equal(t, "vmgroup", result.Name)
	})

	t.Run("AddMember", func(t *testing.T) {
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
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id":"grp-1","name":"group","vm_ids":["vm-1","vm-2"]}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := groups.NewClient(c, "compute", "/{version}/realm-1/{kind}/")

		id := "grp-1"
		input := &model.VMGroupMemberRequest{VM: "vm-2"}
		result, err := client.AddMember(context.Background(), &id, input)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, http.MethodPost, gotMethod)
		assert.Equal(t, "/v1/realm-1/vm-groups/grp-1/:add-member", gotPath)
		assert.Equal(t, "vm-2", gotBody["vm_id"])
		assert.Equal(t, []string{"vm-1", "vm-2"}, result.VMIds)
	})

	t.Run("RemoveMember", func(t *testing.T) {
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
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id":"grp-1","name":"group","vm_ids":["vm-1"]}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := groups.NewClient(c, "compute", "/{version}/realm-1/{kind}/")

		id := "grp-1"
		input := &model.VMGroupMemberRequest{VM: "vm-2"}
		result, err := client.RemoveMember(context.Background(), &id, input)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, http.MethodPost, gotMethod)
		assert.Equal(t, "/v1/realm-1/vm-groups/grp-1/:remove-member", gotPath)
		assert.Equal(t, "vm-2", gotBody["vm_id"])
		assert.Equal(t, []string{"vm-1"}, result.VMIds)
	})

	t.Run("ErrorCase_NotFound", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"detail":"not found"}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := groups.NewClient(c, "compute", "/{version}/realm-1/{kind}/")

		result, err := client.Get(context.Background(), "nonexistent", nil)
		require.Error(t, err)
		assert.Nil(t, result)
	})
}
