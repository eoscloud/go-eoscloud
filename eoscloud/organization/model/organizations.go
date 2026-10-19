package model

import "time"

// Organization represents an organization in the organization service.
type Organization struct {
	Self      string     `json:"_self,omitempty" yaml:"_self,omitempty"`
	Actions   []string   `json:"_actions,omitempty" yaml:"_actions,omitempty"`
	Grid      string     `json:"grid,omitempty" yaml:"grid,omitempty"`
	ID        string     `json:"id,omitempty" yaml:"id,omitempty"`
	Name      string     `json:"name" yaml:"name"`
	Status    string     `json:"status,omitempty" yaml:"status,omitempty"`
	CreatedAt *time.Time `json:"created_at,omitempty" yaml:"created_at,omitempty"`
	UpdatedAt *time.Time `json:"updated_at,omitempty" yaml:"updated_at,omitempty"`
	DeleteAt  *time.Time `json:"delete_at,omitempty" yaml:"delete_at,omitempty"`
}

// OrganizationRequest is the body for creating an organization.
type OrganizationRequest struct {
	Name string `json:"name" yaml:"name"`
}

// PatchedOrganizationRequest is the body for partially updating an organization (PATCH).
type PatchedOrganizationRequest struct {
	Name *string `json:"name,omitempty" yaml:"name,omitempty"`
}
