package api

import (
	"bytes"
	"encoding/json"
	"io"
)

// Encoder defines a function how to encode request or response bodies.
type Encoder func() (contentType string, reader io.Reader, err error)

// NewJSONEncoder returns an Encoder that serialises data as JSON.
func NewJSONEncoder[T any](data *T) Encoder {
	return func() (string, io.Reader, error) {
		payload := new(bytes.Buffer)
		err := json.NewEncoder(payload).Encode(data)
		return "application/json", payload, err
	}
}

// NoBodyEncoder returns an Encoder that sends no request body.
func NoBodyEncoder() Encoder {
	return func() (string, io.Reader, error) {
		return "application/json", nil, nil
	}
}
