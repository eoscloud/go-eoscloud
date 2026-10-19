package model

import "time"

// Realm represents a realm in the organization service.
type Realm struct {
	Self           string     `json:"_self,omitempty" yaml:"_self,omitempty"`
	Actions        []string   `json:"_actions,omitempty" yaml:"_actions,omitempty"`
	Grid           string     `json:"grid,omitempty" yaml:"grid,omitempty"`
	ID             string     `json:"id,omitempty" yaml:"id,omitempty"`
	Name           string     `json:"name" yaml:"name"`
	OrganizationID string     `json:"organization_id,omitempty" yaml:"organization_id,omitempty"`
	Status         string     `json:"status,omitempty" yaml:"status,omitempty"`
	DeleteAt       *time.Time `json:"delete_at,omitempty" yaml:"delete_at,omitempty"`
}

// RealmRequest is the body for creating a realm.
type RealmRequest struct {
	Name           string `json:"name" yaml:"name"`
	OrganizationID string `json:"organization_id" yaml:"organization_id"`
}

// PatchedRealmRequest is the body for partially updating a realm (PATCH).
type PatchedRealmRequest struct {
	Name *string `json:"name,omitempty" yaml:"name,omitempty"`
}
