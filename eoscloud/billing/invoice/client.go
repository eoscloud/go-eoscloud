package invoice

import (
	"context"
	"net/url"

	"go.eoscloud.io/go-eoscloud/eoscloud/billing/model"
	"go.eoscloud.io/go-eoscloud/pkg/api"
)

// APIClient defines the interface for invoice operations.
//
// Supported filters (url.Values) on List include: the standard pagination,
// ordering and search parameters.
type APIClient interface {
	List(ctx context.Context, filter url.Values) (result *api.Page[model.Invoice], err error)
	Get(ctx context.Context, id string, filter url.Values) (result *model.Invoice, err error)
}

var _ APIClient = (*client)(nil)

type client struct {
	api.Lister[model.Invoice]
	api.Getter[model.Invoice]
}

// NewClient returns the invoice client of the billing service.
func NewClient(c api.Caller, service, basePathPattern string) APIClient {
	pathParams := map[string]string{
		"version": "v1",
		"kind":    "invoices",
	}

	return &client{
		api.NewLister[model.Invoice](c, service, basePathPattern, pathParams),
		api.NewGetter[model.Invoice](c, service, basePathPattern, pathParams),
	}
}
