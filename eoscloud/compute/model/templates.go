package model

import "time"

// Template models (tag: templates)

// Template represents a VM template.

type Template struct {
	Self        string     `json:"_self,omitempty" yaml:"_self,omitempty"`
	ID          string     `json:"id,omitempty" yaml:"id,omitempty"`
	Block       string     `json:"block_id,omitempty" yaml:"block_id,omitempty"`
	Name        string     `json:"name" yaml:"name"`
	Description *string    `json:"description,omitempty" yaml:"description,omitempty"`
	DiskSize    int        `json:"disk_size,omitempty" yaml:"disk_size,omitempty"`
	Image       string     `json:"image_id,omitempty" yaml:"image_id,omitempty"`
	Status      StatusEnum `json:"status,omitempty" yaml:"status,omitempty"`
	CreatedAt   time.Time  `json:"created_at,omitempty" yaml:"created_at,omitempty"`
	UpdatedAt   time.Time  `json:"updated_at,omitempty" yaml:"updated_at,omitempty"`
	Actions     []string   `json:"_actions,omitempty" yaml:"_actions,omitempty"`
}

// TemplateRequest represents the payload for creating a template.

type TemplateRequest struct {
	Name        string  `json:"name" yaml:"name"`
	Description *string `json:"description,omitempty" yaml:"description,omitempty" shortflag:"D"`
	VM          string  `json:"vm_id" yaml:"vm_id" shortflag:"V"`
}

type PatchedTemplateRequest struct {
	Name        string  `json:"name,omitempty" yaml:"name,omitempty"`
	Description *string `json:"description,omitempty" yaml:"description,omitempty" shortflag:"D"`
}
