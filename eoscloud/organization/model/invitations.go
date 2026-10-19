package model

import "time"

// Invitation represents an invitation of a user to an organization.
type Invitation struct {
	Self           string     `json:"_self,omitempty" yaml:"_self,omitempty"`
	Actions        []string   `json:"_actions,omitempty" yaml:"_actions,omitempty"`
	Grid           string     `json:"grid,omitempty" yaml:"grid,omitempty"`
	ID             string     `json:"id,omitempty" yaml:"id,omitempty"`
	OrganizationID *string    `json:"organization_id,omitempty" yaml:"organization_id,omitempty"`
	Email          *string    `json:"email,omitempty" yaml:"email,omitempty"`
	Roles          []string   `json:"roles,omitempty" yaml:"roles,omitempty"`
	Status         *string    `json:"status,omitempty" yaml:"status,omitempty"`
	SentAt         *time.Time `json:"sent_at,omitempty" yaml:"sent_at,omitempty"`
	CreatedAt      *time.Time `json:"created_at,omitempty" yaml:"created_at,omitempty"`
	CreatedBy      *string    `json:"created_by,omitempty" yaml:"created_by,omitempty"`
	UpdatedAt      *time.Time `json:"updated_at,omitempty" yaml:"updated_at,omitempty"`
}

// InvitationRequest is the body for creating an invitation.
type InvitationRequest struct {
	Email          string   `json:"email" yaml:"email"`
	OrganizationID string   `json:"organization_id" yaml:"organization_id"`
	Roles          []string `json:"roles,omitempty" yaml:"roles,omitempty"`
}
