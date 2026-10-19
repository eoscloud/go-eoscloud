package api

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithUserAgent_Validation(t *testing.T) {
	tests := []struct {
		name, uName, uVersion string
		wantErr               bool
	}{
		{name: "valid", uName: "my-app", uVersion: "1.2.3"},
		{name: "slash in version is fine", uName: "app", uVersion: "1/2"},
		{name: "empty name", uVersion: "1.2.3", wantErr: true},
		{name: "empty version", uName: "my-app", wantErr: true},
		{name: "both empty", wantErr: true},
		{name: "space in name", uName: "my app", uVersion: "1", wantErr: true},
		{name: "slash in name", uName: "my/app", uVersion: "1", wantErr: true},
		{name: "space in version", uName: "app", uVersion: "1 2", wantErr: true},
		{name: "newline", uName: "app", uVersion: "1\r\nX: y", wantErr: true},
		{name: "non ascii", uName: "äpp", uVersion: "1", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opt, err := WithUserAgent(tt.uName, tt.uVersion)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, opt)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, opt)
		})
	}
}

func TestClient_Call_UserAgentHeader(t *testing.T) {
	platform := fmt.Sprintf("(%s; %s/%s)", runtime.Version(), runtime.GOOS, runtime.GOARCH)

	call := func(t *testing.T, extra ...Option) *http.Request {
		rt := &fakeRoundTripper{}
		opts := append([]Option{
			WithBaseURL("https://api.example.com/"),
			WithHTTPClient(&http.Client{Transport: rt}),
		}, extra...)
		_, err := NewClient(opts...).Call(context.Background(), &CallRequest{
			Method:      GET,
			Path:        "/v1/x",
			Encoder:     func() (string, io.Reader, error) { return "", nil, nil },
			StatusCheck: func(int) bool { return true },
			Decoder:     func(io.ReadCloser) (any, error) { return nil, nil },
		})
		require.NoError(t, err)
		return rt.lastReq
	}

	t.Run("without user part", func(t *testing.T) {
		assert.Equal(t, "go-eoscloud/"+Version+" "+platform, call(t).Header.Get("User-Agent"))
	})

	t.Run("with user part", func(t *testing.T) {
		opt, err := WithUserAgent("my-app", "1.2.3")
		require.NoError(t, err)
		assert.Equal(t,
			"go-eoscloud/"+Version+" my-app/1.2.3 "+platform,
			call(t, opt).Header.Get("User-Agent"),
		)
	})
}

func TestAppendUserAgent(t *testing.T) {
	assert.Equal(t, "ours", appendUserAgent("", "ours"))
	assert.Equal(t, "theirs ours", appendUserAgent("theirs", "ours"))
}
