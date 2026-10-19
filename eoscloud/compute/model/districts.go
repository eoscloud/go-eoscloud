package model

// District models (tag: districts)

// District represents a district (availability zone).

type District struct {
	Self       string `json:"_self,omitempty" yaml:"_self,omitempty"`
	ID         string `json:"id" yaml:"id"`
	RegionID   string `json:"region_id" yaml:"region_id"`
	DataCenter int    `json:"datacenter" yaml:"datacenter"`
	Slug       string `json:"slug" yaml:"slug"`
}
