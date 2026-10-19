package model

// InvoiceItem represents an invoice item as returned by the API.
type InvoiceItem struct {
	ID                  string  `json:"id,omitempty" yaml:"id,omitempty"`
	InvoiceID           string  `json:"invoice_id,omitempty" yaml:"invoice_id,omitempty"`
	ResourceID          string  `json:"resource_id,omitempty" yaml:"resource_id,omitempty"`
	ResourceName        string  `json:"resource_name,omitempty" yaml:"resource_name,omitempty"`
	ResourceKind        string  `json:"resource_kind,omitempty" yaml:"resource_kind,omitempty"`
	Service             string  `json:"service,omitempty" yaml:"service,omitempty"`
	RealmID             string  `json:"realm_id,omitempty" yaml:"realm_id,omitempty"`
	ProductID           string  `json:"product_id,omitempty" yaml:"product_id,omitempty"`
	MicroCreditsPerUnit uint64  `json:"micro_credits_per_unit" yaml:"micro_credits_per_unit"`
	BillingUnit         string  `json:"billing_unit,omitempty" yaml:"billing_unit,omitempty"`
	ProductUnit         string  `json:"product_unit,omitempty" yaml:"product_unit,omitempty"`
	ProductQuantity     uint64  `json:"product_quantity" yaml:"product_quantity"`
	TotalGrossAmount    float64 `json:"total_gross_amount" yaml:"total_gross_amount"`
	TotalNetAmount      float64 `json:"total_net_amount" yaml:"total_net_amount"`
	TotalVatAmount      float64 `json:"total_vat_amount" yaml:"total_vat_amount"`
}
