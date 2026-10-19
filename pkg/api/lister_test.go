package api

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type listItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func TestLister_List(t *testing.T) {
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
			_, _ = w.Write([]byte(`{
				"count": 2,
				"next": null,
				"previous": null,
				"results": [
					{"id":"1","name":"Alice"},
					{"id":"2","name":"Bob"}
				]
			}`))
		}))
		defer srv.Close()

		c := NewClient(WithBaseURL(srv.URL))
		l := NewLister[listItem](c, "svc", "/v1/items/{x}", map[string]string{"x": "X"})

		q := url.Values{}
		q.Set("limit", "50")
		q.Add("tag", "blue")

		out, err := l.List(context.Background(), q)
		require.NoError(t, err, fmt.Sprintf("unexpected error: %v", err))
		require.NotNil(t, out, "expected non-nil page")

		assert.EqualValues(t, http.MethodGet, gotMethod, "method mismatch")
		assert.EqualValues(t, "/v1/items/X", gotPath, fmt.Sprintf("path mismatch: %q", gotPath))
		assert.EqualValues(t, "50", gotQuery.Get("limit"), "query param limit missing")
		assert.EqualValues(t, []string{"blue"}, gotQuery["tag"], "query param tag mismatch")

		assert.EqualValues(t, 2, out.Count)
		assert.Nil(t, out.Next)
		assert.Nil(t, out.Previous)
		if assert.Len(t, out.Results, 2, "results length") {
			assert.EqualValues(t, "1", out.Results[0].ID)
			assert.EqualValues(t, "Alice", out.Results[0].Name)
			assert.EqualValues(t, "2", out.Results[1].ID)
			assert.EqualValues(t, "Bob", out.Results[1].Name)
		}
	})

	t.Run("Non200_Status_ReturnsError", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("bad request"))
		}))
		defer srv.Close()

		c := NewClient(WithBaseURL(srv.URL))
		l := NewLister[listItem](c, "svc", "/v1/items/{x}", map[string]string{"x": "X"})

		out, err := l.List(context.Background(), nil)
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
		l := NewLister[listItem](c, "svc", "/v1/items/{x}", map[string]string{"x": "X"})

		out, err := l.List(context.Background(), nil)
		require.Error(t, err, "expected JSON decode error")
		assert.Nil(t, out, fmt.Sprintf("expected nil output on error, got: %#v", out))
	})

	t.Run("MissingPathParam_ReturnsError", func(t *testing.T) {
		// Server should not receive a request because client errors before sending
		received := false
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			received = true
			w.WriteHeader(http.StatusOK)
		}))
		defer srv.Close()

		c := NewClient(WithBaseURL(srv.URL))
		l := NewLister[listItem](c, "svc", "/v1/items/{x}", map[string]string{}) // missing {x}

		out, err := l.List(context.Background(), nil)
		require.Error(t, err, "expected error due to missing path parameter")
		assert.Nil(t, out, fmt.Sprintf("expected nil output on error, got: %#v", out))
		assert.False(t, received, "server should not have been called")
	})
}
