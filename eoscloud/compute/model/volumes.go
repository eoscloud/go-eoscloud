package model

import (
	"time"
)

// Volume models (tag: volumes)

// ListVolume represents a volume to be used in Lists
type ListVolume struct {
	Self                 string          `json:"_self,omitempty" yaml:"_self,omitempty"`
	ID                   string          `json:"id,omitempty" yaml:"id,omitempty"`
	Block                string          `json:"block_id" yaml:"block_id"`
	Product              string          `json:"product_id" yaml:"product_id"`
	VM                   *string         `json:"vm_id,omitempty" yaml:"vm_id,omitempty"`
	Name                 string          `json:"name" yaml:"name"`
	Description          string          `json:"description,omitempty" yaml:"description,omitempty"`
	Capacity             int             `json:"capacity" yaml:"capacity"`
	CapacityUtilization  int64           `json:"capacity_utilization,omitempty" yaml:"capacity_utilization,omitempty"`
	UtilizationUpdatedAt time.Time       `json:"utilization_updated_at,omitempty" yaml:"utilization_updated_at,omitempty"`
	RootDisk             bool            `json:"root_disk,omitempty" yaml:"root_disk,omitempty"`
	StorageType          StorageTypeEnum `json:"storage_type,omitempty" yaml:"storage_type,omitempty"`
	Status               StatusEnum      `json:"status,omitempty" yaml:"status,omitempty"`
	CreatedAt            time.Time       `json:"created_at,omitempty" yaml:"created_at,omitempty"`
	UpdatedAt            time.Time       `json:"updated_at,omitempty" yaml:"updated_at,omitempty"`
	Actions              []string        `json:"_actions,omitempty" yaml:"_actions,omitempty"`
}

// VolumeDetail represents a volume as defined in compute-api.yaml
type VolumeDetail struct {
	Self                 string          `json:"_self,omitempty" yaml:"_self,omitempty"`
	ID                   string          `json:"id,omitempty" yaml:"id,omitempty"`
	Block                string          `json:"block_id" yaml:"block_id"`
	Product              string          `json:"product_id" yaml:"product_id"`
	VM                   *string         `json:"vm_id,omitempty" yaml:"vm_id,omitempty"`
	Name                 string          `json:"name" yaml:"name"`
	Label                string          `json:"label,omitempty" yaml:"label,omitempty"`
	Description          string          `json:"description,omitempty" yaml:"description,omitempty"`
	Capacity             int             `json:"capacity" yaml:"capacity"`
	CapacityUtilization  int64           `json:"capacity_utilization,omitempty" yaml:"capacity_utilization,omitempty"`
	UtilizationUpdatedAt time.Time       `json:"utilization_updated_at,omitempty" yaml:"utilization_updated_at,omitempty"`
	RootDisk             bool            `json:"root_disk,omitempty" yaml:"root_disk,omitempty"`
	StorageType          StorageTypeEnum `json:"storage_type,omitempty" yaml:"storage_type,omitempty"`
	Status               StatusEnum      `json:"status,omitempty" yaml:"status,omitempty"`
	CreatedAt            time.Time       `json:"created_at,omitempty" yaml:"created_at,omitempty"`
	UpdatedAt            time.Time       `json:"updated_at,omitempty" yaml:"updated_at,omitempty"`
	Actions              []string        `json:"_actions,omitempty" yaml:"_actions,omitempty"`
}

// VolumeRequest represents a request to create a volume
type VolumeRequest struct {
	Block       string          `json:"block_id" yaml:"block_id" shortflag:"B"`
	Name        string          `json:"name" yaml:"name"`
	Description string          `json:"description,omitempty" yaml:"description,omitempty" shortflag:"D"`
	Capacity    int             `json:"capacity" yaml:"capacity" shortflag:"C"`
	StorageType StorageTypeEnum `json:"storage_type,omitempty" yaml:"storage_type,omitempty"`
}

// PatchedVolumeRequest represents a request to partially update a volume
type PatchedVolumeRequest struct {
	Name        string `json:"name,omitempty" yaml:"name,omitempty"`
	Description string `json:"description,omitempty" yaml:"description,omitempty" shortflag:"D"`
}

// VolumeAttachRequest represents a request to attach a volume to a VM
type VolumeAttachRequest struct {
	VM string `json:"vm_id" yaml:"vm_id" shortflag:"V"`
}

// GrowVolumeRequest represents a request to grow a volume
type GrowVolumeRequest struct {
	Capacity int `json:"capacity,omitempty" yaml:"capacity,omitempty" shortflag:"C"`
}
