package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test the concrete helpers directly
func TestHTTPStatusHelpers(t *testing.T) {
	t.Run("IsHTTPStatusCreated", func(t *testing.T) {
		assert.True(t, IsHTTPStatusCreated(http.StatusCreated))
		assert.False(t, IsHTTPStatusCreated(http.StatusOK))
		assert.False(t, IsHTTPStatusCreated(http.StatusAccepted))
		assert.False(t, IsHTTPStatusCreated(http.StatusNoContent))
	})

	t.Run("IsHTTPStatusOK", func(t *testing.T) {
		assert.True(t, IsHTTPStatusOK(http.StatusOK))
		assert.False(t, IsHTTPStatusOK(http.StatusCreated))
		assert.False(t, IsHTTPStatusOK(http.StatusAccepted))
		assert.False(t, IsHTTPStatusOK(http.StatusNoContent))
	})

	t.Run("IsHTTPStatusNoContent", func(t *testing.T) {
		assert.True(t, IsHTTPStatusNoContent(http.StatusNoContent))
		assert.False(t, IsHTTPStatusNoContent(http.StatusOK))
		assert.False(t, IsHTTPStatusNoContent(http.StatusCreated))
		assert.False(t, IsHTTPStatusNoContent(http.StatusAccepted))
	})

	t.Run("IsHTTPStatusAccepted", func(t *testing.T) {
		assert.True(t, IsHTTPStatusAccepted(http.StatusAccepted))
		assert.False(t, IsHTTPStatusAccepted(http.StatusOK))
		assert.False(t, IsHTTPStatusAccepted(http.StatusCreated))
		assert.False(t, IsHTTPStatusAccepted(http.StatusNoContent))
	})

	t.Run("IsHTTPStatusOKOrAccepted", func(t *testing.T) {
		assert.True(t, IsHTTPStatusOKOrAccepted(http.StatusOK))
		assert.True(t, IsHTTPStatusOKOrAccepted(http.StatusAccepted))
		assert.False(t, IsHTTPStatusOKOrAccepted(http.StatusCreated))
		assert.False(t, IsHTTPStatusOKOrAccepted(http.StatusNoContent))
		assert.False(t, IsHTTPStatusOKOrAccepted(http.StatusBadRequest))
	})
}

// Demonstrate the function type can be used with the helpers
func TestHTTPStatusCheckType(t *testing.T) {
	var check HTTPStatusCheck

	check = IsHTTPStatusOK
	assert.True(t, check(http.StatusOK))
	assert.False(t, check(http.StatusCreated))

	check = IsHTTPStatusCreated
	assert.True(t, check(http.StatusCreated))
	assert.False(t, check(http.StatusNoContent))

	check = IsHTTPStatusNoContent
	assert.True(t, check(http.StatusNoContent))
	assert.False(t, check(http.StatusAccepted))

	check = IsHTTPStatusAccepted
	assert.True(t, check(http.StatusAccepted))
	assert.False(t, check(http.StatusOK))
}
