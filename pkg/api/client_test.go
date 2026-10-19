package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.eoscloud.io/go-eoscloud/eoscloud/common/model"
	"go.eoscloud.io/go-eoscloud/pkg/auth"
	"golang.org/x/oauth2"
)

type fakeRoundTripper struct {
	lastReq *http.Request
	resp    *http.Response
	err     error
}

func (f *fakeRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	f.lastReq = req
	if f.err != nil {
		return nil, f.err
	}
	if f.resp == nil {
		f.resp = &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader("")),
		}
	}
	return f.resp, nil
}

func newClientWithFake(rt *fakeRoundTripper) Caller {
	hc := &http.Client{Transport: rt}
	return NewClient(WithBaseURL("https://api.example.com/"), WithHTTPClient(hc))
}

func TestClient_Call(t *testing.T) {
	t.Run("URLAndParamsAndQuery", func(t *testing.T) {
		rt := &fakeRoundTripper{
			resp: &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader("{}")),
			},
		}
		cli := newClientWithFake(rt)

		enc := func() (string, io.Reader, error) {
			return "application/json", strings.NewReader("{}"), nil
		}
		dec := func(body io.ReadCloser) (any, error) {
			_, _ = io.ReadAll(body)
			return map[string]any{"ok": true}, nil
		}
		query := url.Values{"a": {"1"}, "b": {"x y"}}
		res, err := cli.Call(context.Background(), &CallRequest{
			Method:      GET,
			Path:        "/v1/users/{id}",
			PathParams:  map[string]string{"id": "u-1"},
			Query:       query,
			Encoder:     enc,
			StatusCheck: func(code int) bool { return code == 200 },
			Decoder:     dec,
		})
		assert.NoError(t, err, "Call should not return an error")

		expectedRes := map[string]any{}
		assert.IsType(t, expectedRes, res, "unexpected result type")
		require.NotNil(t, rt)
		assert.NotNil(t, rt.lastReq, "no request captured")

		wantURL := "https://api.example.com/v1/users/u-1?a=1&b=x+y"
		acceptHeader := rt.lastReq.Header.Get("Accept")
		contentTypeHeader := rt.lastReq.Header.Get("Content-Type")

		assert.Equal(
			t,
			wantURL,
			rt.lastReq.URL.String(),
			fmt.Sprintf("URL mismatch: acceptHeader %q want %q", rt.lastReq.URL.String(), wantURL),
		)
		assert.Equal(
			t,
			http.MethodGet,
			rt.lastReq.Method,
			fmt.Sprintf("method mismatch: acceptHeader %q", rt.lastReq.Method),
		)
		assert.Equal(
			t,
			"application/json",
			acceptHeader,
			fmt.Sprintf("Accept header mismatch: %q", acceptHeader),
		)
		assert.Equal(
			t,
			"application/json",
			contentTypeHeader,
			fmt.Sprintf("Content-Type header mismatch: %q", contentTypeHeader),
		)
	})

	t.Run("MissingPathParams", func(t *testing.T) {
		p := "/v1/{missing}/x"
		cli := NewClient(WithBaseURL("https://api.example.com/"))
		_, err := cli.Call(context.Background(), &CallRequest{
			Method:      GET,
			Path:        p,
			PathParams:  map[string]string{"other": "1"},
			Encoder:     func() (string, io.Reader, error) { return "", nil, nil },
			StatusCheck: func(int) bool { return true },
		})
		assert.Error(t, err, "expected missing basePathPattern parameters error")
		assert.ErrorContains(t, err, fmt.Sprintf("missing basePathPattern parameters for %q", p))
	})

	t.Run("StatusNotOKIncludesBody", func(t *testing.T) {
		apiError := model.APIError{
			Type:      "sum-thing-wong",
			RequestID: "request-id",
			Errors:    nil,
		}
		buf := bytes.Buffer{}

		err := json.NewEncoder(&buf).Encode(apiError)
		assert.NoError(t, err, "unexpected error encoding APIError")
		rt := &fakeRoundTripper{
			resp: &http.Response{
				StatusCode: 400, Body: io.NopCloser(bytes.NewReader(buf.Bytes())),
			},
		}
		cli := newClientWithFake(rt)
		_, err = cli.Call(context.Background(), &CallRequest{
			Method:      DELETE,
			Path:        "/v1/thing",
			Encoder:     func() (string, io.Reader, error) { return "", nil, nil },
			StatusCheck: func(int) bool { return false },
		})
		assert.Error(t, err, "expected an error")

		assert.ErrorContains(t, err, "type='sum-thing-wong', request_id='request-id', errors=[]", "error does not contain status: %v", err)
	})

	t.Run("EncoderError", func(t *testing.T) {
		rt := &fakeRoundTripper{}
		cli := newClientWithFake(rt)
		expected := errors.New("encode fail")
		_, err := cli.Call(context.Background(), &CallRequest{
			Method:      PUT,
			Path:        "/v1/x",
			Encoder:     func() (string, io.Reader, error) { return "", nil, expected },
			StatusCheck: func(int) bool { return true },
		})
		assert.ErrorContains(t, err, expected.Error(), fmt.Sprintf("expected encoder error, got %v", err))
	})

	t.Run("NoDecoderReturnsNil", func(t *testing.T) {
		rt := &fakeRoundTripper{
			resp: &http.Response{
				StatusCode: 200, Body: io.NopCloser(strings.NewReader("")),
			},
		}
		cli := newClientWithFake(rt)
		v, err := cli.Call(context.Background(), &CallRequest{
			Method:      GET,
			Path:        "/ok",
			Encoder:     func() (string, io.Reader, error) { return "", nil, nil },
			StatusCheck: func(int) bool { return true },
		})
		assert.Nil(t, err, "expected error is nil")
		assert.Nil(t, v, "expected value is nil")
	})

	t.Run("NoContentType_NoAcceptHeader", func(t *testing.T) {
		rt := &fakeRoundTripper{
			resp: &http.Response{
				StatusCode: 200, Body: io.NopCloser(strings.NewReader("")),
			},
		}
		cli := newClientWithFake(rt)
		_, err := cli.Call(context.Background(), &CallRequest{
			Method:      GET,
			Path:        "/ok",
			Encoder:     func() (string, io.Reader, error) { return "", nil, nil },
			StatusCheck: func(int) bool { return true },
		})
		assert.NoError(t, err, "unexpected error")
		acceptHeader := rt.lastReq.Header.Get("Accept")
		assert.Empty(t, acceptHeader, fmt.Sprintf("Accept should be empty when no content type, got %q", acceptHeader))
	})
}

// fakeTokenSource is a minimal auth.TokenSourceWithClaims for testing client behaviour.
type fakeTokenSource struct {
	clearCalled bool
}

func (f *fakeTokenSource) Token() (*oauth2.Token, error) {
	return &oauth2.Token{AccessToken: "fake-token"}, nil
}

func (f *fakeTokenSource) TokenClaims() (*auth.TokenClaims, error) {
	return &auth.TokenClaims{RealmID: "test-realm"}, nil
}

func (f *fakeTokenSource) ClearTokenStore() error {
	f.clearCalled = true
	return nil
}

func TestClient_ClearKeyringOnAuthError(t *testing.T) {
	t.Run("HTTP401_DoesNotClearTokenStore", func(t *testing.T) {
		// A bare 401 response must NOT wipe the token store so that a stored
		// refresh token survives and can be used on the next request.
		fts := &fakeTokenSource{}
		apiErr := model.APIError{Type: "unauthorized", RequestID: "req-1"}
		buf := bytes.Buffer{}
		require.NoError(t, json.NewEncoder(&buf).Encode(apiErr))

		rt := &fakeRoundTripper{
			resp: &http.Response{
				StatusCode: http.StatusUnauthorized,
				Body:       io.NopCloser(bytes.NewReader(buf.Bytes())),
			},
		}
		hc := &http.Client{Transport: rt}
		cli := NewClient(
			WithBaseURL("https://api.example.com/"),
			WithHTTPClient(hc),
			WithTokenSource(fts),
		)

		_, _ = cli.Call(context.Background(), &CallRequest{
			Method:      GET,
			Path:        "/v1/protected",
			Encoder:     func() (string, io.Reader, error) { return "", nil, nil },
			StatusCheck: func(int) bool { return false },
		})

		assert.False(t, fts.clearCalled, "ClearTokenStore must not be called on a bare HTTP 401")
	})

	t.Run("OAuth2TransportError_ClearsTokenStore", func(t *testing.T) {
		// An oauth2: error from the transport (e.g. failed token endpoint) should
		// still clear the store, as it indicates corrupted/unusable credentials.
		fts := &fakeTokenSource{}
		rt := &fakeRoundTripper{
			err: errors.New("oauth2: token expired and refresh token is not set"),
		}
		hc := &http.Client{Transport: rt}
		cli := NewClient(
			WithBaseURL("https://api.example.com/"),
			WithHTTPClient(hc),
			WithTokenSource(fts),
		)

		_, _ = cli.Call(context.Background(), &CallRequest{
			Method:      GET,
			Path:        "/v1/protected",
			Encoder:     func() (string, io.Reader, error) { return "", nil, nil },
			StatusCheck: func(int) bool { return false },
		})

		assert.True(t, fts.clearCalled, "ClearTokenStore must be called when an oauth2 transport error occurs")
	})
}

func TestClient_TokenClaims(t *testing.T) {
	t.Run("WithTokenSource_ReturnsClaims", func(t *testing.T) {
		cli := NewClient(WithTokenSource(&fakeTokenSource{}))

		provider, ok := cli.(ClaimsProvider)
		require.True(t, ok, "client returned by NewClient must implement ClaimsProvider")

		claims, err := provider.TokenClaims()
		require.NoError(t, err)
		require.NotNil(t, claims)
		assert.Equal(t, "test-realm", claims.RealmID)
	})

	t.Run("NoAuthFlow_ReturnsError", func(t *testing.T) {
		cli := NewClient(WithBaseURL("https://api.example.com/"))

		provider, ok := cli.(ClaimsProvider)
		require.True(t, ok, "client returned by NewClient must implement ClaimsProvider")

		claims, err := provider.TokenClaims()
		assert.Error(t, err)
		assert.Nil(t, claims)
	})
}
