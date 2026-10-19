package model

// Currency is an ISO 4217 currency code available for billing.
type Currency struct {
	Currency string `json:"currency,omitempty" yaml:"currency,omitempty"`
}
