package model

// Block models (tag: blocks)

// Block represents a block (availability zone).

type Block struct {
	Self         string           `json:"_self,omitempty" yaml:"_self,omitempty"`
	ID           string           `json:"id,omitempty" yaml:"id,omitempty"`
	RegionID     string           `json:"region_id" yaml:"region_id"`
	Name         string           `json:"name,omitempty" yaml:"name,omitempty"`
	District     int              `json:"district,omitempty" yaml:"district,omitempty"`
	Identifier   string           `json:"identifier,omitempty" yaml:"identifier,omitempty"`
	Architecture ArchitectureEnum `json:"architecture,omitempty" yaml:"architecture,omitempty"`
	Optimization OptimizationEnum `json:"optimization,omitempty" yaml:"optimization,omitempty"`
	Actions      []string         `json:"_actions,omitempty" yaml:"_actions,omitempty"`
}
