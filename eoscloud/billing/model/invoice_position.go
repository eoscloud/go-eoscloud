package model

// InvoicePosition represents an aggregated invoice position (per product) as
// returned by the API.
type InvoicePosition struct {
	InvoiceID        string  `json:"invoice_id,omitempty" yaml:"invoice_id,omitempty"`
	ProductID        string  `json:"product_id,omitempty" yaml:"product_id,omitempty"`
	ProductName      string  `json:"product_name,omitempty" yaml:"product_name,omitempty"`
	BillingQuantity  uint64  `json:"billing_quantity" yaml:"billing_quantity"`
	BillingUnit      string  `json:"billing_unit,omitempty" yaml:"billing_unit,omitempty"`
	ProductUnit      string  `json:"product_unit,omitempty" yaml:"product_unit,omitempty"`
	ProductQuantity  uint64  `json:"product_quantity" yaml:"product_quantity"`
	TotalGrossAmount float64 `json:"total_gross_amount" yaml:"total_gross_amount"`
	TotalNetAmount   float64 `json:"total_net_amount" yaml:"total_net_amount"`
	TotalVatAmount   float64 `json:"total_vat_amount" yaml:"total_vat_amount"`
	Vat              uint8   `json:"vat" yaml:"vat"`
}
