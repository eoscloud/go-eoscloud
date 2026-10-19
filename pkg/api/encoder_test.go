package api

import (
	"encoding/json"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewJSONEncoder(t *testing.T) {
	t.Run("WithData", func(t *testing.T) {
		data := map[string]any{"a": 1}
		enc := NewJSONEncoder(&data)
		ct, r, err := enc()
		require.NoError(t, err, fmt.Sprintf("unexpected error: %v", err))
		require.EqualValues(t, "application/json", ct, fmt.Sprintf("content type mismatch: %q", ct))
		b, readErr := io.ReadAll(r)
		require.NoError(t, readErr, "unexpected read errors")

		var got map[string]any
		unmarshalErr := json.Unmarshal(b, &got)
		require.NoError(t, unmarshalErr)
		assert.EqualValues(t, 1, got["a"].(float64), fmt.Sprintf("payload mismatch: %#v", got))
	})

	t.Run("NilData", func(t *testing.T) {
		enc := NoBodyEncoder()
		ct, r, err := enc()
		require.NoError(t, err, fmt.Sprintf("unexpected error: %v", err))
		require.EqualValues(t, "application/json", ct, fmt.Sprintf("content type mismatch: %q", ct))
		assert.Nil(t, r, fmt.Sprintf("expected no body for nil data, got reader=%v", r))
	})
}
