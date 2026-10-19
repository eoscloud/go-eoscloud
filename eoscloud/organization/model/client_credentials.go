package model

import "time"

// ClientCredential represents an OAuth2 client-credentials grant for an organization or realm.
type ClientCredential struct {
	Self           string     `json:"_self,omitempty" yaml:"_self,omitempty"`
	Actions        []string   `json:"_actions,omitempty" yaml:"_actions,omitempty"`
	Grid           string     `json:"grid,omitempty" yaml:"grid,omitempty"`
	ID             string     `json:"id,omitempty" yaml:"id,omitempty"`
	AssignmentID   string     `json:"assignment_id,omitempty" yaml:"assignment_id,omitempty"`
	AssignmentType string     `json:"assignment_type,omitempty" yaml:"assignment_type,omitempty"`
	Name           *string    `json:"name,omitempty" yaml:"name,omitempty"`
	Roles          []string   `json:"roles,omitempty" yaml:"roles,omitempty"`
	Secret         *string    `json:"secret,omitempty" yaml:"secret,omitempty"`
	CreatedAt      *time.Time `json:"created_at,omitempty" yaml:"created_at,omitempty"`
	UpdatedAt      *time.Time `json:"updated_at,omitempty" yaml:"updated_at,omitempty"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty" yaml:"expires_at,omitempty"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty" yaml:"deleted_at,omitempty"`
}

// ClientCredentialRequest is the body for creating a client-credential.
// AssignmentType must be "organization" or "realm".
type ClientCredentialRequest struct {
	Name           string    `json:"name" yaml:"name"`
	AssignmentID   string    `json:"assignment_id" yaml:"assignment_id"`
	AssignmentType string    `json:"assignment_type" yaml:"assignment_type"`
	Roles          []string  `json:"roles,omitempty" yaml:"roles,omitempty"`
	ExpiresAt      time.Time `json:"expires_at" yaml:"expires_at"`
}

// PatchedClientCredentialRequest is the body for partially updating a client-credential (PATCH).
// A nil Roles field is omitted from the request, while a non-nil but empty
// slice is sent as an empty list, clearing the client-credential's roles.
type PatchedClientCredentialRequest struct {
	Name  *string   `json:"name,omitempty" yaml:"name,omitempty"`
	Roles *[]string `json:"roles,omitempty" yaml:"roles,omitempty"`
}
