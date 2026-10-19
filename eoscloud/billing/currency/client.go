package currency

import (
	"context"
	"net/url"

	"go.eoscloud.io/go-eoscloud/eoscloud/billing/model"
	"go.eoscloud.io/go-eoscloud/pkg/api"
)

// APIClient defines the interface for listing the currencies available for
// billing (i.e. that have a defined credit price).
//
// Supported filters (url.Values) on List include: currency.
type APIClient interface {
	List(ctx context.Context, filter url.Values) (result *api.Page[model.Currency], err error)
}

var _ APIClient = (*client)(nil)

type client struct {
	api.Lister[model.Currency]
}

// NewClient returns the currency client of the billing service.
func NewClient(c api.Caller, service, basePathPattern string) APIClient {
	pathParams := map[string]string{
		"version": "v1",
		"kind":    "currencies",
	}

	return &client{
		api.NewLister[model.Currency](c, service, basePathPattern, pathParams),
	}
}
