package api

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewJSONDecoder(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var in map[string]any
		dec := NewJSONDecoder(&in)
		body := io.NopCloser(strings.NewReader("{\"a\":1}"))
		out, err := dec(body)
		require.NoError(t, err, fmt.Sprintf("unexpected error: %v", err))

		m := out.(*map[string]any)
		assert.EqualValues(t, 1, (*m)["a"].(float64), fmt.Sprintf("decoded mismatch: %#v", m))
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		var in map[string]any
		dec := NewJSONDecoder(&in)
		body := io.NopCloser(strings.NewReader("{"))
		out, err := dec(body)
		require.Error(t, err, "expected error")
		assert.NotNil(t, out, "expected non-nil result on error")
	})

	t.Run("NilResult_DoesNotCloseBody", func(t *testing.T) {
		dec := NoResultDecoder()
		closed := false
		body := nopCloser{
			Reader:  strings.NewReader("x"),
			closeFn: func() error { closed = true; return nil },
		}
		out, err := dec(body)
		require.NoError(t, err, fmt.Sprintf("expected no error, got : %v", err))
		require.Nil(t, out, fmt.Sprintf("expected no value, got : %v", out))
		assert.EqualValues(t, false, closed, "body should not be closed by decoder when result is nil")
	})
}

type nopCloser struct {
	io.Reader
	closeFn func() error
}

func (n nopCloser) Close() error {
	if n.closeFn == nil {
		return errors.New("no closeFn")
	}
	return n.closeFn()
}
