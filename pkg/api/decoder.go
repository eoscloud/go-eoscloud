package api

import (
	"encoding/json"
	"io"
)

// Decoder defines a function to decode a response body
type Decoder[T any] func(body io.ReadCloser) (T, error)

// NewJSONDecoder returns a decoder that unmarshals the response body into out and returns it.
func NewJSONDecoder[T any](out *T) Decoder[any] {
	return func(body io.ReadCloser) (any, error) {
		err := json.NewDecoder(body).Decode(out)
		return out, err
	}
}

// NoResultDecoder returns a decoder that discards the response body and returns nil.
func NoResultDecoder() Decoder[any] {
	return func(body io.ReadCloser) (any, error) { return nil, nil }
}
