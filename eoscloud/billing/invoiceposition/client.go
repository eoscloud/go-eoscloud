package invoiceposition

import (
	"context"
	"net/url"

	"go.eoscloud.io/go-eoscloud/eoscloud/billing/model"
	"go.eoscloud.io/go-eoscloud/pkg/api"
)

// APIClient defines the interface for invoice-position operations.
//
// Supported filters (url.Values) on List include: invoice_id, product_id,
// period__gte, period__lte.
type APIClient interface {
	List(ctx context.Context, filter url.Values) (result *api.Page[model.InvoicePosition], err error)
}

var _ APIClient = (*client)(nil)

type client struct {
	api.Lister[model.InvoicePosition]
}

// NewClient returns the invoice-position client of the billing service.
func NewClient(c api.Caller, service, basePathPattern string) APIClient {
	pathParams := map[string]string{
		"version": "v1",
		"kind":    "invoice-positions",
	}

	return &client{
		api.NewLister[model.InvoicePosition](c, service, basePathPattern, pathParams),
	}
}
