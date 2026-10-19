package model

import "time"

// VMGroup models (tag: vm-groups)

// VMGroup represents a VM group.

type VMGroup struct {
	Self        string     `json:"_self,omitempty" yaml:"_self,omitempty"`
	ID          string     `json:"id,omitempty" yaml:"id,omitempty"`
	Name        string     `json:"name" yaml:"name"`
	Description *string    `json:"description,omitempty" yaml:"description,omitempty"`
	DistrictID  string     `json:"district_id,omitempty" yaml:"district_id,omitempty"`
	VMIds       []string   `json:"vm_ids,omitempty" yaml:"vm_ids,omitempty"`
	Status      StatusEnum `json:"status,omitempty" yaml:"status,omitempty"`
	CreatedAt   time.Time  `json:"created_at,omitempty" yaml:"created_at,omitempty"`
	UpdatedAt   time.Time  `json:"updated_at,omitempty" yaml:"updated_at,omitempty"`
	Actions     []string   `json:"_actions,omitempty" yaml:"_actions,omitempty"`
}

type VMGroupRequest struct {
	Name        string  `json:"name" yaml:"name"`
	Description *string `json:"description,omitempty" yaml:"description,omitempty" shortflag:"D"`
	DistrictID  string  `json:"district_id" yaml:"district_id"`
}

type PatchedVMGroupRequest struct {
	Name        string  `json:"name,omitempty" yaml:"name,omitempty"`
	Description *string `json:"description,omitempty" yaml:"description,omitempty" shortflag:"D"`
}

// VMGroupMemberRequest represents a request to add or remove a VM from a VM group.
type VMGroupMemberRequest struct {
	VM string `json:"vm_id" yaml:"vm_id" shortflag:"V"`
}
