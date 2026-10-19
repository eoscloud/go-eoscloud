package model

// District models (tag: districts)

type District struct {
	ID              string `json:"id,omitempty" yaml:"id,omitempty"`
	RegionID        string `json:"region_id,omitempty" yaml:"region_id,omitempty"`
	Slug            string `json:"slug,omitempty" yaml:"slug,omitempty"`
	BlockStorageURL string `json:"block_storage_url,omitempty" yaml:"block_storage_url,omitempty"`
}
