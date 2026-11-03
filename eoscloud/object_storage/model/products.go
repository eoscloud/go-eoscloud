package model

// Product models

// BucketProduct represents a bucket product (tag: bucket-products).
type BucketProduct struct {
	ID                  string   `json:"id,omitempty" yaml:"id,omitempty"`
	Name                string   `json:"name,omitempty" yaml:"name,omitempty"`
	Description         string   `json:"description,omitempty" yaml:"description,omitempty"`
	Unit                UnitEnum `json:"unit,omitempty" yaml:"unit,omitempty"`
	MicroCreditsPerUnit int      `json:"micro_credits_per_unit,omitempty" yaml:"micro_credits_per_unit,omitempty"`
	Weight              int      `json:"weight,omitempty" yaml:"weight,omitempty"`
	Active              bool     `json:"active,omitempty" yaml:"active,omitempty"`
}
