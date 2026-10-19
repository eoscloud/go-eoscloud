package model

import "time"

// Network models (tag: networks)

type Network struct {
	Self        string     `json:"_self,omitempty" yaml:"_self,omitempty"`
	ID          string     `json:"id,omitempty" yaml:"id,omitempty"`
	Block       string     `json:"block_id" yaml:"block_id"`
	Name        string     `json:"name" yaml:"name"`
	Description string     `json:"description,omitempty" yaml:"description,omitempty"`
	Status      StatusEnum `json:"status,omitempty" yaml:"status,omitempty"`
	DNSZone     string     `json:"dns_zone,omitempty" yaml:"dns_zone,omitempty"`
	V4Prefix    *string    `json:"v4_prefix,omitempty" yaml:"v4_prefix,omitempty"`
	V6Prefix    *string    `json:"v6_prefix,omitempty" yaml:"v6_prefix,omitempty"`
	ACLIds      []string   `json:"acl_ids,omitempty" yaml:"acl_ids,omitempty"`
	CreatedAt   time.Time  `json:"created_at,omitempty" yaml:"created_at,omitempty"`
	UpdatedAt   time.Time  `json:"updated_at,omitempty" yaml:"updated_at,omitempty"`
	Actions     []string   `json:"_actions,omitempty" yaml:"_actions,omitempty"`
}

type NetworkRequest struct {
	Block       string   `json:"block_id" yaml:"block_id" shortflag:"B"`
	Name        string   `json:"name" yaml:"name"`
	Description string   `json:"description,omitempty" yaml:"description,omitempty" shortflag:"D"`
	V4Prefix    *string  `json:"v4_prefix,omitempty" yaml:"v4_prefix,omitempty" shortflag:"4"`
	V6Prefix    *string  `json:"v6_prefix,omitempty" yaml:"v6_prefix,omitempty" shortflag:"6"`
	ACLIds      []string `json:"acl_ids,omitempty" yaml:"acl_ids,omitempty" shortflag:"A"`
}

type PatchedNetworkRequest struct {
	Description string   `json:"description,omitempty" yaml:"description,omitempty" shortflag:"D"`
	ACLIds      []string `json:"acl_ids,omitempty" yaml:"acl_ids,omitempty" shortflag:"A"`
}
