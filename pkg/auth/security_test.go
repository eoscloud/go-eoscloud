package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func skipOnWindows(t *testing.T) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("POSIX permission semantics")
	}
}

// ---- SDK-4: file store hardening ----

func TestFileStore_NoHomeReturnsError(t *testing.T) {
	skipOnWindows(t)
	t.Setenv("HOME", "")
	t.Setenv("XDG_CONFIG_HOME", "")

	store := newFileStore("svc", "")
	assert.ErrorIs(t, store.Set("k", "v"), ErrNoTokenDir)
	_, err := store.Get("k")
	assert.ErrorIs(t, err, ErrNoTokenDir)
	assert.ErrorIs(t, store.Delete("k"), ErrNoTokenDir)
	assert.ErrorIs(t, store.DeleteAll(), ErrNoTokenDir)
}

func TestFileStore_RefusesSymlinkedFile(t *testing.T) {
	skipOnWindows(t)
	dir := filepath.Join(t.TempDir(), "tokens")
	store := &fileStore{service: "svc", tokenDir: dir}
	require.NoError(t, store.Set("k", "v"))

	victim := filepath.Join(t.TempDir(), "victim")
	require.NoError(t, os.WriteFile(victim, []byte("{}"), 0o600))
	require.NoError(t, os.Remove(store.path()))
	require.NoError(t, os.Symlink(victim, store.path()))

	assert.Error(t, store.Set("k2", "v2"))
	_, err := store.Get("k")
	assert.Error(t, err)

	content, err := os.ReadFile(victim)
	require.NoError(t, err)
	assert.Equal(t, "{}", string(content), "symlink target must not be modified")
}

func TestFileStore_RefusesWrongFileMode(t *testing.T) {
	skipOnWindows(t)
	dir := filepath.Join(t.TempDir(), "tokens")
	store := &fileStore{service: "svc", tokenDir: dir}
	require.NoError(t, store.Set("k", "v"))
	require.NoError(t, os.Chmod(store.path(), 0o644))

	assert.ErrorContains(t, store.Set("k2", "v2"), "0600")
	_, err := store.Get("k")
	assert.ErrorContains(t, err, "0600")
}

func TestFileStore_RefusesUnsafeDir(t *testing.T) {
	skipOnWindows(t)
	dir := filepath.Join(t.TempDir(), "tokens")
	require.NoError(t, os.Mkdir(dir, 0o700))
	require.NoError(t, os.Chmod(dir, 0o770))
	store := &fileStore{service: "svc", tokenDir: dir}

	assert.ErrorContains(t, store.Set("k", "v"), "writable")

	file := filepath.Join(t.TempDir(), "afile")
	require.NoError(t, os.WriteFile(file, nil, 0o600))
	assert.ErrorContains(t, (&fileStore{service: "svc", tokenDir: file}).Set("k", "v"), "not a directory")
}

func TestFileStore_AtomicWriteLeavesNoTempFiles(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "tokens")
	store := &fileStore{service: "svc", tokenDir: dir}

	var wg sync.WaitGroup
	for i := range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			assert.NoError(t, store.Set(fmt.Sprintf("k%d", i), "v"))
		}()
	}
	wg.Wait()

	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	require.Len(t, entries, 1, "only the final token file may remain")
	assert.Equal(t, "svc.json", entries[0].Name())

	data, err := os.ReadFile(store.path())
	require.NoError(t, err)
	var m map[string]string
	require.NoError(t, json.Unmarshal(data, &m), "file must always be complete JSON")
	assert.Len(t, m, 20)
}

func TestFileStore_CorruptFileIsNotSilentlyTrusted(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "tokens")
	store := &fileStore{service: "svc", tokenDir: dir}
	require.NoError(t, store.Set("k", "v"))
	require.NoError(t, os.WriteFile(store.path(), []byte("{broken"), 0o600))

	_, err := store.load()
	assert.ErrorIs(t, err, errCorruptTokenFile)

	_, err = store.Get("k")
	assert.ErrorIs(t, err, ErrNotFound, "callers see a clean miss")

	require.NoError(t, store.Set("k", "v2"), "a corrupt file is replaced on write")
	got, err := store.Get("k")
	require.NoError(t, err)
	assert.Equal(t, "v2", got)
}

// ---- SDK-7 / SDK-8: URL validation ----

func TestValidateAuthBaseURL(t *testing.T) {
	for _, ok := range []string{
		"https://auth.example.com", "http://localhost:8080", "http://127.0.0.1:1", "http://127.9.9.9", "http://[::1]:9000",
	} {
		assert.NoError(t, validateAuthBaseURL(ok), ok)
	}
	for _, bad := range []string{
		"", "auth.example.com", "http://auth.example.com", "ftp://auth.example.com", "https://", "javascript:alert(1)",
		"https://u:p@auth.example.com", "http://localhost.evil.com", "https://%zz",
	} {
		assert.Error(t, validateAuthBaseURL(bad), bad)
	}
}

func TestValidateVerificationURL(t *testing.T) {
	base := "https://auth.example.com/oauth2/device/auth"
	assert.NoError(t, validateVerificationURL("https://auth.example.com/verify?code=1", base))
	assert.NoError(t, validateVerificationURL("https://AUTH.example.com/verify", base))
	assert.NoError(t, validateVerificationURL("http://127.0.0.1:8080/verify", "http://127.0.0.1:8080/oauth2"))

	for _, bad := range []string{
		"http://auth.example.com/verify",
		"https://evil.example.com/verify",
		"https://auth.example.com.evil.com/verify",
		"file:///etc/passwd",
		"javascript:alert(1)",
		"calc.exe",
		"https://auth.example.com@evil.com/",
		"",
	} {
		assert.Error(t, validateVerificationURL(bad, base), bad)
	}
}

func TestSanitizeForTerminal(t *testing.T) {
	assert.Equal(t, "a[31mb", sanitizeForTerminal("a\x1b[31m\x07b"))
	assert.NotContains(t, sanitizeForTerminal("x\x1b]0;pwn\x07y"), "\x1b")
}

func TestAuthBaseURLErrorSurfacedFromToken(t *testing.T) {
	d := NewDeviceFlowConfig()
	d.With(WithDeviceFlowAuthBaseURL("http://evil.example.com"))
	ts := NewDeviceKeyringTokenSource(context.Background(), d, newMemoryStore())
	_, err := ts.Token()
	assert.ErrorContains(t, err, "invalid auth base URL")
	_, err = ts.TokenClaims()
	assert.Error(t, err)

	c := NewClientCredentialsConfig("id", "secret")
	c.With(WithClientCredentialsAuthBaseURL("::not a url"))
	cts := NewClientCredentialsKeyringTokenSource(context.Background(), c, nil)
	_, err = cts.Token()
	assert.ErrorContains(t, err, "invalid auth base URL")
}

// ---- SDK-10 / SDK-7: prompt writer and browser opener ----

type deviceServer struct {
	*httptest.Server
	deviceAuthCalls atomic.Int32
	tokenCalls      atomic.Int32
}

// newDeviceServer serves device auth and token endpoints. verificationURI may
// contain %s for the server URL. tokenHandler customizes /oauth2/token.
func newDeviceServer(t *testing.T, verificationURI string, tokenHandler func(w http.ResponseWriter, r *http.Request, n int32)) *deviceServer {
	t.Helper()
	ds := &deviceServer{}
	ds.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/oauth2/device/auth":
			ds.deviceAuthCalls.Add(1)
			uri := strings.ReplaceAll(verificationURI, "%s", ds.URL)
			_, _ = fmt.Fprintf(w, `{"device_code":"dc","user_code":"UC-1","verification_uri":%q,"interval":1,"expires_in":300}`, uri)
		case "/oauth2/token":
			tokenHandler(w, r, ds.tokenCalls.Add(1))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(ds.Close)
	return ds
}

func okToken(w http.ResponseWriter, jwt string) {
	_, _ = fmt.Fprintf(w, `{"access_token":%q,"token_type":"Bearer","expires_in":3600,"refresh_token":"r"}`, jwt)
}

func TestDeviceFlow_PromptWriterAndBrowserOpener(t *testing.T) {
	jwt := makeTestJWT(map[string]any{"eos": map[string]any{"realm": "r", "org": "o"}})
	srv := newDeviceServer(t, "%s/verify", func(w http.ResponseWriter, _ *http.Request, _ int32) { okToken(w, jwt) })

	var out bytes.Buffer
	var opened []string
	c := NewDeviceFlowConfig()
	c.With(
		WithDeviceFlowAuthBaseURL(srv.URL),
		WithDeviceFlowPromptWriter(&out),
		WithDeviceFlowBrowserOpener(func(u string) error { opened = append(opened, u); return nil }),
	)

	_, err := NewDeviceKeyringTokenSource(context.Background(), c, newMemoryStore()).Token()
	require.NoError(t, err)
	assert.Equal(t, []string{srv.URL + "/verify"}, opened)
	assert.Contains(t, out.String(), "UC-1")
	assert.Contains(t, out.String(), srv.URL+"/verify")
}

func TestDeviceFlow_DoesNotWriteToStdout(t *testing.T) {
	jwt := makeTestJWT(map[string]any{"eos": map[string]any{"realm": "r", "org": "o"}})
	srv := newDeviceServer(t, "%s/verify", func(w http.ResponseWriter, _ *http.Request, _ int32) { okToken(w, jwt) })

	r, w, err := os.Pipe()
	require.NoError(t, err)
	orig := os.Stdout
	os.Stdout = w
	defer func() { os.Stdout = orig }()

	c := NewDeviceFlowConfig()
	c.With(WithDeviceFlowAuthBaseURL(srv.URL), WithDeviceFlowPromptWriter(&bytes.Buffer{}))
	_, tokErr := NewDeviceKeyringTokenSource(context.Background(), c, newMemoryStore()).Token()
	require.NoError(t, tokErr)

	_ = w.Close()
	os.Stdout = orig
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	assert.Empty(t, buf.String())
}

func TestDeviceFlow_OpenerNilDisablesBrowser(t *testing.T) {
	jwt := makeTestJWT(map[string]any{"eos": map[string]any{"realm": "r", "org": "o"}})
	srv := newDeviceServer(t, "%s/verify", func(w http.ResponseWriter, _ *http.Request, _ int32) { okToken(w, jwt) })

	called := false
	orig := openBrowserFunc
	openBrowserFunc = func(string) error { called = true; return nil }
	defer func() { openBrowserFunc = orig }()

	var out bytes.Buffer
	c := NewDeviceFlowConfig()
	c.With(WithDeviceFlowAuthBaseURL(srv.URL), WithDeviceFlowPromptWriter(&out), WithDeviceFlowBrowserOpener(nil))
	_, err := NewDeviceKeyringTokenSource(context.Background(), c, newMemoryStore()).Token()
	require.NoError(t, err)
	assert.False(t, called)
	assert.Contains(t, out.String(), "UC-1")
}

func TestDeviceFlow_DefaultOpenerRejectsForeignHost(t *testing.T) {
	jwt := makeTestJWT(map[string]any{"eos": map[string]any{"realm": "r", "org": "o"}})
	srv := newDeviceServer(t, "https://evil.example.net/verify", func(w http.ResponseWriter, _ *http.Request, _ int32) { okToken(w, jwt) })

	called := false
	orig := openBrowserFunc
	openBrowserFunc = func(string) error { called = true; return nil }
	defer func() { openBrowserFunc = orig }()

	var out bytes.Buffer
	c := NewDeviceFlowConfig()
	c.With(WithDeviceFlowAuthBaseURL(srv.URL), WithDeviceFlowPromptWriter(&out))
	_, err := NewDeviceKeyringTokenSource(context.Background(), c, newMemoryStore()).Token()
	require.NoError(t, err)
	assert.False(t, called, "browser must not be opened for a foreign host")
	assert.Contains(t, out.String(), "Not opening a browser")
}

// ---- SDK-6: refresh error handling ----

func expiredStore(t *testing.T, c *DeviceFlowConfig, refresh string) *memoryStore {
	t.Helper()
	store := newMemoryStore()
	data, _ := json.Marshal(&oauth2.Token{AccessToken: "old", RefreshToken: refresh, TokenType: "Bearer", Expiry: time.Now().Add(-time.Hour)})
	_ = store.Set(namespacedTokenKey(DefaultDeviceFlowTokenKey, c.authBaseURL, c.clientID), string(data))
	return store
}

func TestDeviceFlow_TransientRefreshErrorsDoNotReauthenticate(t *testing.T) {
	for _, status := range []int{http.StatusInternalServerError, http.StatusBadGateway, http.StatusTooManyRequests} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			srv := newDeviceServer(t, "%s/verify", func(w http.ResponseWriter, _ *http.Request, _ int32) {
				w.WriteHeader(status)
			})
			c := NewDeviceFlowConfig()
			c.With(WithDeviceFlowAuthBaseURL(srv.URL), WithDeviceFlowBrowserOpener(nil), WithDeviceFlowPromptWriter(&bytes.Buffer{}))

			_, err := NewDeviceKeyringTokenSource(context.Background(), c, expiredStore(t, c, "rt")).Token()
			require.Error(t, err)
			assert.Zero(t, srv.deviceAuthCalls.Load(), "must not start a new device flow")
			var rerr *oauth2.RetrieveError
			assert.ErrorAs(t, err, &rerr)
		})
	}
}

func TestDeviceFlow_OtherOAuthErrorsDoNotReauthenticate(t *testing.T) {
	srv := newDeviceServer(t, "%s/verify", func(w http.ResponseWriter, _ *http.Request, _ int32) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"server_error"}`))
	})
	c := NewDeviceFlowConfig()
	c.With(WithDeviceFlowAuthBaseURL(srv.URL), WithDeviceFlowBrowserOpener(nil), WithDeviceFlowPromptWriter(&bytes.Buffer{}))

	_, err := NewDeviceKeyringTokenSource(context.Background(), c, expiredStore(t, c, "rt")).Token()
	require.Error(t, err)
	assert.Zero(t, srv.deviceAuthCalls.Load())
}

func TestDeviceFlow_MissingRefreshTokenReauthenticates(t *testing.T) {
	jwt := makeTestJWT(map[string]any{"eos": map[string]any{"realm": "r", "org": "o"}})
	srv := newDeviceServer(t, "%s/verify", func(w http.ResponseWriter, _ *http.Request, _ int32) { okToken(w, jwt) })
	c := NewDeviceFlowConfig()
	c.With(WithDeviceFlowAuthBaseURL(srv.URL), WithDeviceFlowBrowserOpener(nil), WithDeviceFlowPromptWriter(&bytes.Buffer{}))

	_, err := NewDeviceKeyringTokenSource(context.Background(), c, expiredStore(t, c, "")).Token()
	require.NoError(t, err)
	assert.Equal(t, int32(1), srv.deviceAuthCalls.Load())
}

func TestDeviceFlow_ClearTokenStoreOnlyDeletesOwnKey(t *testing.T) {
	c := NewDeviceFlowConfig()
	c.With(WithDeviceFlowAuthBaseURL("https://auth.example.com"))
	store := newMemoryStore()
	own := namespacedTokenKey(DefaultDeviceFlowTokenKey, c.authBaseURL, c.clientID)
	_ = store.Set(own, "x")
	_ = store.Set("other-flow", "keep")
	_ = store.Set(namespacedTokenKey(DefaultDeviceFlowTokenKey, "https://other.example.com", c.clientID), "keep2")

	require.NoError(t, NewDeviceKeyringTokenSource(context.Background(), c, store).ClearTokenStore())

	_, err := store.Get(own)
	assert.ErrorIs(t, err, ErrNotFound)
	assert.Len(t, store.data, 2)

	// Clearing again (entry absent) is not an error.
	require.NoError(t, NewDeviceKeyringTokenSource(context.Background(), c, store).ClearTokenStore())
}

// ---- SDK-9: concurrency and ctx decoupling ----

func TestDeviceFlow_ParallelCallersStartOneFlow(t *testing.T) {
	jwt := makeTestJWT(map[string]any{"eos": map[string]any{"realm": "r", "org": "o"}})
	srv := newDeviceServer(t, "%s/verify", func(w http.ResponseWriter, _ *http.Request, _ int32) { okToken(w, jwt) })
	c := NewDeviceFlowConfig()
	c.With(WithDeviceFlowAuthBaseURL(srv.URL), WithDeviceFlowBrowserOpener(nil), WithDeviceFlowPromptWriter(&bytes.Buffer{}))
	ts := NewDeviceKeyringTokenSource(context.Background(), c, newMemoryStore())

	var wg sync.WaitGroup
	for range 16 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			tok, err := ts.Token()
			assert.NoError(t, err)
			assert.NotNil(t, tok)
			cl, err := ts.TokenClaims()
			assert.NoError(t, err)
			if cl != nil {
				assert.Equal(t, "r", cl.RealmID)
			}
		}()
	}
	wg.Wait()
	assert.Equal(t, int32(1), srv.deviceAuthCalls.Load(), "exactly one device flow")
}

func TestDeviceFlow_ParallelRefreshHitsServerOnce(t *testing.T) {
	jwt := makeTestJWT(map[string]any{"eos": map[string]any{"realm": "r", "org": "o"}})
	srv := newDeviceServer(t, "%s/verify", func(w http.ResponseWriter, _ *http.Request, _ int32) { okToken(w, jwt) })
	c := NewDeviceFlowConfig()
	c.With(WithDeviceFlowAuthBaseURL(srv.URL))
	ts := NewDeviceKeyringTokenSource(context.Background(), c, expiredStore(t, c, "rt"))

	var wg sync.WaitGroup
	for range 16 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := ts.Token()
			assert.NoError(t, err)
		}()
	}
	wg.Wait()
	assert.Equal(t, int32(1), srv.tokenCalls.Load())
}

func TestDeviceFlow_CancelledConstructorContextDoesNotBreakRefresh(t *testing.T) {
	jwt := makeTestJWT(map[string]any{"eos": map[string]any{"realm": "r", "org": "o"}})
	srv := newDeviceServer(t, "%s/verify", func(w http.ResponseWriter, _ *http.Request, _ int32) { okToken(w, jwt) })
	c := NewDeviceFlowConfig()
	c.With(WithDeviceFlowAuthBaseURL(srv.URL))

	ctx, cancel := context.WithCancel(context.Background())
	ts := NewDeviceKeyringTokenSource(ctx, c, expiredStore(t, c, "rt"))
	cancel()

	_, err := ts.Token()
	assert.NoError(t, err)
}

func TestDeviceFlow_ConstructorContextValuesPreserved(t *testing.T) {
	jwt := makeTestJWT(map[string]any{"eos": map[string]any{"realm": "r", "org": "o"}})
	srv := newDeviceServer(t, "%s/verify", func(w http.ResponseWriter, _ *http.Request, _ int32) { okToken(w, jwt) })
	c := NewDeviceFlowConfig()
	c.With(WithDeviceFlowAuthBaseURL(srv.URL))

	var used atomic.Int32
	hc := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		used.Add(1)
		return http.DefaultTransport.RoundTrip(r)
	})}
	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, hc)
	_, err := NewDeviceKeyringTokenSource(ctx, c, expiredStore(t, c, "rt")).Token()
	require.NoError(t, err)
	assert.Positive(t, used.Load(), "oauth2.HTTPClient from ctx must be honored")
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestClientCredentials_ParallelFetchesOnce(t *testing.T) {
	jwt := makeTestJWT(map[string]any{"eos": map[string]any{"realm": "r", "org": "o"}})
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintf(w, `{"access_token":%q,"token_type":"Bearer","expires_in":3600}`, jwt)
	}))
	defer srv.Close()

	c := NewClientCredentialsConfig("id", "secret")
	c.With(WithClientCredentialsAuthBaseURL(srv.URL))
	ts := NewClientCredentialsKeyringTokenSource(context.Background(), c, nil)

	var wg sync.WaitGroup
	for range 16 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := ts.Token()
			assert.NoError(t, err)
			cl, err := ts.TokenClaims()
			assert.NoError(t, err)
			assert.NotNil(t, cl)
		}()
	}
	wg.Wait()
	assert.Equal(t, int32(1), calls.Load())
}

func TestClientCredentials_NeverWritesStore(t *testing.T) {
	jwt := makeTestJWT(map[string]any{"eos": map[string]any{"realm": "r", "org": "o"}})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintf(w, `{"access_token":%q,"token_type":"Bearer","expires_in":3600}`, jwt)
	}))
	defer srv.Close()

	c := NewClientCredentialsConfig("id", "secret")
	c.With(WithClientCredentialsAuthBaseURL(srv.URL))
	store := newMemoryStore()
	_, err := NewClientCredentialsKeyringTokenSource(context.Background(), c, store).Token()
	require.NoError(t, err)
	assert.Empty(t, store.data)
}
