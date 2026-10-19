package model

// Role represents an assignable role in the organization service.
type Role struct {
	Role    string `json:"role" yaml:"role"`
	Title   string `json:"title" yaml:"title"`
	Service string `json:"service" yaml:"service"`
	Scope   string `json:"scope" yaml:"scope"`
}
