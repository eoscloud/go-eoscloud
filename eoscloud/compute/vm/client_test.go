package vm_test

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
	"go.eoscloud.io/go-eoscloud/eoscloud/compute/vm"
	"go.eoscloud.io/go-eoscloud/pkg/api"
)

var vmDetailJSON = map[string]any{
	"id":         "vm-1",
	"name":       "test-vm",
	"block_id":   "b-1",
	"product_id": "p-1",
	"network_id": "n-1",
	"image_id":   "i-1",
	"disk_size":  20,
}

func newVMClient(srvURL string) vm.APIClient {
	apiClient := api.NewClient(api.WithBaseURL(srvURL))
	return vm.NewClient(apiClient, "compute", "/{version}/realm-1/{kind}/")
}

func vmErrorHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"type":       "error",
			"request_id": "req-1",
			"errors":     []map[string]any{{"path": "/", "kind": "not_found", "message": "not found"}},
		})
	}
}

func TestVMClient_List(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var gotPath, gotMethod string

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"count":   1,
				"results": []map[string]any{vmDetailJSON},
			})
		}))
		defer srv.Close()

		c := newVMClient(srv.URL)
		result, err := c.List(context.Background(), nil)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, "/v1/realm-1/vms/", gotPath)
		assert.Equal(t, http.MethodGet, gotMethod)
		assert.Equal(t, 1, result.Count)
		require.Len(t, result.Results, 1)
		assert.Equal(t, "vm-1", result.Results[0].ID)
		assert.Equal(t, "test-vm", result.Results[0].Name)
	})
}

func TestVMClient_Get(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var gotPath, gotMethod string

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(vmDetailJSON)
		}))
		defer srv.Close()

		c := newVMClient(srv.URL)
		result, err := c.Get(context.Background(), "vm-1", nil)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, "/v1/realm-1/vms/vm-1/", gotPath)
		assert.Equal(t, http.MethodGet, gotMethod)
		assert.Equal(t, "vm-1", result.ID)
		assert.Equal(t, "test-vm", result.Name)
	})
}

func TestVMClient_Create(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var gotPath, gotMethod string
		var gotBody map[string]any

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method
			body, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(body, &gotBody)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(vmDetailJSON)
		}))
		defer srv.Close()

		c := newVMClient(srv.URL)
		input := &model.VMRequest{
			Block:    "b-1",
			Product:  "p-1",
			Network:  "n-1",
			Image:    "i-1",
			Name:     "test-vm",
			DiskSize: 20,
		}
		result, err := c.Create(context.Background(), input)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, "/v1/realm-1/vms/", gotPath)
		assert.Equal(t, http.MethodPost, gotMethod)
		assert.Equal(t, "b-1", gotBody["block_id"])
		assert.Equal(t, "vm-1", result.ID)
	})
}

func TestVMClient_Update(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var gotPath, gotMethod string

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(vmDetailJSON)
		}))
		defer srv.Close()

		c := newVMClient(srv.URL)
		input := &model.PatchedVMRequest{Description: "updated"}
		result, err := c.Update(context.Background(), "vm-1", input)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, "/v1/realm-1/vms/vm-1/", gotPath)
		assert.Equal(t, http.MethodPut, gotMethod)
	})
}

func TestVMClient_UpdatePartial(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var gotPath, gotMethod string

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(vmDetailJSON)
		}))
		defer srv.Close()

		c := newVMClient(srv.URL)
		input := &model.PatchedVMRequest{Description: "patched"}
		result, err := c.UpdatePartial(context.Background(), "vm-1", input)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, "/v1/realm-1/vms/vm-1/", gotPath)
		assert.Equal(t, http.MethodPatch, gotMethod)
	})
}

func TestVMClient_Delete(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var gotPath, gotMethod string

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method
			w.WriteHeader(http.StatusNoContent)
		}))
		defer srv.Close()

		c := newVMClient(srv.URL)
		err := c.Delete(context.Background(), "vm-1")
		require.NoError(t, err)

		assert.Equal(t, "/v1/realm-1/vms/vm-1/", gotPath)
		assert.Equal(t, http.MethodDelete, gotMethod)
	})
}

func TestVMClient_Actions(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var gotPath, gotMethod string

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"name":        "vm",
				"description": "VM actions",
				"renders":     []string{"application/json"},
				"parses":      []string{"application/json"},
				"actions": map[string]any{
					"available":   []string{"power-on", "power-off"},
					"unavailable": map[string]any{},
				},
			})
		}))
		defer srv.Close()

		c := newVMClient(srv.URL)
		result, err := c.Actions(context.Background(), "vm-1")
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, "/v1/realm-1/vms/vm-1/", gotPath)
		assert.Equal(t, http.MethodOptions, gotMethod)
		assert.Equal(t, "vm", result.Name)
	})
}

func TestVMClient_PowerOn(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var gotPath, gotMethod string
		var gotBody map[string]any

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method
			body, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(body, &gotBody)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusAccepted)
			_ = json.NewEncoder(w).Encode(vmDetailJSON)
		}))
		defer srv.Close()

		c := newVMClient(srv.URL)
		id := "123"
		input := &model.VMPowerActionRequest{Timeout: 30}
		result, err := c.PowerOn(context.Background(), &id, input)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, "/v1/realm-1/vms/123/:power-on", gotPath)
		assert.Equal(t, http.MethodPost, gotMethod)
		assert.Equal(t, float64(30), gotBody["timeout"])
		assert.Equal(t, "vm-1", result.ID)
	})
}

func TestVMClient_PowerOff(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var gotPath, gotMethod string

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusAccepted)
			_ = json.NewEncoder(w).Encode(vmDetailJSON)
		}))
		defer srv.Close()

		c := newVMClient(srv.URL)
		id := "123"
		input := &model.VMPowerActionRequest{Timeout: 30}
		result, err := c.PowerOff(context.Background(), &id, input)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, "/v1/realm-1/vms/123/:power-off", gotPath)
		assert.Equal(t, http.MethodPost, gotMethod)
	})
}

func TestVMClient_ForcePowerOff(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var gotPath, gotMethod string

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusAccepted)
			_ = json.NewEncoder(w).Encode(vmDetailJSON)
		}))
		defer srv.Close()

		c := newVMClient(srv.URL)
		id := "123"
		input := &model.VMPowerActionRequest{Timeout: 30}
		result, err := c.ForcePowerOff(context.Background(), &id, input)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, "/v1/realm-1/vms/123/:force-power-off", gotPath)
		assert.Equal(t, http.MethodPost, gotMethod)
	})
}

func TestVMClient_Reboot(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var gotPath, gotMethod string

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusAccepted)
			_ = json.NewEncoder(w).Encode(vmDetailJSON)
		}))
		defer srv.Close()

		c := newVMClient(srv.URL)
		id := "123"
		input := &model.VMPowerActionRequest{Timeout: 30}
		result, err := c.Reboot(context.Background(), &id, input)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, "/v1/realm-1/vms/123/:reboot", gotPath)
		assert.Equal(t, http.MethodPost, gotMethod)
	})
}

func TestVMClient_Upscale(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var gotPath, gotMethod string
		var gotBody map[string]any

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method
			body, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(body, &gotBody)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusAccepted)
			_ = json.NewEncoder(w).Encode(vmDetailJSON)
		}))
		defer srv.Close()

		c := newVMClient(srv.URL)
		id := "123"
		input := &model.VMUpscaleRequest{ProductID: "p-2"}
		result, err := c.Upscale(context.Background(), &id, input)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, "/v1/realm-1/vms/123/:upscale", gotPath)
		assert.Equal(t, http.MethodPost, gotMethod)
		assert.Equal(t, "p-2", gotBody["product_id"])
		assert.Equal(t, "vm-1", result.ID)
	})
}

func TestVMClient_ChangeNetwork(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var gotPath, gotMethod string
		var gotBody map[string]any

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method
			body, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(body, &gotBody)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusAccepted)
			_ = json.NewEncoder(w).Encode(vmDetailJSON)
		}))
		defer srv.Close()

		c := newVMClient(srv.URL)
		id := "123"
		input := &model.VMChangeNetworkRequest{NetworkID: "n-2"}
		result, err := c.ChangeNetwork(context.Background(), &id, input)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, "/v1/realm-1/vms/123/:change-network", gotPath)
		assert.Equal(t, http.MethodPost, gotMethod)
		assert.Equal(t, "n-2", gotBody["network_id"])
		assert.Equal(t, "vm-1", result.ID)
	})
}

func TestVMClient_Stats(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var gotPath, gotMethod string

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"vm_public_ips_total_count": 2,
				"vm_public_ips_v4_count":    1,
				"vm_public_ips_v6_count":    1,
				"vm_volumes_total_count":    3,
			})
		}))
		defer srv.Close()

		c := newVMClient(srv.URL)
		id := "123"
		result, err := c.Stats(context.Background(), &id)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, "/v1/realm-1/vms/123/:stats", gotPath)
		assert.Equal(t, http.MethodGet, gotMethod)
		assert.Equal(t, float64(2), (*result)["vm_public_ips_total_count"])
		assert.Equal(t, float64(3), (*result)["vm_volumes_total_count"])
	})
}

func TestVMClient_Error(t *testing.T) {
	t.Run("404_ReturnsError", func(t *testing.T) {
		srv := httptest.NewServer(vmErrorHandler())
		defer srv.Close()

		c := newVMClient(srv.URL)
		result, err := c.List(context.Background(), nil)
		require.Error(t, err)
		assert.Nil(t, result)
	})
}
