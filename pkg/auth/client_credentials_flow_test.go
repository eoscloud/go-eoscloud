package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
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

		// SDK-13: client credentials tokens are never persisted.
		assert.Empty(t, store.data, "client credentials token must stay in memory")
	})

	t.Run("InMemoryToken_ReusedWithoutServerCall", func(t *testing.T) {
		jwt := makeTestJWT(map[string]any{
			"eos": map[string]any{"realm": "realm-cached", "org": "org-cached"},
		})

		var calls atomic.Int32
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			calls.Add(1)
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprintf(w, `{"access_token":%q,"token_type":"Bearer","expires_in":3600}`, jwt)
		}))
		defer srv.Close()

		c := NewClientCredentialsConfig("cid", "csecret")
		c.With(WithClientCredentialsAuthBaseURL(srv.URL))

		ts := NewClientCredentialsKeyringTokenSource(context.Background(), c, newMemoryStore())
		_, err := ts.Token()
		require.NoError(t, err)
		_, err = ts.Token()
		require.NoError(t, err)
		assert.Equal(t, int32(1), calls.Load(), "second call must reuse the in-memory token")
	})

	t.Run("StoredToken_IsIgnored", func(t *testing.T) {
		freshJWT := makeTestJWT(map[string]any{
			"eos": map[string]any{"realm": "realm-fresh", "org": "org-fresh"},
		})
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprintf(w, `{"access_token":%q,"token_type":"Bearer","expires_in":3600}`, freshJWT)
		}))
		defer srv.Close()

		c := NewClientCredentialsConfig("cid", "csecret")
		c.With(WithClientCredentialsAuthBaseURL(srv.URL))

		store := newMemoryStore()
		key := namespacedTokenKey(DefaultClientCredentialsFlowTokenKey, c.authBaseURL, c.clientID)
		data, _ := json.Marshal(&oauth2.Token{AccessToken: "persisted", Expiry: time.Now().Add(time.Hour)})
		_ = store.Set(key, string(data))
		_ = store.Set(DefaultClientCredentialsFlowTokenKey, "{corrupt")

		ts := NewClientCredentialsKeyringTokenSource(context.Background(), c, store)
		token, err := ts.Token()
		require.NoError(t, err)
		assert.Equal(t, freshJWT, token.AccessToken)
		assert.Equal(t, string(data), store.data[key], "store must not be written")

		require.NoError(t, ts.ClearTokenStore())
		_, err = store.Get(key)
		assert.ErrorIs(t, err, ErrNotFound, "ClearTokenStore removes legacy persisted entry")
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
