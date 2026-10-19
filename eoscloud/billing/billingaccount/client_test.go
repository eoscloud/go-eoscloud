package billingaccount_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.eoscloud.io/go-eoscloud/eoscloud/billing/billingaccount"
	"go.eoscloud.io/go-eoscloud/eoscloud/billing/model"
	"go.eoscloud.io/go-eoscloud/pkg/api"
)

func TestNewClient(t *testing.T) {
	c := api.NewClient()
	var _ = billingaccount.NewClient(c, "billing", "/{version}/{kind}/")
}

func TestBillingAccountClient(t *testing.T) {
	t.Run("Get", func(t *testing.T) {
		var gotPath, gotMethod string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id":"ba-1","organization_id":"org-1","name":"Acme"}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := billingaccount.NewClient(c, "billing", "/{version}/{kind}/")

		result, err := client.Get(context.Background(), "ba-1", nil)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, http.MethodGet, gotMethod)
		assert.Equal(t, "/v1/billingaccount/ba-1/", gotPath)
		assert.Equal(t, "ba-1", result.ID)
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
			_, _ = w.Write([]byte(`{"id":"ba-1","organization_id":"org-1"}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := billingaccount.NewClient(c, "billing", "/{version}/{kind}/")

		req := &model.CreateBillingAccountRequest{OrganizationID: "org-1"}
		result, err := client.Create(context.Background(), req)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, http.MethodPost, gotMethod)
		assert.Equal(t, "/v1/billingaccount/", gotPath)
		assert.Equal(t, "org-1", gotBody["organization_id"])
		assert.Equal(t, "ba-1", result.ID)
	})

	t.Run("UpdatePartial", func(t *testing.T) {
		var gotPath, gotMethod string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id":"ba-1","name":"Updated"}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := billingaccount.NewClient(c, "billing", "/{version}/{kind}/")

		name := "Updated"
		req := &model.UpdateBillingAccountRequest{BillingAccount: model.BillingAccount{Name: &name}}
		result, err := client.UpdatePartial(context.Background(), "ba-1", req)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, http.MethodPatch, gotMethod)
		assert.Equal(t, "/v1/billingaccount/ba-1/", gotPath)
		assert.Equal(t, "Updated", *result.Name)
	})

	t.Run("GetCreditInfo", func(t *testing.T) {
		var gotPath, gotMethod, gotQuery string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method
			gotQuery = r.URL.RawQuery
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"net":1.5,"gross":1.8,"vat_rate":20,"vat":0.3,"currency":"EUR"}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := billingaccount.NewClient(c, "billing", "/{version}/{kind}/")

		result, err := client.GetCreditInfo(context.Background(), "ba-1", billingaccount.GetCreditInfoOptions{
			ProductID:    "prod-1",
			MicroCredits: 1000,
		})
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, http.MethodGet, gotMethod)
		assert.Equal(t, "/v1/billingaccount/ba-1/:credit", gotPath)
		assert.Contains(t, gotQuery, "product_id=prod-1")
		assert.Contains(t, gotQuery, "micro_credits=1000")
		assert.Equal(t, "EUR", result.Currency)
	})

	t.Run("ErrorCase_NotFound", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"detail":"not found"}`))
		}))
		defer srv.Close()

		c := api.NewClient(api.WithBaseURL(srv.URL))
		client := billingaccount.NewClient(c, "billing", "/{version}/{kind}/")

		result, err := client.Get(context.Background(), "nonexistent", nil)
		require.Error(t, err)
		assert.Nil(t, result)
	})
}
