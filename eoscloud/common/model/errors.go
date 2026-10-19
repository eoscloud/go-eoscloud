package model

import (
	"fmt"
	"strings"
)

type APIErrorMessage struct {
	Path    string            `json:"path" yaml:"path"`
	Kind    string            `json:"kind" yaml:"kind"`
	Message string            `json:"message" yaml:"message"`
	Context map[string]string `json:"context,omitempty" yaml:"context,omitempty"`
}

func (e *APIErrorMessage) Error() string {
	return fmt.Sprintf("path='%s', kind='%s', message='%s', context='%v'", e.Path, e.Kind, e.Message, e.Context)
}

type APIError struct {
	Type       string            `json:"type" yaml:"type"`
	RequestID  string            `json:"request_id" yaml:"request_id"`
	Errors     []APIErrorMessage `json:"errors,omitempty" yaml:"errors,omitempty"`
	StatusCode int               `json:"status_code,omitempty" yaml:"status_code,omitempty"`
}

func (e *APIError) Error() string {
	errors := make([]string, len(e.Errors))
	for i, e := range e.Errors {
		errors[i] = e.Error()
	}
	return fmt.Sprintf("type='%s', request_id='%s', errors=[%s]", e.Type, e.RequestID, strings.Join(errors, "|"))
}

type ServerErrorResponse struct {
	Type      string `json:"type" yaml:"type"`
	RequestID string `json:"request_id" yaml:"request_id"`
}
