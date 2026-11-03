package api

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.eoscloud.io/go-eoscloud/pkg/auth"
	"golang.org/x/oauth2"
)

type fakeTS struct {
	mu       sync.Mutex
	calls    int
	claims   *auth.TokenClaims
	cleared  int32
	tokenErr error
}

func (f *fakeTS) Token() (*oauth2.Token, error) {
	if f.tokenErr != nil {
		return nil, f.tokenErr
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls++
	f.claims = &auth.TokenClaims{RealmID: "realm"}
	return &oauth2.Token{AccessToken: "tok", Expiry: time.Now().Add(time.Hour)}, nil
}

func (f *fakeTS) TokenClaims() (*auth.TokenClaims, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.claims, nil
}

func (f *fakeTS) ClearTokenStore() error {
	atomic.AddInt32(&f.cleared, 1)
	return nil
}

// withFakeAuth mimics WithDeviceFlow with a fake token source.
func withFakeAuth(ts *fakeTS, gotCtx *context.Context) Option {
	return func(c *client) {
		c.authCtx = context.Background()
		c.authFactory = func(ctx context.Context) auth.TokenSourceWithClaims {
			if gotCtx != nil {
				*gotCtx = ctx
			}
			return ts
		}
	}
}

func TestOptions_Modifiers(t *testing.T) {
	rt := &fakeRoundTripper{}
	custom := &http.Client{Transport: rt, Timeout: 7 * time.Second}
	c := NewClient(
		WithBaseURL("https://x"),
		WithHTTPClient(custom),
	).(*client)

	assert.Equal(t, 7*time.Second, c.httpClient.Timeout)
	assert.NotNil(t, c.httpClient.CheckRedirect)
}

func TestOptions_AuthOrderIrrelevant(t *testing.T) {
	for name, order := range map[string]func(hc *http.Client, ts *fakeTS) []Option{
		"httpClientFirst": func(hc *http.Client, ts *fakeTS) []Option {
			return []Option{WithHTTPClient(hc), withFakeAuth(ts, nil)}
		},
		"authFirst": func(hc *http.Client, ts *fakeTS) []Option {
			return []Option{withFakeAuth(ts, nil), WithHTTPClient(hc)}
		},
	} {
		t.Run(name, func(t *testing.T) {
			rt := &fakeRoundTripper{}
			hc := &http.Client{Transport: rt, Timeout: 7 * time.Second}
			c := NewClient(order(hc, &fakeTS{})...).(*client)

			assert.Equal(t, 7*time.Second, c.httpClient.Timeout)
			req, _ := http.NewRequest(http.MethodGet, "https://api.example.com/", nil)
			resp, err := c.httpClient.Do(req)
			require.NoError(t, err)
			_ = resp.Body.Close()
			assert.Equal(t, "Bearer tok", rt.lastReq.Header.Get("Authorization"))
		})
	}
}

func TestOptions_DefaultClientKeepsTimeoutWithAuth(t *testing.T) {
	var gotCtx context.Context
	c := NewClient(withFakeAuth(&fakeTS{}, &gotCtx)).(*client)
	assert.Equal(t, 60*time.Second, c.httpClient.Timeout)
	base, ok := gotCtx.Value(oauth2.HTTPClient).(*http.Client)
	require.True(t, ok, "token endpoint client must be passed via context")
	assert.Equal(t, 60*time.Second, base.Timeout)
}

func TestOptions_DebugOrderIrrelevant(t *testing.T) {
	rt := &fakeRoundTripper{}
	hc := &http.Client{Transport: rt}
	c := NewClient(WithDebugTokenPrinting(), withFakeAuth(&fakeTS{}, nil), WithHTTPClient(hc)).(*client)
	tr, ok := c.httpClient.Transport.(*oauth2.Transport)
	require.True(t, ok)
	_, ok = tr.Base.(*debugRoundTripper)
	assert.True(t, ok)
}

func TestRedirect_NoAuthorizationLeak(t *testing.T) {
	var leaked atomic.Bool
	other := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "" {
			leaked.Store(true)
		}
	}))
	defer other.Close()
	origin := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, other.URL, http.StatusFound)
	}))
	defer origin.Close()

	hc := origin.Client()
	c := NewClient(withFakeAuth(&fakeTS{}, nil), WithHTTPClient(hc)).(*client)
	req, _ := http.NewRequest(http.MethodGet, origin.URL, nil)
	resp, err := c.httpClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.False(t, leaked.Load())
}

func TestRedirect_Policy(t *testing.T) {
	mk := func(u string) *http.Request {
		r, _ := http.NewRequest(http.MethodGet, u, nil)
		return r
	}
	assert.NoError(t, safeCheckRedirect(mk("https://a.io/y"), []*http.Request{mk("https://a.io/x")}))
	assert.ErrorIs(t, safeCheckRedirect(mk("https://b.io/y"), []*http.Request{mk("https://a.io/x")}), http.ErrUseLastResponse)
	assert.ErrorIs(t, safeCheckRedirect(mk("http://a.io/y"), []*http.Request{mk("https://a.io/x")}), http.ErrUseLastResponse)
	assert.ErrorIs(t, safeCheckRedirect(mk("http://a.io/y"), []*http.Request{mk("http://a.io/x")}), http.ErrUseLastResponse)
	assert.NoError(t, safeCheckRedirect(mk("http://127.0.0.1:1/y"), []*http.Request{mk("http://127.0.0.1:1/x")}))
}

func TestRedirect_ChainsBaseCheckRedirect(t *testing.T) {
	mk := func(u string) *http.Request {
		r, _ := http.NewRequest(http.MethodGet, u, nil)
		return r
	}
	errCustom := errors.New("custom")
	called := false
	check := chainCheckRedirect(func(*http.Request, []*http.Request) error {
		called = true
		return errCustom
	})

	// Unsafe redirect is stopped before the custom policy runs.
	assert.ErrorIs(t, check(mk("https://b.io/y"), []*http.Request{mk("https://a.io/x")}), http.ErrUseLastResponse)
	assert.False(t, called)

	// Safe redirect is passed on to the custom policy.
	assert.ErrorIs(t, check(mk("https://a.io/y"), []*http.Request{mk("https://a.io/x")}), errCustom)
	assert.True(t, called)

	assert.NoError(t, chainCheckRedirect(nil)(mk("https://a.io/y"), []*http.Request{mk("https://a.io/x")}))
}

func TestWithBaseURL_Validation(t *testing.T) {
	for _, bad := range []string{"http://example.com", "ftp://x", "://", "https://", "example.com"} {
		c := NewClient(WithBaseURL(bad))
		_, err := c.Call(context.Background(), &CallRequest{Service: "s", Path: "/"})
		assert.Error(t, err, bad)
	}
	for _, good := range []string{"https://{service}.eoscloud.io", "http://localhost:8080", "http://127.0.0.5", "http://[::1]:80", "https://x"} {
		assert.NoError(t, validateBaseURL(good), good)
		assert.NoError(t, NewClient(WithBaseURL(good)).(*client).optErr, good)
	}
}

func TestClearKeyringOnAuthError(t *testing.T) {
	ts := &fakeTS{}
	c := &client{tokenSource: ts}

	c.clearKeyringOnAuthError(&oauth2.RetrieveError{Response: &http.Response{StatusCode: 503}, ErrorCode: ""}, 0)
	c.clearKeyringOnAuthError(&oauth2.RetrieveError{Response: &http.Response{StatusCode: 429}}, 0)
	c.clearKeyringOnAuthError(errors.New("oauth2: something"), 0)
	assert.Zero(t, atomic.LoadInt32(&ts.cleared))

	wrapped := errors.Join(errors.New("x"), &oauth2.RetrieveError{Response: &http.Response{StatusCode: 400}, ErrorCode: "invalid_grant"})
	c.clearKeyringOnAuthError(wrapped, 0)
	assert.EqualValues(t, 1, atomic.LoadInt32(&ts.cleared))
}

func TestSharedTokenSource_ConcurrentClaims(t *testing.T) {
	ts := &fakeTS{}
	// Zero expiry forces refetch so Token() keeps writing claims.
	shared := &sharedTokenSource{inner: ts, reuse: oauth2.ReuseTokenSource(nil, ts)}
	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := shared.Token()
			assert.NoError(t, err)
			cl, err := shared.TokenClaims()
			assert.NoError(t, err)
			assert.Equal(t, "realm", cl.RealmID)
		}()
	}
	wg.Wait()
}
