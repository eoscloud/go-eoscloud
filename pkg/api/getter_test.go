package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type getOut struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func TestGetter_Get(t *testing.T) {
	t.Run("Success_WithQuery", func(t *testing.T) {
		var gotPath string
		var gotMethod string
		var gotQuery url.Values

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method
			gotQuery = r.URL.Query()

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id":"123","name":"Alice"}`))
		}))
		defer srv.Close()

		c := NewClient(WithBaseURL(srv.URL))
		g := NewGetter[getOut](c, "svc", "/v1/items/{x}", map[string]string{"x": "X"})

		q := url.Values{}
		q.Set("a", "1")
		q.Add("b", "two")

		out, err := g.Get(context.Background(), "123", q)
		require.NoError(t, err, fmt.Sprintf("unexpected error: %v", err))
		require.NotNil(t, out, "expected non-nil output")

		assert.EqualValues(t, http.MethodGet, gotMethod, "method mismatch")
		assert.EqualValues(t, "/v1/items/X/123/", gotPath, fmt.Sprintf("path mismatch: %q", gotPath))
		assert.EqualValues(t, "1", gotQuery.Get("a"), "query param a missing")
		assert.EqualValues(t, []string{"two"}, gotQuery["b"], "query param b mismatch")
		assert.EqualValues(t, "123", out.ID)
		assert.EqualValues(t, "Alice", out.Name)
	})

	t.Run("Non200_Status_ReturnsError", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte("not found"))
		}))
		defer srv.Close()

		c := NewClient(WithBaseURL(srv.URL))
		g := NewGetter[getOut](c, "svc", "/v1/items/{x}", map[string]string{"x": "X"})

		out, err := g.Get(context.Background(), "nope", nil)
		require.Error(t, err, "expected error on non-200 response")
		assert.Nil(t, out, fmt.Sprintf("expected nil output on error, got: %#v", out))
	})

	t.Run("InvalidJSON_ReturnsError", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("{")) // invalid JSON
		}))
		defer srv.Close()

		c := NewClient(WithBaseURL(srv.URL))
		g := NewGetter[getOut](c, "svc", "/v1/items/{x}", map[string]string{"x": "X"})

		out, err := g.Get(context.Background(), "123", nil)
		require.Error(t, err, "expected JSON decode error")
		assert.Nil(t, out, fmt.Sprintf("expected nil output on error, got: %#v", out))
	})

	t.Run("MissingPathParam_ReturnsError", func(t *testing.T) {
		// Server never receives a request because client should error before sending
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// If we get here, fail the test to ensure path param validation happened
			b, _ := io.ReadAll(r.Body)
			_ = r.Body.Close()
			_ = json.Unmarshal(b, &struct{}{})
			w.WriteHeader(http.StatusOK)
		}))
		defer srv.Close()

		c := NewClient(WithBaseURL(srv.URL))
		g := NewGetter[getOut](c, "svc", "/v1/items/{x}", map[string]string{}) // missing {x}

		out, err := g.Get(context.Background(), "123", nil)
		require.Error(t, err, "expected error due to missing path parameter")
		assert.Nil(t, out, fmt.Sprintf("expected nil output on error, got: %#v", out))
	})
}
