package model

import "time"

// Invoice represents an invoice as returned by the API.
type Invoice struct {
	Self                 string     `json:"_self,omitempty" yaml:"_self,omitempty"`
	Actions              []string   `json:"_actions,omitempty" yaml:"_actions,omitempty"`
	Grid                 string     `json:"grid,omitempty" yaml:"grid,omitempty"`
	ID                   string     `json:"id,omitempty" yaml:"id,omitempty"`
	BillingAccountID     string     `json:"billing_account_id,omitempty" yaml:"billing_account_id,omitempty"`
	Currency             string     `json:"currency,omitempty" yaml:"currency,omitempty"`
	TotalGrossAmount     float64    `json:"total_gross_amount" yaml:"total_gross_amount"`
	TotalNetAmount       float64    `json:"total_net_amount" yaml:"total_net_amount"`
	TotalVatAmount       float64    `json:"total_vat_amount" yaml:"total_vat_amount"`
	Status               string     `json:"status,omitempty" yaml:"status,omitempty"`
	PeriodStart          time.Time  `json:"period_start" yaml:"period_start"`
	PeriodEnd            time.Time  `json:"period_end" yaml:"period_end"`
	CalculationTimestamp time.Time  `json:"calculation_timestamp" yaml:"calculation_timestamp"`
	PaymentTimestamp     *time.Time `json:"payment_timestamp,omitempty" yaml:"payment_timestamp,omitempty"`
	PaymentStatus        string     `json:"payment_status,omitempty" yaml:"payment_status,omitempty"`
	PaymentFailureReason string     `json:"payment_failure_reason,omitempty" yaml:"payment_failure_reason,omitempty"`
}
