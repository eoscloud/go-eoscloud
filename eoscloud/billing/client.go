// Package billing is the go-eoscloud client for the billing service's public
// API (billing accounts, invoices, invoice items/positions, currencies). The
// billing service is GLOBAL, not realm-scoped.
package billing

import (
	"go.eoscloud.io/go-eoscloud/eoscloud/billing/billingaccount"
	"go.eoscloud.io/go-eoscloud/eoscloud/billing/currency"
	"go.eoscloud.io/go-eoscloud/eoscloud/billing/invoice"
	"go.eoscloud.io/go-eoscloud/eoscloud/billing/invoiceitem"
	"go.eoscloud.io/go-eoscloud/eoscloud/billing/invoiceposition"
	"go.eoscloud.io/go-eoscloud/pkg/api"
)

const serviceName = "billing"

// Client represents the billing service client.
type Client struct {
	BillingAccount  billingaccount.APIClient
	Invoice         invoice.APIClient
	InvoiceItem     invoiceitem.APIClient
	InvoicePosition invoiceposition.APIClient
	Currency        currency.APIClient
}

// NewClient returns a billing service client.
func NewClient(c api.Caller, baseGlobalPathPattern string) *Client {
	return &Client{
		BillingAccount:  billingaccount.NewClient(c, serviceName, baseGlobalPathPattern),
		Invoice:         invoice.NewClient(c, serviceName, baseGlobalPathPattern),
		InvoiceItem:     invoiceitem.NewClient(c, serviceName, baseGlobalPathPattern),
		InvoicePosition: invoiceposition.NewClient(c, serviceName, baseGlobalPathPattern),
		Currency:        currency.NewClient(c, serviceName, baseGlobalPathPattern),
	}
}
