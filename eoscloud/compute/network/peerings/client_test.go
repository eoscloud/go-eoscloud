package peerings_test

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
	netPeerings "go.eoscloud.io/go-eoscloud/eoscloud/compute/network/peerings"
	"go.eoscloud.io/go-eoscloud/pkg/api"
)

func TestNewClient(t *testing.T) {
	c := api.NewClient()
	var _ = netPeerings.NewClient(c, "compute", "/{version}/realm-1/{kind}/")
}

func TestNetworkPeeringClient(t *testing.T) {
	t.Run("List", func(t *testing.T) {
		var gotPath string
		var gotMethod string

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"count":1,"results":[{"id":"peer-1","name":"peering","description":"test peering","district_id":"district-1","network_ids":["net-1","net-2"],"status":"ready"}]}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := netPeerings.NewClient(c, "compute", "/{version}/realm-1/{kind}/")

		result, err := client.List(context.Background(), nil)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, http.MethodGet, gotMethod)
		assert.Equal(t, "/v1/realm-1/network-peerings/", gotPath)
		assert.Equal(t, 1, result.Count)
		require.Len(t, result.Results, 1)
		assert.Equal(t, "peer-1", result.Results[0].ID)
		assert.Equal(t, "peering", result.Results[0].Name)
		require.NotNil(t, result.Results[0].Description)
		assert.Equal(t, "test peering", *result.Results[0].Description)
		assert.Equal(t, []string{"net-1", "net-2"}, result.Results[0].NetworkIds)
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
			_, _ = w.Write([]byte(`{"id":"peer-1","name":"peering","description":"test peering","district_id":"district-1","network_ids":["net-1"],"status":"ready"}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := netPeerings.NewClient(c, "compute", "/{version}/realm-1/{kind}/")

		result, err := client.Get(context.Background(), "peer-1", nil)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, http.MethodGet, gotMethod)
		assert.Equal(t, "/v1/realm-1/network-peerings/peer-1/", gotPath)
		assert.Equal(t, "peer-1", result.ID)
		assert.Equal(t, "peering", result.Name)
		assert.Equal(t, []string{"net-1"}, result.NetworkIds)
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
			_, _ = w.Write([]byte(`{"id":"peer-1","name":"peering","description":"test peering","district_id":"district-1","network_ids":["net-1","net-2"]}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := netPeerings.NewClient(c, "compute", "/{version}/realm-1/{kind}/")

		desc := "test peering"
		req := &model.NetworkPeeringRequest{Name: "peering", Description: &desc, DistrictID: "district-1"}
		result, err := client.Create(context.Background(), req)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, http.MethodPost, gotMethod)
		assert.Equal(t, "/v1/realm-1/network-peerings/", gotPath)
		assert.Equal(t, "peering", gotBody["name"])
		assert.Equal(t, "test peering", gotBody["description"])
		assert.Equal(t, "district-1", gotBody["district_id"])
		assert.Equal(t, "peer-1", result.ID)
		assert.Equal(t, []string{"net-1", "net-2"}, result.NetworkIds)
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
			_, _ = w.Write([]byte(`{"id":"peer-1","name":"updated","description":"updated desc","network_ids":["net-1"]}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := netPeerings.NewClient(c, "compute", "/{version}/realm-1/{kind}/")

		desc := "updated desc"
		input := &model.PatchedNetworkPeeringRequest{Name: "updated", Description: &desc}
		result, err := client.Update(context.Background(), "peer-1", input)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, http.MethodPut, gotMethod)
		assert.Equal(t, "/v1/realm-1/network-peerings/peer-1/", gotPath)
		assert.Equal(t, "peer-1", result.ID)
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
			_, _ = w.Write([]byte(`{"id":"peer-1","name":"peering","description":"patched desc","network_ids":["net-1"]}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := netPeerings.NewClient(c, "compute", "/{version}/realm-1/{kind}/")

		desc := "patched desc"
		input := &model.PatchedNetworkPeeringRequest{Description: &desc}
		result, err := client.UpdatePartial(context.Background(), "peer-1", input)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, http.MethodPatch, gotMethod)
		assert.Equal(t, "/v1/realm-1/network-peerings/peer-1/", gotPath)
		assert.Equal(t, "peer-1", result.ID)
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
		client := netPeerings.NewClient(c, "compute", "/{version}/realm-1/{kind}/")

		err := client.Delete(context.Background(), "peer-1")
		require.NoError(t, err)

		assert.Equal(t, http.MethodDelete, gotMethod)
		assert.Equal(t, "/v1/realm-1/network-peerings/peer-1/", gotPath)
	})

	t.Run("Actions", func(t *testing.T) {
		var gotPath string
		var gotMethod string

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"name":"networkpeering","description":"","renders":[],"parses":[],"actions":{"available":["delete"],"unavailable":{}}}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := netPeerings.NewClient(c, "compute", "/{version}/realm-1/{kind}/")

		result, err := client.Actions(context.Background(), "peer-1")
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, http.MethodOptions, gotMethod)
		assert.Equal(t, "/v1/realm-1/network-peerings/peer-1/", gotPath)
		assert.Equal(t, "networkpeering", result.Name)
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
			_, _ = w.Write([]byte(`{"id":"peer-1","name":"peering","network_ids":["net-1","net-2"]}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := netPeerings.NewClient(c, "compute", "/{version}/realm-1/{kind}/")

		id := "peer-1"
		input := &model.NetworkPeeringMemberRequest{Network: "net-2"}
		result, err := client.AddMember(context.Background(), &id, input)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, http.MethodPost, gotMethod)
		assert.Equal(t, "/v1/realm-1/network-peerings/peer-1/:add-member", gotPath)
		assert.Equal(t, "net-2", gotBody["network_id"])
		assert.Equal(t, []string{"net-1", "net-2"}, result.NetworkIds)
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
			_, _ = w.Write([]byte(`{"id":"peer-1","name":"peering","network_ids":["net-1"]}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := netPeerings.NewClient(c, "compute", "/{version}/realm-1/{kind}/")

		id := "peer-1"
		input := &model.NetworkPeeringMemberRequest{Network: "net-2"}
		result, err := client.RemoveMember(context.Background(), &id, input)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, http.MethodPost, gotMethod)
		assert.Equal(t, "/v1/realm-1/network-peerings/peer-1/:remove-member", gotPath)
		assert.Equal(t, "net-2", gotBody["network_id"])
		assert.Equal(t, []string{"net-1"}, result.NetworkIds)
	})

	t.Run("ErrorCase_NotFound", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"detail":"not found"}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := netPeerings.NewClient(c, "compute", "/{version}/realm-1/{kind}/")

		result, err := client.Get(context.Background(), "nonexistent", nil)
		require.Error(t, err)
		assert.Nil(t, result)
	})
}
