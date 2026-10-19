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

func TestNewClientCredentialsConfig(t *testing.T) {
	t.Run("Defaults", func(t *testing.T) {
		c := NewClientCredentialsConfig("cid", "csecret")

		assert.Equal(t, "cid", c.clientID)
		assert.Equal(t, "csecret", c.clientSecret)
		assert.Equal(t, DefaultAuthBaseURL, c.authBaseURL)
		assert.Equal(t, DefaultKeyringServiceName, c.keyringServiceName)
	})

	t.Run("KeyRingServiceName", func(t *testing.T) {
		c := NewClientCredentialsConfig("cid", "csecret")
		assert.Equal(t, DefaultKeyringServiceName, c.KeyRingServiceName())
	})
}

func TestClientCredentialsConfig_With(t *testing.T) {
	t.Run("WithClientCredentialsAuthBaseURL", func(t *testing.T) {
		c := NewClientCredentialsConfig("cid", "csecret")
		c.With(WithClientCredentialsAuthBaseURL("https://custom.auth"))
		assert.Equal(t, "https://custom.auth", c.authBaseURL)
	})

	t.Run("WithClientCredentialsKeyringServiceName", func(t *testing.T) {
		c := NewClientCredentialsConfig("cid", "csecret")
		c.With(WithClientCredentialsKeyringServiceName("custom-kr"))
		assert.Equal(t, "custom-kr", c.keyringServiceName)
		assert.Equal(t, "custom-kr", c.KeyRingServiceName())
	})

	t.Run("MultipleOptions", func(t *testing.T) {
		c := NewClientCredentialsConfig("cid", "csecret")
		c.With(
			WithClientCredentialsAuthBaseURL("https://auth2"),
			WithClientCredentialsKeyringServiceName("kr2"),
		)
		assert.Equal(t, "https://auth2", c.authBaseURL)
		assert.Equal(t, "kr2", c.keyringServiceName)
	})
}

func TestClientCredentialsConfig_GetClientCredentialsConfig(t *testing.T) {
	c := NewClientCredentialsConfig("my-id", "my-secret")
	c.With(WithClientCredentialsAuthBaseURL("https://auth.test"))

	cc := c.GetClientCredentialsConfig()
	assert.Equal(t, "my-id", cc.ClientID)
	assert.Equal(t, "my-secret", cc.ClientSecret)
	assert.Equal(t, "https://auth.test/oauth2/token", cc.TokenURL)
}

func TestNewClientCredentialsKeyringTokenSource(t *testing.T) {
	t.Run("ReturnsNonNil", func(t *testing.T) {
		store := newMemoryStore()
		c := NewClientCredentialsConfig("cid", "csecret")
		ts := NewClientCredentialsKeyringTokenSource(context.Background(), c, store)
		require.NotNil(t, ts)
	})

	t.Run("ConfiguresInternals", func(t *testing.T) {
		store := newMemoryStore()
		c := NewClientCredentialsConfig("cid", "csecret")
		c.With(WithClientCredentialsAuthBaseURL("https://myauth.example.com"))
		ts := NewClientCredentialsKeyringTokenSource(context.Background(), c, store)

		ccts, ok := ts.(*clientCredentialsKeyringTokenSource)
		require.True(t, ok)
		expected := namespacedTokenKey(DefaultClientCredentialsFlowTokenKey, c.authBaseURL, c.clientID)
		assert.Equal(t, expected, ccts.tokenKey)
		assert.NotEqual(t, DefaultClientCredentialsFlowTokenKey, ccts.tokenKey, "tokenKey must be namespaced, not the legacy fixed key")
		assert.Equal(t, DefaultClientCredentialsFlowTokenKey, ccts.legacyKey)
		assert.Equal(t, "https://myauth.example.com/oauth2/token", ccts.conf.TokenURL)
		assert.Equal(t, "cid", ccts.conf.ClientID)
		assert.Equal(t, "csecret", ccts.conf.ClientSecret)
	})

	t.Run("ImplementsTokenSourceWithClaims", func(t *testing.T) {
		store := newMemoryStore()
		c := NewClientCredentialsConfig("cid", "csecret")
		var _ = NewClientCredentialsKeyringTokenSource(context.Background(), c, store)
	})
}

func TestClientCredentialsKeyringTokenSource_loadFromKeyring(t *testing.T) {
	t.Run("NotFound_ReturnsNil", func(t *testing.T) {
		store := newMemoryStore()
		s := &clientCredentialsKeyringTokenSource{

			tokenKey: DefaultClientCredentialsFlowTokenKey,
			store:    store,
		}

		token, err := s.loadFromKeyring()
		assert.NoError(t, err)
		assert.Nil(t, token)
	})

	t.Run("EmptyString_ReturnsNil", func(t *testing.T) {
		store := newMemoryStore()
		_ = store.Set(DefaultClientCredentialsFlowTokenKey, "")
		s := &clientCredentialsKeyringTokenSource{

			tokenKey: DefaultClientCredentialsFlowTokenKey,
			store:    store,
		}

		token, err := s.loadFromKeyring()
		assert.NoError(t, err)
		assert.Nil(t, token)
	})

	t.Run("ValidToken_ReturnsToken", func(t *testing.T) {
		store := newMemoryStore()
		expected := &oauth2.Token{
			AccessToken: "cc-access",
			TokenType:   "Bearer",
			Expiry:      time.Now().Add(time.Hour),
		}
		data, _ := json.Marshal(expected)
		_ = store.Set(DefaultClientCredentialsFlowTokenKey, string(data))

		s := &clientCredentialsKeyringTokenSource{

			tokenKey: DefaultClientCredentialsFlowTokenKey,
			store:    store,
		}

		token, err := s.loadFromKeyring()
		require.NoError(t, err)
		require.NotNil(t, token)
		assert.Equal(t, "cc-access", token.AccessToken)
		assert.Equal(t, "Bearer", token.TokenType)
	})

	t.Run("InvalidJSON_ReturnsError_DeletesEntry", func(t *testing.T) {
		store := newMemoryStore()
		_ = store.Set(DefaultClientCredentialsFlowTokenKey, "{bad json")

		s := &clientCredentialsKeyringTokenSource{

			tokenKey: DefaultClientCredentialsFlowTokenKey,
			store:    store,
		}

		token, err := s.loadFromKeyring()
		assert.Error(t, err)
		assert.Nil(t, token)

		_, getErr := store.Get(DefaultClientCredentialsFlowTokenKey)
		assert.ErrorIs(t, getErr, ErrNotFound)
	})
}

func TestClientCredentialsKeyringTokenSource_saveToKeyring(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		store := newMemoryStore()
		s := &clientCredentialsKeyringTokenSource{

			tokenKey: DefaultClientCredentialsFlowTokenKey,
			store:    store,
		}

		token := &oauth2.Token{
			AccessToken: "saved-cc",
			TokenType:   "Bearer",
		}

		err := s.saveToKeyring(token)
		require.NoError(t, err)

		stored, getErr := store.Get(DefaultClientCredentialsFlowTokenKey)
		require.NoError(t, getErr)

		var loaded oauth2.Token
		require.NoError(t, json.Unmarshal([]byte(stored), &loaded))
		assert.Equal(t, "saved-cc", loaded.AccessToken)
	})
}

func TestClientCredentialsKeyringTokenSource_Token(t *testing.T) {
	t.Run("NoStoredToken_FetchesNew", func(t *testing.T) {
		freshJWT := makeTestJWT(map[string]any{
			"eos": map[string]any{"realm": "realm-fresh", "org": "org-fresh"},
		})

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = fmt.Fprintf(w, `{"access_token":%q,"token_type":"Bearer","expires_in":3600}`, freshJWT)
		}))
		defer srv.Close()

		store := newMemoryStore()
		c := NewClientCredentialsConfig("cid", "csecret")
		c.With(WithClientCredentialsAuthBaseURL(srv.URL))

		ts := NewClientCredentialsKeyringTokenSource(context.Background(), c, store)
		token, err := ts.Token()
		require.NoError(t, err)
		require.NotNil(t, token)
		assert.Equal(t, freshJWT, token.AccessToken)

		// Verify token was persisted under the namespaced key (not the legacy key).
		namespacedKey := namespacedTokenKey(DefaultClientCredentialsFlowTokenKey, c.authBaseURL, c.clientID)
		stored, getErr := store.Get(namespacedKey)
		require.NoError(t, getErr)
		assert.NotEmpty(t, stored)
	})

	t.Run("ValidStoredToken_ReturnsCached", func(t *testing.T) {
		cachedJWT := makeTestJWT(map[string]any{
			"eos": map[string]any{"realm": "realm-cached", "org": "org-cached"},
		})

		// Server should NOT be called if cached token is still valid.
		serverCalled := false
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			serverCalled = true
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer srv.Close()

		c := NewClientCredentialsConfig("cid", "csecret")
		c.With(WithClientCredentialsAuthBaseURL(srv.URL))

		store := newMemoryStore()
		validToken := &oauth2.Token{
			AccessToken: cachedJWT,
			TokenType:   "Bearer",
			Expiry:      time.Now().Add(time.Hour),
		}
		data, _ := json.Marshal(validToken)
		_ = store.Set(namespacedTokenKey(DefaultClientCredentialsFlowTokenKey, c.authBaseURL, c.clientID), string(data))

		ts := NewClientCredentialsKeyringTokenSource(context.Background(), c, store)
		token, err := ts.Token()
		require.NoError(t, err)
		require.NotNil(t, token)
		assert.Equal(t, cachedJWT, token.AccessToken)
		assert.False(t, serverCalled, "server should not be called for a valid cached token")
	})

	t.Run("ExpiredStoredToken_FetchesNew", func(t *testing.T) {
		renewedJWT := makeTestJWT(map[string]any{
			"eos": map[string]any{"realm": "realm-renewed", "org": "org-renewed"},
		})

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = fmt.Fprintf(w, `{"access_token":%q,"token_type":"Bearer","expires_in":3600}`, renewedJWT)
		}))
		defer srv.Close()

		c := NewClientCredentialsConfig("cid", "csecret")
		c.With(WithClientCredentialsAuthBaseURL(srv.URL))

		store := newMemoryStore()
		expiredToken := &oauth2.Token{
			AccessToken: "old-token",
			TokenType:   "Bearer",
			Expiry:      time.Now().Add(-time.Hour),
		}
		data, _ := json.Marshal(expiredToken)
		_ = store.Set(namespacedTokenKey(DefaultClientCredentialsFlowTokenKey, c.authBaseURL, c.clientID), string(data))

		ts := NewClientCredentialsKeyringTokenSource(context.Background(), c, store)
		token, err := ts.Token()
		require.NoError(t, err)
		require.NotNil(t, token)
		assert.Equal(t, renewedJWT, token.AccessToken)
	})

	t.Run("TokenMissingEosClaims_ReturnsError", func(t *testing.T) {
		noClaimsJWT := makeTestJWT(map[string]any{"sub": "service-user"})

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = fmt.Fprintf(w, `{"access_token":%q,"token_type":"Bearer","expires_in":3600}`, noClaimsJWT)
		}))
		defer srv.Close()

		store := newMemoryStore()
		c := NewClientCredentialsConfig("cid", "csecret")
		c.With(WithClientCredentialsAuthBaseURL(srv.URL))

		ts := NewClientCredentialsKeyringTokenSource(context.Background(), c, store)
		token, err := ts.Token()
		assert.Error(t, err)
		assert.Nil(t, token)
		assert.Contains(t, err.Error(), "eos claims")
	})

	t.Run("CorruptStoredToken_ReturnsError", func(t *testing.T) {
		c := NewClientCredentialsConfig("cid", "csecret")
		c.With(WithClientCredentialsAuthBaseURL("https://unused.example.com"))

		store := newMemoryStore()
		// Store corrupt JSON under the namespaced key; corrupt legacy entries are
		// silently skipped by the read-only fallback.
		_ = store.Set(namespacedTokenKey(DefaultClientCredentialsFlowTokenKey, c.authBaseURL, c.clientID), "{corrupt")

		ts := NewClientCredentialsKeyringTokenSource(context.Background(), c, store)
		token, err := ts.Token()
		assert.Error(t, err)
		assert.Nil(t, token)
	})

	t.Run("FetchNewToken_ServerError", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error":"invalid_client"}`))
		}))
		defer srv.Close()

		store := newMemoryStore()
		c := NewClientCredentialsConfig("bad-id", "bad-secret")
		c.With(WithClientCredentialsAuthBaseURL(srv.URL))

		ts := NewClientCredentialsKeyringTokenSource(context.Background(), c, store)
		token, err := ts.Token()
		assert.Error(t, err)
		assert.Nil(t, token)
	})
}

func TestClientCredentialsKeyringTokenSource_TokenClaims(t *testing.T) {
	t.Run("ReturnsCachedClaimsWithoutExtraServerCall", func(t *testing.T) {
		accessJWT := makeTestJWT(map[string]any{
			"eos": map[string]any{"realm": "realm-cc", "org": "org-cc"},
		})

		callCount := 0
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callCount++
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = fmt.Fprintf(w, `{"access_token":%q,"token_type":"Bearer","expires_in":3600}`, accessJWT)
		}))
		defer srv.Close()

		store := newMemoryStore()
		c := NewClientCredentialsConfig("cid", "csecret")
		c.With(WithClientCredentialsAuthBaseURL(srv.URL))
		ts := NewClientCredentialsKeyringTokenSource(context.Background(), c, store)

		// First call fetches token and caches claims
		claims, err := ts.TokenClaims()
		require.NoError(t, err)
		require.NotNil(t, claims)
		assert.Equal(t, "realm-cc", claims.RealmID)
		assert.Equal(t, "org-cc", claims.OrganizationID)
		firstCallCount := callCount

		// Second call returns cached claims without hitting the server
		claims2, err := ts.TokenClaims()
		require.NoError(t, err)
		assert.Equal(t, claims, claims2)
		assert.Equal(t, firstCallCount, callCount, "server should not be called again for cached claims")
	})

	t.Run("TriggersTokenIfNotYetCalled", func(t *testing.T) {
		accessJWT := makeTestJWT(map[string]any{
			"eos": map[string]any{"realm": "lazy-cc-realm", "org": "lazy-cc-org"},
		})

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = fmt.Fprintf(w, `{"access_token":%q,"token_type":"Bearer","expires_in":3600}`, accessJWT)
		}))
		defer srv.Close()

		store := newMemoryStore()
		c := NewClientCredentialsConfig("cid", "csecret")
		c.With(WithClientCredentialsAuthBaseURL(srv.URL))
		ts := NewClientCredentialsKeyringTokenSource(context.Background(), c, store)

		claims, err := ts.TokenClaims()
		require.NoError(t, err)
		assert.Equal(t, "lazy-cc-realm", claims.RealmID)
	})
}
