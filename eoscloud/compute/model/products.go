package model

// VMProduct represents a VM product (tag: vm-products).
type VMProduct struct {
	Self                string                `json:"_self,omitempty" yaml:"_self,omitempty"`
	ID                  string                `json:"id,omitempty" yaml:"id,omitempty"`
	Name                string                `json:"name,omitempty" yaml:"name,omitempty"`
	Active              bool                  `json:"active,omitempty" yaml:"active,omitempty"`
	Description         string                `json:"description,omitempty" yaml:"description,omitempty"`
	Architecture        ArchitectureEnum      `json:"architecture,omitempty" yaml:"architecture,omitempty"`
	Category            VMProductCategoryEnum `json:"category,omitempty" yaml:"category,omitempty"`
	CPUCount            int                   `json:"cpu_count,omitempty" yaml:"cpu_count,omitempty"`
	RAMSize             int                   `json:"ram_size,omitempty" yaml:"ram_size,omitempty"`
	NICSpeed            int                   `json:"nic_speed,omitempty" yaml:"nic_speed,omitempty"`
	BlockIds            []string              `json:"block_ids,omitempty" yaml:"block_ids,omitempty"`
	MicroCreditsPerUnit int                   `json:"micro_credits_per_unit,omitempty" yaml:"micro_credits_per_unit,omitempty"`
	Unit                UnitEnum              `json:"unit,omitempty" yaml:"unit,omitempty"`
	Weight              int                   `json:"weight,omitempty" yaml:"weight,omitempty"`
	Actions             []string              `json:"_actions,omitempty" yaml:"_actions,omitempty"`
}

// VolumeProduct represents a volume product (tag: volume-products).
type VolumeProduct struct {
	Self                string          `json:"_self,omitempty" yaml:"_self,omitempty"`
	ID                  string          `json:"id,omitempty" yaml:"id,omitempty"`
	Name                string          `json:"name,omitempty" yaml:"name,omitempty"`
	Active              bool            `json:"active,omitempty" yaml:"active,omitempty"`
	Description         string          `json:"description,omitempty" yaml:"description,omitempty"`
	StorageType         StorageTypeEnum `json:"storage_type,omitempty" yaml:"storage_type,omitempty"`
	IOPSLimit           int             `json:"iops_limit,omitempty" yaml:"iops_limit,omitempty"`
	BlockIds            []string        `json:"block_ids,omitempty" yaml:"block_ids,omitempty"`
	MicroCreditsPerUnit int             `json:"micro_credits_per_unit,omitempty" yaml:"micro_credits_per_unit,omitempty"`
	Unit                UnitEnum        `json:"unit,omitempty" yaml:"unit,omitempty"`
	Weight              int             `json:"weight,omitempty" yaml:"weight,omitempty"`
	Actions             []string        `json:"_actions,omitempty" yaml:"_actions,omitempty"`
}

// PublicIPProduct represents a public IP product (tag: public-ip-products).
type PublicIPProduct struct {
	Self                string       `json:"_self,omitempty" yaml:"_self,omitempty"`
	ID                  string       `json:"id,omitempty" yaml:"id,omitempty"`
	Name                string       `json:"name,omitempty" yaml:"name,omitempty"`
	Active              bool         `json:"active,omitempty" yaml:"active,omitempty"`
	Description         string       `json:"description,omitempty" yaml:"description,omitempty"`
	IPFamily            IpFamilyEnum `json:"ip_family,omitempty" yaml:"ip_family,omitempty"`
	BlockIds            []string     `json:"block_ids,omitempty" yaml:"block_ids,omitempty"`
	MicroCreditsPerUnit int          `json:"micro_credits_per_unit,omitempty" yaml:"micro_credits_per_unit,omitempty"`
	Unit                UnitEnum     `json:"unit,omitempty" yaml:"unit,omitempty"`
	Weight              int          `json:"weight,omitempty" yaml:"weight,omitempty"`
	Actions             []string     `json:"_actions,omitempty" yaml:"_actions,omitempty"`
}

// LoadBalancerProduct represents a load balancer product (tag: load-balancer-products).
type LoadBalancerProduct struct {
	Self                string       `json:"_self,omitempty" yaml:"_self,omitempty"`
	ID                  string       `json:"id,omitempty" yaml:"id,omitempty"`
	Name                string       `json:"name,omitempty" yaml:"name,omitempty"`
	Active              bool         `json:"active,omitempty" yaml:"active,omitempty"`
	Description         string       `json:"description,omitempty" yaml:"description,omitempty"`
	IPFamily            IpFamilyEnum `json:"ip_family,omitempty" yaml:"ip_family,omitempty"`
	BlockIds            []string     `json:"block_ids,omitempty" yaml:"block_ids,omitempty"`
	MicroCreditsPerUnit int          `json:"micro_credits_per_unit,omitempty" yaml:"micro_credits_per_unit,omitempty"`
	Unit                UnitEnum     `json:"unit,omitempty" yaml:"unit,omitempty"`
	Weight              int          `json:"weight,omitempty" yaml:"weight,omitempty"`
	Actions             []string     `json:"_actions,omitempty" yaml:"_actions,omitempty"`
}
