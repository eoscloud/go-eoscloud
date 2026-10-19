package model

import "encoding/json"

// Product models (tag: products)

// Product represents a compute product.

type Product struct {
	Self                string          `json:"_self,omitempty" yaml:"_self,omitempty"`
	ID                  string          `json:"id,omitempty" yaml:"id,omitempty"`
	BlockIds            []string        `json:"block_ids,omitempty" yaml:"block_ids,omitempty"`
	Weight              int             `json:"weight,omitempty" yaml:"weight,omitempty"`
	Unit                string          `json:"unit,omitempty" yaml:"unit,omitempty"`
	MicroCreditsPerUnit int             `json:"micro_credits_per_unit,omitempty" yaml:"micro_credits_per_unit,omitempty"`
	Name                string          `json:"name,omitempty" yaml:"name,omitempty"`
	Description         string          `json:"description,omitempty" yaml:"description,omitempty"`
	Kind                string          `json:"kind,omitempty" yaml:"kind,omitempty"`
	Active              bool            `json:"active,omitempty" yaml:"active,omitempty"`
	ProductData         json.RawMessage `json:"product_data,omitempty" yaml:"product_data,omitempty"`
	Actions             []string        `json:"_actions,omitempty" yaml:"_actions,omitempty"`
}
