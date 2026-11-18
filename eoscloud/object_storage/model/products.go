package model

// Product models

// BucketProduct represents a bucket product (tag: bucket-products).
type BucketProduct struct {
	Self                string   `json:"_self,omitempty" yaml:"_self,omitempty"`
	ID                  string   `json:"id,omitempty" yaml:"id,omitempty"`
	Name                string   `json:"name,omitempty" yaml:"name,omitempty"`
	Description         string   `json:"description,omitempty" yaml:"description,omitempty"`
	Unit                UnitEnum `json:"unit,omitempty" yaml:"unit,omitempty"`
	MicroCreditsPerUnit int      `json:"micro_credits_per_unit,omitempty" yaml:"micro_credits_per_unit,omitempty"`
	Weight              int      `json:"weight,omitempty" yaml:"weight,omitempty"`
	Active              bool     `json:"active,omitempty" yaml:"active,omitempty"`
	MinQuotaBytes       int64    `json:"min_quota_bytes,omitempty" yaml:"min_quota_bytes,omitempty"`
	MaxQuotaBytes       int64    `json:"max_quota_bytes,omitempty" yaml:"max_quota_bytes,omitempty"`
	DefaultQuotaBytes   int64    `json:"default_quota_bytes,omitempty" yaml:"default_quota_bytes,omitempty"`
	Actions             []string `json:"_actions,omitempty" yaml:"_actions,omitempty"`
}
