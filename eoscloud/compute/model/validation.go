package model

// Shared validation/error models (not tied to a specific tag)

// ValidationError represents an API validation error response.

type ValidationError struct {
	NonFieldErrors []string `json:"non_field_errors,omitempty" yaml:"non_field_errors,omitempty"`
	Timeout        []string `json:"timeout,omitempty" yaml:"timeout,omitempty"`
	Detail         string   `json:"detail,omitempty" yaml:"detail,omitempty"`
}

// ConflictError represents an API 409 conflict error response.

type ConflictError struct {
	Detail string `json:"detail" yaml:"detail"`
}
