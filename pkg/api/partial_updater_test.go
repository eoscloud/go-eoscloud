package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type pupdateIn struct {
	Name string `json:"name"`
}

type pupdateOut struct {
	ID  string `json:"id"`
	Msg string `json:"msg"`
}

func TestPartialUpdater_UpdatePartial(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var gotPath string
		var gotMethod string
		var gotBody map[string]any

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method
			b, _ := io.ReadAll(r.Body)
			_ = r.Body.Close()
			_ = json.Unmarshal(b, &gotBody)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id":"42","msg":"patched"}`))
		}))
		defer srv.Close()

		c := NewClient(WithBaseURL(srv.URL))
		pu := NewPartialUpdater[pupdateIn, pupdateOut](c, "svc", "/v1/items/{x}", map[string]string{"x": "X"})

		in := pupdateIn{Name: "Alice"}
		out, err := pu.UpdatePartial(context.Background(), "123", &in)
		require.NoError(t, err, fmt.Sprintf("unexpected error: %v", err))
		require.NotNil(t, out, "expected non-nil output")

		assert.EqualValues(t, http.MethodPatch, gotMethod, "method mismatch")
		assert.EqualValues(t, "/v1/items/X/123/", gotPath, fmt.Sprintf("path mismatch: %q", gotPath))
		assert.EqualValues(t, "Alice", gotBody["name"], fmt.Sprintf("payload mismatch: %#v", gotBody))
		assert.EqualValues(t, "42", out.ID)
		assert.EqualValues(t, "patched", out.Msg)
	})

	t.Run("Accepted_AsyncResponse", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusAccepted)
			_, _ = w.Write([]byte(`{"id":"42","msg":"updating"}`))
		}))
		defer srv.Close()

		c := NewClient(WithBaseURL(srv.URL))
		pu := NewPartialUpdater[pupdateIn, pupdateOut](c, "svc", "/v1/items/{x}", map[string]string{"x": "X"})

		in := pupdateIn{Name: "Alice"}
		out, err := pu.UpdatePartial(context.Background(), "123", &in)
		require.NoError(t, err, fmt.Sprintf("unexpected error on 202 response: %v", err))
		require.NotNil(t, out, "expected non-nil output on 202 response")
		assert.EqualValues(t, "42", out.ID)
		assert.EqualValues(t, "updating", out.Msg)
	})

	t.Run("NonOK_Status_ReturnsError", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("bad request"))
		}))
		defer srv.Close()

		c := NewClient(WithBaseURL(srv.URL))
		pu := NewPartialUpdater[map[string]string, pupdateOut](c, "svc", "/v1/items/{x}", map[string]string{"x": "X"})

		in := map[string]string{"name": "Bob"}
		out, err := pu.UpdatePartial(context.Background(), "123", &in)
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
		pu := NewPartialUpdater[pupdateIn, pupdateOut](c, "svc", "/v1/items/{x}", map[string]string{"x": "X"})

		in := pupdateIn{Name: "Mallory"}
		out, err := pu.UpdatePartial(context.Background(), "123", &in)
		require.Error(t, err, "expected JSON decode error")
		assert.Nil(t, out, fmt.Sprintf("expected nil output on error, got: %#v", out))
	})
}
