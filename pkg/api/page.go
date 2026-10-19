package api

// Page represents a generic paged result from the API
type Page[T any] struct {
	Count    int      `json:"count"`
	Next     *string  `json:"next"`
	Previous *string  `json:"previous"`
	Results  []T      `json:"results"`
	Actions  []string `json:"_actions,omitempty" yaml:"_actions,omitempty"`
}
