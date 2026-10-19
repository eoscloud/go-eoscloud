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

type execIn struct {
	Name string `json:"name"`
}

type execOut struct {
	OK  bool   `json:"ok"`
	Msg string `json:"msg"`
}

func TestExecutor_Execute(t *testing.T) {
	t.Run("Success_WithID", func(t *testing.T) {
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
			_, _ = w.Write([]byte(`{"ok":true, "msg":"done"}`))
		}))
		defer srv.Close()

		c := NewClient(WithBaseURL(srv.URL))

		pathPattern := "/v1/items/{x}"
		params := map[string]string{"x": "X"}
		ex := NewExecutor[execIn, execOut](
			c, "svc",
			pathPattern,
			params,
			"do",
			POST,
			IsHTTPStatusOK,
		)
		id := "123"
		in := execIn{Name: "Bob"}

		out, err := ex.Execute(context.Background(), &id, &in)
		require.NoError(t, err, fmt.Sprintf("unexpected error: %v", err))
		require.NotNil(t, out, "expected non-nil output")

		assert.EqualValues(t, http.MethodPost, gotMethod, "method mismatch")
		assert.EqualValues(t, "/v1/items/X/123/:do", gotPath, fmt.Sprintf("path mismatch: %q", gotPath))
		assert.EqualValues(t, "Bob", gotBody["name"], fmt.Sprintf("payload mismatch: %#v", gotBody))
		assert.EqualValues(t, true, out.OK, fmt.Sprintf("decoded mismatch: %#v", out))
		assert.EqualValues(t, "done", out.Msg, fmt.Sprintf("decoded mismatch: %#v", out))
	})

	t.Run("Success_WithoutID", func(t *testing.T) {
		var gotPath string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"ok":true, "msg":"no-id"}`))
		}))
		defer srv.Close()

		c := NewClient(WithBaseURL(srv.URL))

		ex := NewExecutor[map[string]string, execOut](
			c,
			"svc",
			"/v1/items/{x}",
			map[string]string{"x": "X"},
			"do",
			PUT,
			IsHTTPStatusOK,
		)
		out, err := ex.Execute(context.Background(), nil, &map[string]string{"name": "Alice"})
		require.NoError(t, err, fmt.Sprintf("unexpected error: %v", err))
		require.NotNil(t, out)
		assert.EqualValues(t, "/v1/items/X/:do", gotPath, fmt.Sprintf("path mismatch: %q", gotPath))
		assert.EqualValues(t, "no-id", out.Msg)
	})

	t.Run("NonOK_Status_ReturnsError", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("bad request"))
		}))
		defer srv.Close()

		c := NewClient(WithBaseURL(srv.URL))
		ex := NewExecutor[execIn, execOut](
			c,
			"svc",
			"/v1/items/{x}",
			map[string]string{"x": "X"},
			"do",
			POST,
			IsHTTPStatusOK,
		)
		in := execIn{Name: "Bob"}
		out, err := ex.Execute(context.Background(), nil, &in)
		require.Error(t, err, "expected error on non-200 response")
		assert.Nil(t, out, fmt.Sprintf("expected nil output on error, got: %#v", out))
	})

	// Decoder error (invalid JSON)
	t.Run("InvalidJSON_ReturnsError", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("{")) // invalid JSON
		}))
		defer srv.Close()

		c := NewClient(WithBaseURL(srv.URL))
		ex := NewExecutor[map[string]any, execOut](
			c,
			"svc",
			"/v1/items/{x}",
			map[string]string{"x": "X"},
			"do",
			PATCH,
			IsHTTPStatusOK,
		)
		in := map[string]any{"a": 1}
		out, err := ex.Execute(context.Background(), nil, &in)
		require.Error(t, err, "expected JSON decode error")
		assert.Nil(t, out, fmt.Sprintf("expected nil output on error, got: %#v", out))
	})
}

func TestNoResultExecutor_Execute(t *testing.T) {
	t.Run("Success_WithID", func(t *testing.T) {
		var gotPath string
		var gotMethod string
		var gotBody map[string]any

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method
			b, _ := io.ReadAll(r.Body)
			_ = r.Body.Close()
			_ = json.Unmarshal(b, &gotBody)

			w.WriteHeader(http.StatusNoContent)
		}))
		defer srv.Close()

		c := NewClient(WithBaseURL(srv.URL))
		ex := NewNoResultExecutor[execIn](
			c, "svc",
			"/v1/items/{x}",
			map[string]string{"x": "X"},
			"archive",
			POST,
			IsHTTPStatusNoContent,
		)
		id := "456"
		in := execIn{Name: "Bob"}

		err := ex.Execute(context.Background(), &id, &in)
		require.NoError(t, err, fmt.Sprintf("unexpected error: %v", err))

		assert.EqualValues(t, http.MethodPost, gotMethod, "method mismatch")
		assert.EqualValues(t, "/v1/items/X/456/:archive", gotPath, fmt.Sprintf("path mismatch: %q", gotPath))
		assert.EqualValues(t, "Bob", gotBody["name"], fmt.Sprintf("payload mismatch: %#v", gotBody))
	})

	t.Run("Success_WithoutID", func(t *testing.T) {
		var gotPath string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			w.WriteHeader(http.StatusNoContent)
		}))
		defer srv.Close()

		c := NewClient(WithBaseURL(srv.URL))
		ex := NewNoResultExecutor[execIn](
			c, "svc",
			"/v1/items/{x}",
			map[string]string{"x": "X"},
			"archive",
			POST,
			IsHTTPStatusNoContent,
		)
		in := execIn{Name: "Alice"}

		err := ex.Execute(context.Background(), nil, &in)
		require.NoError(t, err, fmt.Sprintf("unexpected error: %v", err))
		assert.EqualValues(t, "/v1/items/X/:archive", gotPath, fmt.Sprintf("path mismatch: %q", gotPath))
	})

	t.Run("NonNoContent_Status_ReturnsError", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("bad request"))
		}))
		defer srv.Close()

		c := NewClient(WithBaseURL(srv.URL))
		ex := NewNoResultExecutor[execIn](
			c, "svc",
			"/v1/items/{x}",
			map[string]string{"x": "X"},
			"archive",
			POST,
			IsHTTPStatusNoContent,
		)
		in := execIn{Name: "Bob"}

		err := ex.Execute(context.Background(), nil, &in)
		require.Error(t, err, "expected error on non-204 response")
	})
}

func TestNoInputExecutor_Execute(t *testing.T) {
	t.Run("Success_WithID", func(t *testing.T) {
		var gotPath string
		var gotMethod string

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"ok":true,"msg":"restored"}`))
		}))
		defer srv.Close()

		c := NewClient(WithBaseURL(srv.URL))
		ex := NewNoInputExecutor[execOut](
			c, "svc",
			"/v1/items/{x}",
			map[string]string{"x": "X"},
			"restore",
			POST,
			IsHTTPStatusOK,
		)
		id := "789"

		out, err := ex.Execute(context.Background(), &id)
		require.NoError(t, err, fmt.Sprintf("unexpected error: %v", err))
		require.NotNil(t, out, "expected non-nil output")

		assert.EqualValues(t, http.MethodPost, gotMethod, "method mismatch")
		assert.EqualValues(t, "/v1/items/X/789/:restore", gotPath, fmt.Sprintf("path mismatch: %q", gotPath))
		assert.EqualValues(t, true, out.OK)
		assert.EqualValues(t, "restored", out.Msg)
	})

	t.Run("Success_WithoutID", func(t *testing.T) {
		var gotPath string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"ok":true,"msg":"no-id"}`))
		}))
		defer srv.Close()

		c := NewClient(WithBaseURL(srv.URL))
		ex := NewNoInputExecutor[execOut](
			c, "svc",
			"/v1/items/{x}",
			map[string]string{"x": "X"},
			"restore",
			POST,
			IsHTTPStatusOK,
		)

		out, err := ex.Execute(context.Background(), nil)
		require.NoError(t, err, fmt.Sprintf("unexpected error: %v", err))
		require.NotNil(t, out)
		assert.EqualValues(t, "/v1/items/X/:restore", gotPath, fmt.Sprintf("path mismatch: %q", gotPath))
		assert.EqualValues(t, "no-id", out.Msg)
	})

	t.Run("NonOK_Status_ReturnsError", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte("not found"))
		}))
		defer srv.Close()

		c := NewClient(WithBaseURL(srv.URL))
		ex := NewNoInputExecutor[execOut](
			c, "svc",
			"/v1/items/{x}",
			map[string]string{"x": "X"},
			"restore",
			POST,
			IsHTTPStatusOK,
		)

		out, err := ex.Execute(context.Background(), nil)
		require.Error(t, err, "expected error on non-200 response")
		assert.Nil(t, out, fmt.Sprintf("expected nil output on error, got: %#v", out))
	})
}

func TestNoIOExecutor_Execute(t *testing.T) {
	t.Run("Success_WithID", func(t *testing.T) {
		var gotPath string
		var gotMethod string

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method
			w.WriteHeader(http.StatusAccepted)
		}))
		defer srv.Close()

		c := NewClient(WithBaseURL(srv.URL))
		ex := NewNoIOExecutor(
			c, "svc",
			"/v1/items/{x}",
			map[string]string{"x": "X"},
			"recover-policy",
			POST,
			IsHTTPStatusAccepted,
		)
		id := "123"

		err := ex.Execute(context.Background(), &id)
		require.NoError(t, err, fmt.Sprintf("unexpected error: %v", err))

		assert.EqualValues(t, http.MethodPost, gotMethod, "method mismatch")
		assert.EqualValues(t, "/v1/items/X/123/:recover-policy", gotPath, fmt.Sprintf("path mismatch: %q", gotPath))
	})

	t.Run("Success_WithoutID", func(t *testing.T) {
		var gotPath string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			w.WriteHeader(http.StatusAccepted)
		}))
		defer srv.Close()

		c := NewClient(WithBaseURL(srv.URL))
		ex := NewNoIOExecutor(
			c, "svc",
			"/v1/items/{x}",
			map[string]string{"x": "X"},
			"recover-policy",
			POST,
			IsHTTPStatusAccepted,
		)

		err := ex.Execute(context.Background(), nil)
		require.NoError(t, err, fmt.Sprintf("unexpected error: %v", err))
		assert.EqualValues(t, "/v1/items/X/:recover-policy", gotPath, fmt.Sprintf("path mismatch: %q", gotPath))
	})

	t.Run("NonAccepted_Status_ReturnsError", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("bad request"))
		}))
		defer srv.Close()

		c := NewClient(WithBaseURL(srv.URL))
		ex := NewNoIOExecutor(
			c, "svc",
			"/v1/items/{x}",
			map[string]string{"x": "X"},
			"recover-policy",
			POST,
			IsHTTPStatusAccepted,
		)

		err := ex.Execute(context.Background(), nil)
		require.Error(t, err, "expected error on non-202 response")
	})
}
