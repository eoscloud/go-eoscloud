package invoiceitem

import (
	"context"
	"net/url"

	"go.eoscloud.io/go-eoscloud/eoscloud/billing/model"
	"go.eoscloud.io/go-eoscloud/pkg/api"
)

// APIClient defines the interface for invoice-item operations.
//
// Supported filters (url.Values) on List include: invoice_id, realm_id,
// resource_id, product_id.
type APIClient interface {
	List(ctx context.Context, filter url.Values) (result *api.Page[model.InvoiceItem], err error)
}

var _ APIClient = (*client)(nil)

type client struct {
	api.Lister[model.InvoiceItem]
}

// NewClient returns the invoice-item client of the billing service.
func NewClient(c api.Caller, service, basePathPattern string) APIClient {
	pathParams := map[string]string{
		"version": "v1",
		"kind":    "invoice-items",
	}

	return &client{
		api.NewLister[model.InvoiceItem](c, service, basePathPattern, pathParams),
	}
}
