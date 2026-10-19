package model

// BillingAccount holds the address, contact and billing attributes shared by
// the create/update requests and the BillingAccountItem response.
type BillingAccount struct {
	Street1      *string `json:"street1,omitempty" yaml:"street1,omitempty"`
	Street2      *string `json:"street2,omitempty" yaml:"street2,omitempty"`
	City         *string `json:"city,omitempty" yaml:"city,omitempty"`
	State        *string `json:"state,omitempty" yaml:"state,omitempty"`
	PostalCode   *string `json:"postal_code,omitempty" yaml:"postal_code,omitempty"`
	CountryCode  *string `json:"country_code,omitempty" yaml:"country_code,omitempty"`
	Name         *string `json:"name,omitempty" yaml:"name,omitempty"`
	Email        *string `json:"contact_email,omitempty" yaml:"contact_email,omitempty"`
	Phone        *string `json:"contact_phone,omitempty" yaml:"contact_phone,omitempty"`
	TaxId        *string `json:"tax_id,omitempty" yaml:"tax_id,omitempty"`
	CurrencyCode *string `json:"currency_code,omitempty" yaml:"currency_code,omitempty"`
}

// BillingAccountItem represents a billing account as returned by the API.
type BillingAccountItem struct {
	Self           string   `json:"_self,omitempty" yaml:"_self,omitempty"`
	Actions        []string `json:"_actions,omitempty" yaml:"_actions,omitempty"`
	Grid           string   `json:"grid,omitempty" yaml:"grid,omitempty"`
	ID             string   `json:"id,omitempty" yaml:"id,omitempty"`
	OrganizationID string   `json:"organization_id,omitempty" yaml:"organization_id,omitempty"`
	*BillingAccount
}

// CreateBillingAccountRequest is the body for creating a billing account.
type CreateBillingAccountRequest struct {
	OrganizationID string `json:"organization_id" yaml:"organization_id"`
	BillingAccount
}

// UpdateBillingAccountRequest is the body for updating a billing account.
type UpdateBillingAccountRequest struct {
	BillingAccount
}
