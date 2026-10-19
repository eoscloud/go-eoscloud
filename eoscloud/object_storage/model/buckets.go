package model

import (
	"encoding/json"
	"time"
)

// Bucket models (tag: buckets)

type Bucket struct {
	Self                 string          `json:"_self,omitempty" yaml:"_self,omitempty"`
	Actions              []string        `json:"_actions,omitempty" yaml:"_actions,omitempty"`
	ID                   string          `json:"id,omitempty" yaml:"id,omitempty"`
	Name                 string          `json:"name" yaml:"name"`
	URL                  string          `json:"url,omitempty" yaml:"url,omitempty"`
	ProductID            string          `json:"product_id,omitempty" yaml:"product_id,omitempty"`
	DistrictID           string          `json:"district_id,omitempty" yaml:"district_id,omitempty"`
	Description          *string         `json:"description,omitempty" yaml:"description,omitempty"`
	Quota                int             `json:"quota" yaml:"quota"`
	Status               StatusEnum      `json:"status,omitempty" yaml:"status,omitempty"`
	Utilization          json.RawMessage `json:"utilization,omitempty" yaml:"utilization,omitempty"`
	UtilizationUpdatedAt time.Time       `json:"utilization_updated_at,omitempty" yaml:"utilization_updated_at,omitempty"`
	CreatedAt            time.Time       `json:"created_at,omitempty" yaml:"created_at,omitempty"`
	UpdatedAt            time.Time       `json:"updated_at,omitempty" yaml:"updated_at,omitempty"`
}

type BucketRequest struct {
	Name        string  `json:"name" yaml:"name"`
	ProductID   string  `json:"product_id" yaml:"product_id" shortflag:"P"`
	DistrictID  string  `json:"district_id" yaml:"district_id"`
	Description *string `json:"description,omitempty" yaml:"description,omitempty" shortflag:"D"`
	Quota       int     `json:"quota" yaml:"quota"`
}

type PatchedBucketRequest struct {
	Description *string `json:"description,omitempty" yaml:"description,omitempty" shortflag:"D"`
}

type BucketGrowRequest struct {
	Quota int `json:"quota" yaml:"quota"`
}
