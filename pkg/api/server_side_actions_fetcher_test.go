package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServerSideActionsFetcher_Actions(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var gotPath string
		var gotMethod string

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
				"name": "test-object",
				"description": "test-description",
				"actions": {
					"available": [ "test-action" ]
				}
			}`))
		}))
		defer srv.Close()

		c := NewClient(WithBaseURL(srv.URL))
		ai := NewServerSideActionsFetcher(c, "svc", "/v1/items/{x}", map[string]string{"x": "X"})

		out, err := ai.Actions(context.Background(), "123")
		require.NoError(t, err)
		require.NotNil(t, out)

		assert.Equal(t, http.MethodOptions, gotMethod)
		assert.Equal(t, "/v1/items/X/123/", gotPath)
		assert.Equal(t, "test-object", out.Name)
		assert.Equal(t, "test-description", out.Description)
		require.Len(t, out.Actions.Available, 1)
		assert.Equal(t, "test-action", out.Actions.Available[0])
	})

	t.Run("Non200_Status_ReturnsError", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer srv.Close()

		c := NewClient(WithBaseURL(srv.URL))
		ai := NewServerSideActionsFetcher(c, "svc", "/v1/items/{x}", map[string]string{"x": "X"})

		out, err := ai.Actions(context.Background(), "nope")
		require.Error(t, err)
		assert.Nil(t, out)
	})

	t.Run("InvalidJSON_ReturnsError", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("{"))
		}))
		defer srv.Close()

		c := NewClient(WithBaseURL(srv.URL))
		ai := NewServerSideActionsFetcher(c, "svc", "/v1/items/{x}", map[string]string{"x": "X"})

		out, err := ai.Actions(context.Background(), "123")
		require.Error(t, err)
		assert.Nil(t, out)
	})

	t.Run("MissingPathParam_ReturnsError", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer srv.Close()

		c := NewClient(WithBaseURL(srv.URL))
		ai := NewServerSideActionsFetcher(c, "svc", "/v1/items/{x}", map[string]string{})

		out, err := ai.Actions(context.Background(), "123")
		require.Error(t, err)
		assert.Nil(t, out)
	})
}
