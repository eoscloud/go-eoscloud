package model

// VM Image models (tag: vm-images)

// VMImage represents a VM image.

type VMImage struct {
	Self         string           `json:"_self,omitempty" yaml:"_self,omitempty"`
	ID           string           `json:"id,omitempty" yaml:"id,omitempty"`
	Name         string           `json:"name,omitempty" yaml:"name,omitempty"`
	Architecture ArchitectureEnum `json:"architecture,omitempty" yaml:"architecture,omitempty"`
	CloudInit    bool             `json:"cloud_init,omitempty" yaml:"cloud_init,omitempty"`
	Active       bool             `json:"active,omitempty" yaml:"active,omitempty"`
	Actions      []string         `json:"_actions,omitempty" yaml:"_actions,omitempty"`
}
