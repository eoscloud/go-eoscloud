package api

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeleter_Delete(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var gotPath string
		var gotMethod string

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method
			w.WriteHeader(http.StatusNoContent)
		}))
		defer srv.Close()

		c := NewClient(WithBaseURL(srv.URL))
		d := NewDeleter(c, "svc", "/v1/items/{x}", map[string]string{"x": "X"})

		err := d.Delete(context.Background(), "123")
		require.NoError(t, err, fmt.Sprintf("unexpected error: %v", err))

		assert.EqualValues(t, http.MethodDelete, gotMethod, "method mismatch")
		assert.EqualValues(t, "/v1/items/X/123/", gotPath, fmt.Sprintf("path mismatch: %q", gotPath))
	})

	t.Run("Non204_Status_ReturnsError", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("bad request"))
		}))
		defer srv.Close()

		c := NewClient(WithBaseURL(srv.URL))
		d := NewDeleter(c, "svc", "/v1/items/{x}", map[string]string{"x": "X"})

		err := d.Delete(context.Background(), "123")
		require.Error(t, err, "expected error on non-204 response")
	})

	t.Run("MissingPathParam_ReturnsError", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}))
		defer srv.Close()

		c := NewClient(WithBaseURL(srv.URL))
		// Intentionally omit required path parameter {x}
		d := NewDeleter(c, "svc", "/v1/items/{x}", map[string]string{})

		err := d.Delete(context.Background(), "123")
		require.Error(t, err, "expected error due to missing path parameter")
	})
}
