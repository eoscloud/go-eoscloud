package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

// memoryStore is a simple in-memory TokenStore for testing.
type memoryStore struct {
	data map[string]string
}

func newMemoryStore() *memoryStore {
	return &memoryStore{data: make(map[string]string)}
}

func (m *memoryStore) Get(key string) (string, error) {
	if val, ok := m.data[key]; ok {
		return val, nil
	}
	return "", ErrNotFound
}

func (m *memoryStore) Set(key, value string) error {
	m.data[key] = value
	return nil
}

func (m *memoryStore) Delete(key string) error {
	delete(m.data, key)
	return nil
}

func (m *memoryStore) DeleteAll() error {
	m.data = make(map[string]string)
	return nil
}

func TestMain(m *testing.M) {
	// Prevent tests from opening a real browser.
	openBrowserFunc = func(_ string) {}
	m.Run()
}

func TestNewDeviceFlowConfig(t *testing.T) {
	t.Run("Defaults", func(t *testing.T) {
		c := NewDeviceFlowConfig()

		assert.Equal(t, DefaultClientID, c.clientID)
		assert.Equal(t, DefaultAuthBaseURL, c.authBaseURL)
		assert.Equal(t, DefaultKeyringServiceName, c.keyringServiceName)
	})

	t.Run("KeyRingServiceName", func(t *testing.T) {
		c := NewDeviceFlowConfig()
		assert.Equal(t, DefaultKeyringServiceName, c.KeyRingServiceName())
	})
}

func TestDeviceFlowConfig_With(t *testing.T) {
	t.Run("WithDeviceFlowAuthBaseURL", func(t *testing.T) {
		c := NewDeviceFlowConfig()
		c.With(WithDeviceFlowAuthBaseURL("https://custom.auth"))
		assert.Equal(t, "https://custom.auth", c.authBaseURL)
	})

	t.Run("WithDeviceFlowClientID", func(t *testing.T) {
		c := NewDeviceFlowConfig()
		c.With(WithDeviceFlowClientID("custom-id"))
		assert.Equal(t, "custom-id", c.clientID)
	})

	t.Run("WithDeviceFlowKeyringServiceName", func(t *testing.T) {
		c := NewDeviceFlowConfig()
		c.With(WithDeviceFlowKeyringServiceName("custom-keyring"))
		assert.Equal(t, "custom-keyring", c.keyringServiceName)
		assert.Equal(t, "custom-keyring", c.KeyRingServiceName())
	})

	t.Run("MultipleOptions", func(t *testing.T) {
		c := NewDeviceFlowConfig()
		c.With(
			WithDeviceFlowAuthBaseURL("https://auth2"),
			WithDeviceFlowClientID("id2"),
			WithDeviceFlowKeyringServiceName("kr2"),
		)
		assert.Equal(t, "https://auth2", c.authBaseURL)
		assert.Equal(t, "id2", c.clientID)
		assert.Equal(t, "kr2", c.keyringServiceName)
	})
}

func TestNewDeviceKeyringTokenSource(t *testing.T) {
	t.Run("ReturnsNonNil", func(t *testing.T) {
		store := newMemoryStore()
		c := NewDeviceFlowConfig()
		ts := NewDeviceKeyringTokenSource(context.Background(), c, store)
		require.NotNil(t, ts, "expected non-nil token source")
	})

	t.Run("ConfiguresOAuthEndpoints", func(t *testing.T) {
		store := newMemoryStore()
		c := NewDeviceFlowConfig()
		c.With(WithDeviceFlowAuthBaseURL("https://myauth.example.com"))
		ts := NewDeviceKeyringTokenSource(context.Background(), c, store)

		dts, ok := ts.(*deviceKeyringTokenSource)
		require.True(t, ok, "expected *deviceKeyringTokenSource")

		assert.Equal(t, "https://myauth.example.com/oauth2/auth", dts.conf.Endpoint.AuthURL)
		assert.Equal(t, "https://myauth.example.com/oauth2/device/auth", dts.conf.Endpoint.DeviceAuthURL)
		assert.Equal(t, "https://myauth.example.com/oauth2/token", dts.conf.Endpoint.TokenURL)
		assert.Equal(t, oauth2.AuthStyleInParams, dts.conf.Endpoint.AuthStyle)
	})

	t.Run("ConfiguresScopes", func(t *testing.T) {
		store := newMemoryStore()
		c := NewDeviceFlowConfig()
		ts := NewDeviceKeyringTokenSource(context.Background(), c, store)

		dts := ts.(*deviceKeyringTokenSource)
		assert.Contains(t, dts.conf.Scopes, "openid")
		assert.Contains(t, dts.conf.Scopes, "offline")
		assert.Len(t, dts.conf.Scopes, 2)
	})

	t.Run("UsesNamespacedTokenKey", func(t *testing.T) {
		store := newMemoryStore()
		c := NewDeviceFlowConfig()
		ts := NewDeviceKeyringTokenSource(context.Background(), c, store)

		dts := ts.(*deviceKeyringTokenSource)
		expected := namespacedTokenKey(DefaultDeviceFlowTokenKey, c.authBaseURL, c.clientID)
		assert.Equal(t, expected, dts.tokenKey)
		assert.NotEqual(t, DefaultDeviceFlowTokenKey, dts.tokenKey, "tokenKey must be namespaced, not the legacy fixed key")
		assert.Equal(t, DefaultDeviceFlowTokenKey, dts.legacyKey)
	})

	t.Run("ImplementsTokenSourceWithClaims", func(t *testing.T) {
		store := newMemoryStore()
		c := NewDeviceFlowConfig()
		var _ = NewDeviceKeyringTokenSource(context.Background(), c, store)
	})
}

func TestDeviceKeyringTokenSource_loadFromKeyring(t *testing.T) {
	t.Run("NotFound_ReturnsNil", func(t *testing.T) {
		store := newMemoryStore()
		dts := &deviceKeyringTokenSource{
			tokenKey: DefaultDeviceFlowTokenKey,
			store:    store,
		}

		token, err := dts.loadFromKeyring()
		assert.NoError(t, err)
		assert.Nil(t, token)
	})

	t.Run("EmptyString_ReturnsNil", func(t *testing.T) {
		store := newMemoryStore()
		_ = store.Set(DefaultDeviceFlowTokenKey, "")
		dts := &deviceKeyringTokenSource{
			tokenKey: DefaultDeviceFlowTokenKey,
			store:    store,
		}

		token, err := dts.loadFromKeyring()
		assert.NoError(t, err)
		assert.Nil(t, token)
	})

	t.Run("ValidToken_ReturnsToken", func(t *testing.T) {
		store := newMemoryStore()
		expected := &oauth2.Token{
			AccessToken:  "access-123",
			RefreshToken: "refresh-456",
			TokenType:    "Bearer",
			Expiry:       time.Now().Add(time.Hour),
		}
		data, _ := json.Marshal(expected)
		_ = store.Set(DefaultDeviceFlowTokenKey, string(data))

		dts := &deviceKeyringTokenSource{
			tokenKey: DefaultDeviceFlowTokenKey,
			store:    store,
		}

		token, err := dts.loadFromKeyring()
		require.NoError(t, err)
		require.NotNil(t, token)
		assert.Equal(t, "access-123", token.AccessToken)
		assert.Equal(t, "refresh-456", token.RefreshToken)
		assert.Equal(t, "Bearer", token.TokenType)
	})

	t.Run("InvalidJSON_ReturnsError_DeletesEntry", func(t *testing.T) {
		store := newMemoryStore()
		_ = store.Set(DefaultDeviceFlowTokenKey, "{not valid json")

		dts := &deviceKeyringTokenSource{
			tokenKey: DefaultDeviceFlowTokenKey,
			store:    store,
		}

		token, err := dts.loadFromKeyring()
		assert.Error(t, err)
		assert.Nil(t, token)

		// Verify the corrupt entry was deleted
		_, getErr := store.Get(DefaultDeviceFlowTokenKey)
		assert.ErrorIs(t, getErr, ErrNotFound)
	})
}

func TestDeviceKeyringTokenSource_saveToKeyring(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		store := newMemoryStore()
		dts := &deviceKeyringTokenSource{
			tokenKey: DefaultDeviceFlowTokenKey,
			store:    store,
		}

		token := &oauth2.Token{
			AccessToken:  "saved-token",
			RefreshToken: "saved-refresh",
			TokenType:    "Bearer",
		}

		err := dts.saveToKeyring(token)
		require.NoError(t, err)

		stored, getErr := store.Get(DefaultDeviceFlowTokenKey)
		require.NoError(t, getErr)

		var loaded oauth2.Token
		require.NoError(t, json.Unmarshal([]byte(stored), &loaded))
		assert.Equal(t, "saved-token", loaded.AccessToken)
		assert.Equal(t, "saved-refresh", loaded.RefreshToken)
	})
}

func TestPerformDeviceFlow(t *testing.T) {
	t.Run("Success_WithVerificationURIComplete", func(t *testing.T) {
		// Fake OAuth2 server that handles device auth and token exchange.
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			switch r.URL.Path {
			case "/device/auth":
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{
					"device_code": "dev-code-1",
					"user_code": "ABCD-1234",
					"verification_uri": "https://example.com/verify",
					"verification_uri_complete": "https://example.com/verify?code=ABCD-1234",
					"interval": 0,
					"expires_in": 300
				}`))
			case "/token":
				w.WriteHeader(http.StatusOK)
				_, _ = fmt.Fprintf(w, `{
					"access_token": "device-access-token",
					"token_type": "Bearer",
					"expires_in": 3600,
					"refresh_token": "device-refresh-token"
				}`)
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		}))
		defer srv.Close()

		conf := &oauth2.Config{
			ClientID: "test-client",
			Endpoint: oauth2.Endpoint{
				DeviceAuthURL: srv.URL + "/device/auth",
				TokenURL:      srv.URL + "/token",
				AuthStyle:     oauth2.AuthStyleInParams,
			},
		}

		token, err := PerformDeviceFlow(context.Background(), conf)
		require.NoError(t, err)
		require.NotNil(t, token)
		assert.Equal(t, "device-access-token", token.AccessToken)
		assert.Equal(t, "device-refresh-token", token.RefreshToken)
		assert.Equal(t, "Bearer", token.TokenType)
	})

	t.Run("Success_WithoutVerificationURIComplete", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			switch r.URL.Path {
			case "/device/auth":
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{
					"device_code": "dev-code-2",
					"user_code": "EFGH-5678",
					"verification_uri": "https://example.com/verify",
					"interval": 0,
					"expires_in": 300
				}`))
			case "/token":
				w.WriteHeader(http.StatusOK)
				_, _ = fmt.Fprintf(w, `{
					"access_token": "access-2",
					"token_type": "Bearer",
					"expires_in": 3600
				}`)
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		}))
		defer srv.Close()

		conf := &oauth2.Config{
			ClientID: "test-client",
			Endpoint: oauth2.Endpoint{
				DeviceAuthURL: srv.URL + "/device/auth",
				TokenURL:      srv.URL + "/token",
				AuthStyle:     oauth2.AuthStyleInParams,
			},
		}

		token, err := PerformDeviceFlow(context.Background(), conf)
		require.NoError(t, err)
		require.NotNil(t, token)
		assert.Equal(t, "access-2", token.AccessToken)
	})

	t.Run("DeviceAuth_Error", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer srv.Close()

		conf := &oauth2.Config{
			ClientID: "test-client",
			Endpoint: oauth2.Endpoint{
				DeviceAuthURL: srv.URL + "/device/auth",
				TokenURL:      srv.URL + "/token",
				AuthStyle:     oauth2.AuthStyleInParams,
			},
		}

		token, err := PerformDeviceFlow(context.Background(), conf)
		assert.Error(t, err)
		assert.Nil(t, token)
		assert.Contains(t, err.Error(), "device authorization failed")
	})

	t.Run("TokenExchange_Error", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			switch r.URL.Path {
			case "/device/auth":
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{
					"device_code": "dev-code-3",
					"user_code": "IJKL-9012",
					"verification_uri": "https://example.com/verify",
					"interval": 0,
					"expires_in": 300
				}`))
			case "/token":
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(`{"error":"access_denied"}`))
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		}))
		defer srv.Close()

		conf := &oauth2.Config{
			ClientID: "test-client",
			Endpoint: oauth2.Endpoint{
				DeviceAuthURL: srv.URL + "/device/auth",
				TokenURL:      srv.URL + "/token",
				AuthStyle:     oauth2.AuthStyleInParams,
			},
		}

		token, err := PerformDeviceFlow(context.Background(), conf)
		assert.Error(t, err)
		assert.Nil(t, token)
		assert.Contains(t, err.Error(), "device access token failed")
	})
}

func TestDeviceKeyringTokenSource_Token(t *testing.T) {
	t.Run("NoStoredToken_PerformsDeviceFlow", func(t *testing.T) {
		accessJWT := makeTestJWT(map[string]any{
			"eos": map[string]any{"realm": "realm-from-flow", "org": "org-from-flow"},
		})

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			switch r.URL.Path {
			case "/oauth2/device/auth":
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{
					"device_code": "dc-1",
					"user_code": "UC-1",
					"verification_uri": "https://example.com/verify",
					"interval": 0,
					"expires_in": 300
				}`))
			case "/oauth2/token":
				w.WriteHeader(http.StatusOK)
				_, _ = fmt.Fprintf(w, `{"access_token":%q,"token_type":"Bearer","expires_in":3600,"refresh_token":"new-refresh"}`, accessJWT)
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		}))
		defer srv.Close()

		store := newMemoryStore()
		c := NewDeviceFlowConfig()
		c.With(WithDeviceFlowAuthBaseURL(srv.URL))

		ts := NewDeviceKeyringTokenSource(context.Background(), c, store)
		token, err := ts.Token()
		require.NoError(t, err)
		require.NotNil(t, token)
		assert.Equal(t, accessJWT, token.AccessToken)

		// Verify token was saved to store under the namespaced key (not the legacy key).
		namespacedKey := namespacedTokenKey(DefaultDeviceFlowTokenKey, c.authBaseURL, c.clientID)
		stored, getErr := store.Get(namespacedKey)
		require.NoError(t, getErr)
		assert.NotEmpty(t, stored)
	})

	t.Run("ExpiredStoredToken_RefreshesSuccessfully", func(t *testing.T) {
		refreshedJWT := makeTestJWT(map[string]any{
			"eos": map[string]any{"realm": "realm-refreshed", "org": "org-refreshed"},
		})

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			switch r.URL.Path {
			case "/oauth2/token":
				w.WriteHeader(http.StatusOK)
				_, _ = fmt.Fprintf(w, `{"access_token":%q,"token_type":"Bearer","expires_in":3600,"refresh_token":"refreshed-refresh"}`, refreshedJWT)
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		}))
		defer srv.Close()

		c := NewDeviceFlowConfig()
		c.With(WithDeviceFlowAuthBaseURL(srv.URL))

		store := newMemoryStore()
		// Pre-store an expired token so refreshOrReauthenticate is triggered.
		existingToken := &oauth2.Token{
			AccessToken:  "old-access",
			RefreshToken: "old-refresh",
			TokenType:    "Bearer",
			Expiry:       time.Now().Add(-time.Hour), // expired
		}
		data, _ := json.Marshal(existingToken)
		_ = store.Set(namespacedTokenKey(DefaultDeviceFlowTokenKey, c.authBaseURL, c.clientID), string(data))

		ts := NewDeviceKeyringTokenSource(context.Background(), c, store)
		token, err := ts.Token()
		require.NoError(t, err)
		require.NotNil(t, token)
		assert.Equal(t, refreshedJWT, token.AccessToken)
	})

	t.Run("TokenMissingEosClaims_ReturnsError", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			switch r.URL.Path {
			case "/oauth2/device/auth":
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"device_code":"dc-1","user_code":"UC-1","verification_uri":"https://example.com/verify","interval":0,"expires_in":300}`))
			case "/oauth2/token":
				// Return a JWT without eos claims
				noClaimsJWT := makeTestJWT(map[string]any{"sub": "user-1"})
				w.WriteHeader(http.StatusOK)
				_, _ = fmt.Fprintf(w, `{"access_token":%q,"token_type":"Bearer","expires_in":3600}`, noClaimsJWT)
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		}))
		defer srv.Close()

		store := newMemoryStore()
		c := NewDeviceFlowConfig()
		c.With(WithDeviceFlowAuthBaseURL(srv.URL))

		ts := NewDeviceKeyringTokenSource(context.Background(), c, store)
		token, err := ts.Token()
		assert.Error(t, err)
		assert.Nil(t, token)
		assert.Contains(t, err.Error(), "eos claims")
	})

	t.Run("CorruptStoredToken_ReturnsError", func(t *testing.T) {
		c := NewDeviceFlowConfig()
		c.With(WithDeviceFlowAuthBaseURL("https://unused.example.com"))

		store := newMemoryStore()
		// Store corrupt JSON under the namespaced key to exercise the primary-key
		// corruption path; corrupt legacy entries are silently skipped (see legacy fallback).
		_ = store.Set(namespacedTokenKey(DefaultDeviceFlowTokenKey, c.authBaseURL, c.clientID), "{corrupt json")

		ts := NewDeviceKeyringTokenSource(context.Background(), c, store)
		token, err := ts.Token()
		assert.Error(t, err)
		assert.Nil(t, token)
	})

	t.Run("NearExpiryStoredToken_RefreshesWithRefreshToken_NeverInteractive", func(t *testing.T) {
		// A token within the 60-second buffer must be silently refreshed via the
		// refresh_token grant — no device-flow (interactive) endpoints are called.
		refreshedJWT := makeTestJWT(map[string]any{
			"eos": map[string]any{"realm": "realm-near-expiry", "org": "org-near-expiry"},
		})

		deviceFlowCalled := false
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			switch r.URL.Path {
			case "/oauth2/device/auth":
				// Must NOT be reached for a near-expiry token with a valid refresh token.
				deviceFlowCalled = true
				w.WriteHeader(http.StatusInternalServerError)
			case "/oauth2/token":
				w.WriteHeader(http.StatusOK)
				_, _ = fmt.Fprintf(w, `{"access_token":%q,"token_type":"Bearer","expires_in":3600,"refresh_token":"rotated-refresh"}`, refreshedJWT)
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		}))
		defer srv.Close()

		c := NewDeviceFlowConfig()
		c.With(WithDeviceFlowAuthBaseURL(srv.URL))

		store := newMemoryStore()
		// Store a token expiring in 30s — within the 60s buffer.
		nearExpiryToken := &oauth2.Token{
			AccessToken:  "old-access",
			RefreshToken: "old-refresh",
			TokenType:    "Bearer",
			Expiry:       time.Now().Add(30 * time.Second),
		}
		data, _ := json.Marshal(nearExpiryToken)
		namespacedKey := namespacedTokenKey(DefaultDeviceFlowTokenKey, c.authBaseURL, c.clientID)
		_ = store.Set(namespacedKey, string(data))

		ts := NewDeviceKeyringTokenSource(context.Background(), c, store)
		token, err := ts.Token()
		require.NoError(t, err)
		require.NotNil(t, token)
		assert.Equal(t, refreshedJWT, token.AccessToken, "expected the refreshed token")
		assert.False(t, deviceFlowCalled, "device flow must not be triggered for a near-expiry refresh")

		// The rotated token must be persisted to the store under the namespaced key.
		stored, getErr := store.Get(namespacedKey)
		require.NoError(t, getErr)
		var persisted oauth2.Token
		require.NoError(t, json.Unmarshal([]byte(stored), &persisted))
		assert.Equal(t, refreshedJWT, persisted.AccessToken)
		assert.Equal(t, "rotated-refresh", persisted.RefreshToken)
	})

	t.Run("InvalidGrantOnRefresh_FallsBackToDeviceFlow", func(t *testing.T) {
		// When the refresh_token is rejected with invalid_grant, the code must
		// fall back to a full interactive device flow.
		freshJWT := makeTestJWT(map[string]any{
			"eos": map[string]any{"realm": "realm-reauthed", "org": "org-reauthed"},
		})

		tokenCallCount := 0
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			switch r.URL.Path {
			case "/oauth2/device/auth":
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{
					"device_code": "dc-fallback",
					"user_code": "UC-FB",
					"verification_uri": "https://example.com/verify",
					"interval": 0,
					"expires_in": 300
				}`))
			case "/oauth2/token":
				tokenCallCount++
				if tokenCallCount == 1 {
					// First call is the refresh attempt — return invalid_grant.
					w.WriteHeader(http.StatusBadRequest)
					_, _ = w.Write([]byte(`{"error":"invalid_grant","error_description":"refresh token expired"}`))
					return
				}
				// Second call is from the device flow.
				w.WriteHeader(http.StatusOK)
				_, _ = fmt.Fprintf(w, `{"access_token":%q,"token_type":"Bearer","expires_in":3600,"refresh_token":"brand-new-refresh"}`, freshJWT)
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		}))
		defer srv.Close()

		c := NewDeviceFlowConfig()
		c.With(WithDeviceFlowAuthBaseURL(srv.URL))

		store := newMemoryStore()
		expiredToken := &oauth2.Token{
			AccessToken:  "stale-access",
			RefreshToken: "dead-refresh",
			TokenType:    "Bearer",
			Expiry:       time.Now().Add(-time.Hour),
		}
		data, _ := json.Marshal(expiredToken)
		_ = store.Set(namespacedTokenKey(DefaultDeviceFlowTokenKey, c.authBaseURL, c.clientID), string(data))

		ts := NewDeviceKeyringTokenSource(context.Background(), c, store)
		token, err := ts.Token()
		require.NoError(t, err)
		require.NotNil(t, token)
		assert.Equal(t, freshJWT, token.AccessToken, "expected fresh token from device flow fallback")
		assert.Equal(t, 2, tokenCallCount, "expected two token-endpoint calls: refresh attempt + device flow")
	})

	t.Run("WellBeforeExpiryToken_ReturnedWithoutRefresh", func(t *testing.T) {
		// A token that is well before expiry (more than 60s remaining) must be
		// returned as-is without any network call.
		freshJWT := makeTestJWT(map[string]any{
			"eos": map[string]any{"realm": "realm-fresh", "org": "org-fresh"},
		})

		serverCalled := false
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			serverCalled = true
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer srv.Close()

		c := NewDeviceFlowConfig()
		c.With(WithDeviceFlowAuthBaseURL(srv.URL))

		store := newMemoryStore()
		validToken := &oauth2.Token{
			AccessToken:  freshJWT,
			RefreshToken: "valid-refresh",
			TokenType:    "Bearer",
			Expiry:       time.Now().Add(10 * time.Minute), // well within validity
		}
		data, _ := json.Marshal(validToken)
		_ = store.Set(namespacedTokenKey(DefaultDeviceFlowTokenKey, c.authBaseURL, c.clientID), string(data))

		ts := NewDeviceKeyringTokenSource(context.Background(), c, store)
		token, err := ts.Token()
		require.NoError(t, err)
		require.NotNil(t, token)
		assert.Equal(t, freshJWT, token.AccessToken)
		assert.False(t, serverCalled, "server must not be called for a token well within validity")
	})
}

func TestDeviceKeyringTokenSource_TokenClaims(t *testing.T) {
	t.Run("ReturnsCachedClaimsWithoutExtraTokenCall", func(t *testing.T) {
		accessJWT := makeTestJWT(map[string]any{
			"eos": map[string]any{"realm": "realm-abc", "org": "org-xyz"},
		})

		callCount := 0
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			switch r.URL.Path {
			case "/oauth2/device/auth":
				callCount++
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"device_code":"dc-1","user_code":"UC-1","verification_uri":"https://example.com/verify","interval":0,"expires_in":300}`))
			case "/oauth2/token":
				callCount++
				w.WriteHeader(http.StatusOK)
				_, _ = fmt.Fprintf(w, `{"access_token":%q,"token_type":"Bearer","expires_in":3600}`, accessJWT)
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		}))
		defer srv.Close()

		store := newMemoryStore()
		c := NewDeviceFlowConfig()
		c.With(WithDeviceFlowAuthBaseURL(srv.URL))
		ts := NewDeviceKeyringTokenSource(context.Background(), c, store)

		// First call triggers device flow
		claims, err := ts.TokenClaims()
		require.NoError(t, err)
		require.NotNil(t, claims)
		assert.Equal(t, "realm-abc", claims.RealmID)
		assert.Equal(t, "org-xyz", claims.OrganizationID)
		firstCallCount := callCount

		// Second call must return cached claims without hitting the server again
		claims2, err := ts.TokenClaims()
		require.NoError(t, err)
		assert.Equal(t, claims, claims2)
		assert.Equal(t, firstCallCount, callCount, "server should not be called again for cached claims")
	})

	t.Run("TriggersTokenIfNotYetCalled", func(t *testing.T) {
		accessJWT := makeTestJWT(map[string]any{
			"eos": map[string]any{"realm": "lazy-realm", "org": "lazy-org"},
		})

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			switch r.URL.Path {
			case "/oauth2/device/auth":
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"device_code":"dc-1","user_code":"UC-1","verification_uri":"https://example.com/verify","interval":0,"expires_in":300}`))
			case "/oauth2/token":
				w.WriteHeader(http.StatusOK)
				_, _ = fmt.Fprintf(w, `{"access_token":%q,"token_type":"Bearer","expires_in":3600}`, accessJWT)
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		}))
		defer srv.Close()

		store := newMemoryStore()
		c := NewDeviceFlowConfig()
		c.With(WithDeviceFlowAuthBaseURL(srv.URL))
		ts := NewDeviceKeyringTokenSource(context.Background(), c, store)

		// TokenClaims() without a prior Token() call must trigger auth and return claims
		claims, err := ts.TokenClaims()
		require.NoError(t, err)
		assert.Equal(t, "lazy-realm", claims.RealmID)
	})
}
