package invoice_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.eoscloud.io/go-eoscloud/eoscloud/billing/invoice"
	"go.eoscloud.io/go-eoscloud/pkg/api"
)

func TestInvoiceClient(t *testing.T) {
	t.Run("List", func(t *testing.T) {
		var gotPath string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"count":1,"results":[{"id":"inv-1","currency":"EUR"}]}`))
		}))
		defer srv.Close()

		c := invoice.NewClient(api.NewClient(api.WithBaseURL(srv.URL)), "billing", "/{version}/{kind}/")
		result, err := c.List(context.Background(), nil)
		require.NoError(t, err)

		assert.Equal(t, "/v1/invoices/", gotPath)
		require.Len(t, result.Results, 1)
		assert.Equal(t, "inv-1", result.Results[0].ID)
	})

	t.Run("Get", func(t *testing.T) {
		var gotPath string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id":"inv-1","currency":"EUR"}`))
		}))
		defer srv.Close()

		c := invoice.NewClient(api.NewClient(api.WithBaseURL(srv.URL)), "billing", "/{version}/{kind}/")
		result, err := c.Get(context.Background(), "inv-1", nil)
		require.NoError(t, err)

		assert.Equal(t, "/v1/invoices/inv-1/", gotPath)
		assert.Equal(t, "inv-1", result.ID)
	})
}
