package model

import "time"

// ACL models (tag: acls)

type ACL struct {
	Self        string     `json:"_self,omitempty" yaml:"_self,omitempty"`
	ID          string     `json:"id,omitempty" yaml:"id,omitempty"`
	Block       string     `json:"block_id" yaml:"block_id"`
	Name        string     `json:"name" yaml:"name"`
	Description string     `json:"description,omitempty" yaml:"description,omitempty"`
	Default     bool       `json:"default,omitempty" yaml:"default,omitempty"`
	Status      StatusEnum `json:"status,omitempty" yaml:"status,omitempty"`
	CreatedAt   time.Time  `json:"created_at,omitempty" yaml:"created_at,omitempty"`
	UpdatedAt   time.Time  `json:"updated_at,omitempty" yaml:"updated_at,omitempty"`
	Actions     []string   `json:"_actions,omitempty" yaml:"_actions,omitempty"`
}

type ACLRequest struct {
	Block       string `json:"block_id" yaml:"block_id" shortflag:"B"`
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description,omitempty" yaml:"description,omitempty" shortflag:"D"`
}

type PatchedACLRequest struct {
	Description string `json:"description,omitempty" yaml:"description,omitempty" shortflag:"D"`
}
