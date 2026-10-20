package model

import "time"

// PeeringGroup models (tag: peering-groups)

// PeeringGroup represents a peering group.

type PeeringGroup struct {
	Self        string     `json:"_self,omitempty" yaml:"_self,omitempty"`
	ID          string     `json:"id,omitempty" yaml:"id,omitempty"`
	Name        string     `json:"name" yaml:"name"`
	Description *string    `json:"description,omitempty" yaml:"description,omitempty"`
	DistrictID  string     `json:"district_id,omitempty" yaml:"district_id,omitempty"`
	NetworkIds  []string   `json:"network_ids,omitempty" yaml:"network_ids,omitempty"`
	Status      StatusEnum `json:"status,omitempty" yaml:"status,omitempty"`
	CreatedAt   time.Time  `json:"created_at,omitempty" yaml:"created_at,omitempty"`
	UpdatedAt   time.Time  `json:"updated_at,omitempty" yaml:"updated_at,omitempty"`
	Actions     []string   `json:"_actions,omitempty" yaml:"_actions,omitempty"`
}

type PeeringGroupRequest struct {
	Name        string   `json:"name" yaml:"name"`
	Description *string  `json:"description,omitempty" yaml:"description,omitempty" shortflag:"D"`
	DistrictID  string   `json:"district_id" yaml:"district_id"`
	NetworkIds  []string `json:"network_ids,omitempty" yaml:"network_ids,omitempty" shortflag:"N"`
}

type PatchedPeeringGroupRequest struct {
	Name        string  `json:"name,omitempty" yaml:"name,omitempty"`
	Description *string `json:"description,omitempty" yaml:"description,omitempty" shortflag:"D"`
}

// PeeringGroupMemberRequest represents a request to add or remove a Network from a peering group.
type PeeringGroupMemberRequest struct {
	Network string `json:"network_id" yaml:"network_id" shortflag:"N"`
}
