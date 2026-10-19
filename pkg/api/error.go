package api

import (
	"fmt"
	"net/http"
)

type UnexpectedHttpStatusError struct {
	StatusCode int
	Method     Method
	URL        string
	Body       string
}

func (e *UnexpectedHttpStatusError) Error() string {
	return fmt.Sprintf(
		"unexpected status: %d on %s %s returned: %s",
		e.StatusCode,
		e.Method,
		e.URL,
		e.Body,
	)
}

func (e *UnexpectedHttpStatusError) IsUnauthorized() bool {
	return e.StatusCode == http.StatusUnauthorized
}

func (e *UnexpectedHttpStatusError) IsForbidden() bool {
	return e.StatusCode == http.StatusForbidden
}
