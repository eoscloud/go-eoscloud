package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJoinIDPath(t *testing.T) {
	cases := map[string]string{
		"abc":  "/vms/abc/",
		"a?b":  "/vms/a%3Fb/",
		"a#b":  "/vms/a%23b/",
		"a/b":  "/vms/a%2Fb/",
		"../x": "/vms/..%2Fx/",
	}
	for id, want := range cases {
		got, err := joinIDPath("/vms/", id)
		assert.NoError(t, err, id)
		assert.Equal(t, want, got, id)
	}
	for _, id := range []string{"", ".", ".."} {
		_, err := joinIDPath("/vms", id)
		assert.Error(t, err, id)
	}
}

func TestGetter_RejectsBadID(t *testing.T) {
	g := NewGetter[any](nil, "svc", "/vms", nil)
	_, err := g.Get(t.Context(), "..", nil)
	assert.Error(t, err)

	id := ".."
	e := NewNoIOExecutor(nil, "svc", "/vms", nil, "x", POST, IsHTTPStatusOK)
	assert.Error(t, e.Execute(t.Context(), &id))
}
