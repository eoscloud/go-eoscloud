package model

import "time"

// Public IP models (tag: public-ips)

// PublicIP represents a public IP as defined in compute-api.yaml
type PublicIP struct {
	Self        string       `json:"_self,omitempty" yaml:"_self,omitempty"`
	ID          string       `json:"id,omitempty" yaml:"id,omitempty"`
	IP          *string      `json:"ip,omitempty" yaml:"ip,omitempty"`
	IPFamily    IpFamilyEnum `json:"ip_family" yaml:"ip_family"`
	Description string       `json:"description,omitempty" yaml:"description,omitempty"`
	Default     bool         `json:"default,omitempty" yaml:"default,omitempty"`
	Network     string       `json:"network_id" yaml:"network_id"`
	Product     string       `json:"product,omitempty" yaml:"product,omitempty"`
	VM          *string      `json:"vm_id,omitempty" yaml:"vm_id,omitempty"`
	Status      StatusEnum   `json:"status,omitempty" yaml:"status,omitempty"`
	CreatedAt   time.Time    `json:"created_at,omitempty" yaml:"created_at,omitempty"`
	UpdatedAt   time.Time    `json:"updated_at,omitempty" yaml:"updated_at,omitempty"`
	Actions     []string     `json:"_actions,omitempty" yaml:"_actions,omitempty"`
}

// PublicIPRequest represents a request to create a public IP
type PublicIPRequest struct {
	IPFamily    IpFamilyEnum `json:"ip_family" yaml:"ip_family" shortflag:"F"`
	Network     string       `json:"network_id" yaml:"network_id" shortflag:"N"`
	Description string       `json:"description,omitempty" yaml:"description,omitempty" shortflag:"D"`
}

// PublicIPAttachRequest represents a request to attach a public IP to a VM
type PublicIPAttachRequest struct {
	VM string `json:"vm_id" yaml:"vm_id" shortflag:"V"`
}

// PatchedPublicIPRequest represents a request to update a public IP
type PatchedPublicIPRequest struct {
	Description string `json:"description,omitempty" yaml:"description,omitempty" shortflag:"D"`
}
