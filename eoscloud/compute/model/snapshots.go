package model

import "time"

// Snapshot models (tag: snapshots)

// Snapshot represents a VM snapshot.

type Snapshot struct {
	Self        string     `json:"_self,omitempty" yaml:"_self,omitempty"`
	ID          string     `json:"id,omitempty" yaml:"id,omitempty"`
	VM          string     `json:"vm_id" yaml:"vm_id"`
	Name        string     `json:"name" yaml:"name"`
	Description string     `json:"description" yaml:"description"`
	Status      StatusEnum `json:"status,omitempty" yaml:"status,omitempty"`
	CreatedAt   time.Time  `json:"created_at,omitempty" yaml:"created_at,omitempty"`
	UpdatedAt   time.Time  `json:"updated_at,omitempty" yaml:"updated_at,omitempty"`
	Actions     []string   `json:"_actions,omitempty" yaml:"_actions,omitempty"`
}

type SnapshotRequest struct {
	VM          string `json:"vm_id" yaml:"vm_id" shortflag:"V"`
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description" yaml:"description"`
}

type PatchedSnapshotRequest struct {
	Name        string `json:"name,omitempty" yaml:"name,omitempty"`
	Description string `json:"description,omitempty" yaml:"description,omitempty" shortflag:"D"`
}
