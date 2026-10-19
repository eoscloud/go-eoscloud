package model

// CreditInfo is the price information for a given amount of credits, as
// returned by GET /billingaccount/{id}/:credit.
type CreditInfo struct {
	Net      float64 `json:"net" yaml:"net"`
	Gross    float64 `json:"gross" yaml:"gross"`
	VatRate  uint8   `json:"vat_rate" yaml:"vat_rate"`
	Vat      float64 `json:"vat" yaml:"vat"`
	Currency string  `json:"currency,omitempty" yaml:"currency,omitempty"`
}
