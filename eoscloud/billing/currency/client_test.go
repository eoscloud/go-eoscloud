package currency_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.eoscloud.io/go-eoscloud/eoscloud/billing/currency"
	"go.eoscloud.io/go-eoscloud/pkg/api"
)

func TestCurrencyClient_List(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"count":1,"results":[{"currency":"EUR"}]}`))
	}))
	defer srv.Close()

	c := currency.NewClient(api.NewClient(api.WithBaseURL(srv.URL)), "billing", "/{version}/{kind}/")
	result, err := c.List(context.Background(), nil)
	require.NoError(t, err)

	assert.Equal(t, "/v1/currencies/", gotPath)
	require.Len(t, result.Results, 1)
	assert.Equal(t, "EUR", result.Results[0].Currency)
}
