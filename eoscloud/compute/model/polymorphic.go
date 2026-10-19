package model

import (
	"bytes"
	"encoding/json"

	"go.yaml.in/yaml/v3"
)

// peekType reads only the "type" discriminator of a JSON object.
func peekType(data []byte) (string, error) {
	var head struct {
		Type string `json:"type"`
	}
	err := json.Unmarshal(data, &head)
	return head.Type, err
}

// peekYAMLType reads only the "type" discriminator of a YAML mapping.
func peekYAMLType(node *yaml.Node) (string, error) {
	var head struct {
		Type string `yaml:"type"`
	}
	err := node.Decode(&head)
	return head.Type, err
}

func isJSONNull(data []byte) bool {
	return bytes.Equal(bytes.TrimSpace(data), []byte("null"))
}

func isYAMLNull(node *yaml.Node) bool {
	return node.Tag == "!!null"
}
