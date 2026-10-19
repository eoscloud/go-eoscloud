package model

import "time"

// Member represents a member of an organization or realm.
type Member struct {
	Self           string     `json:"_self,omitempty" yaml:"_self,omitempty"`
	Actions        []string   `json:"_actions,omitempty" yaml:"_actions,omitempty"`
	Grid           string     `json:"grid,omitempty" yaml:"grid,omitempty"`
	ID             string     `json:"id,omitempty" yaml:"id,omitempty"`
	AssignmentID   string     `json:"assignment_id,omitempty" yaml:"assignment_id,omitempty"`
	AssignmentType string     `json:"assignment_type,omitempty" yaml:"assignment_type,omitempty"`
	IdentityID     string     `json:"identity_id,omitempty" yaml:"identity_id,omitempty"`
	Email          *string    `json:"email,omitempty" yaml:"email,omitempty"`
	Firstname      *string    `json:"firstname,omitempty" yaml:"firstname,omitempty"`
	Lastname       *string    `json:"lastname,omitempty" yaml:"lastname,omitempty"`
	Roles          []string   `json:"roles,omitempty" yaml:"roles,omitempty"`
	DirectMember   bool       `json:"direct_member,omitempty" yaml:"direct_member,omitempty"`
	CreatedAt      *time.Time `json:"created_at,omitempty" yaml:"created_at,omitempty"`
	CreatedBy      *string    `json:"created_by,omitempty" yaml:"created_by,omitempty"`
	UpdatedAt      *time.Time `json:"updated_at,omitempty" yaml:"updated_at,omitempty"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty" yaml:"deleted_at,omitempty"`
}

// MemberRequest is the body for creating a realm member.
type MemberRequest struct {
	RealmID    string   `json:"realm_id" yaml:"realm_id"`
	IdentityID string   `json:"identity_id" yaml:"identity_id"`
	Roles      []string `json:"roles" yaml:"roles"`
}

// PatchedMemberRequest is the body for partially updating a member (PATCH).
// A nil Roles field is omitted from the request, while a non-nil but empty
// slice is sent as an empty list, clearing the member's roles.
type PatchedMemberRequest struct {
	Roles *[]string `json:"roles,omitempty" yaml:"roles,omitempty"`
}
