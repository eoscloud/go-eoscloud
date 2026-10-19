package model

type Action struct {
	Kind    string            `json:"kind" yaml:"kind"`
	Message string            `json:"message" yaml:"message"`
	Context map[string]string `json:"context,omitempty" yaml:"context,omitempty"`
}

type Actions struct {
	Available   []string            `json:"available" yaml:"available"`
	Unavailable map[string][]Action `json:"unavailable" yaml:"unavailable"`
}

type ObjectOptionsResponse struct {
	Name        string   `json:"name" yaml:"name"`
	Description string   `json:"description" yaml:"description"`
	Renders     []string `json:"renders" yaml:"renders"`
	Parses      []string `json:"parses" yaml:"parses"`
	Actions     Actions  `json:"actions" yaml:"actions"`
}
