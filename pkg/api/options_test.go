package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOptions_Modifiers(t *testing.T) {
	custom := &http.Client{}
	c := NewClient(
		WithBaseURL("https://x"),
		WithHTTPClient(custom),
	).(*client)

	assert.Equal(t, custom, c.httpClient, "WithHTTPClient not applied")
}
