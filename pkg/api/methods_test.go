package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMethods_Constants(t *testing.T) {
	assert.EqualValues(t, http.MethodGet, string(GET))
	assert.EqualValues(t, http.MethodPost, string(POST))
	assert.EqualValues(t, http.MethodPut, string(PUT))
	assert.EqualValues(t, http.MethodPatch, string(PATCH))
	assert.EqualValues(t, http.MethodDelete, string(DELETE))
}
