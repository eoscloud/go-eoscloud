package auth

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

// makeTestJWTWithIss returns a test JWT whose registered "iss" claim is set to
// the given issuer. Use this for legacy-fallback tests that need issuer matching.
func makeTestJWTWithIss(iss string, extra map[string]any) string {
	payload := make(map[string]any)
	for k, v := range extra {
		payload[k] = v
	}
	payload["iss"] = iss
	return makeTestJWT(payload)
}

// TestNamespacedTokenKey verifies the helper's contract: same inputs → same
// key, different authBaseURL or clientID → different key, prefix preserved.
func TestNamespacedTokenKey(t *testing.T) {
	t.Run("SameInputs_SameKey", func(t *testing.T) {
		k1 := namespacedTokenKey("eosctl-device", "https://auth.example.com", "cid")
		k2 := namespacedTokenKey("eosctl-device", "https://auth.example.com", "cid")
		assert.Equal(t, k1, k2)
	})

	t.Run("DifferentAuthBaseURL_DifferentKey", func(t *testing.T) {
		k1 := namespacedTokenKey("eosctl-device", "https://auth-a.example.com", "cid")
		k2 := namespacedTokenKey("eosctl-device", "https://auth-b.example.com", "cid")
		assert.NotEqual(t, k1, k2)
	})

	t.Run("DifferentClientID_DifferentKey", func(t *testing.T) {
		k1 := namespacedTokenKey("eosctl-device", "https://auth.example.com", "client-a")
		k2 := namespacedTokenKey("eosctl-device", "https://auth.example.com", "client-b")
		assert.NotEqual(t, k1, k2)
	})

	t.Run("KeyHasLegacyPrefix", func(t *testing.T) {
		k := namespacedTokenKey("eosctl-device", "https://auth.example.com", "cid")
		assert.True(t, len(k) > len("eosctl-device-"), "key should have prefix + 16 hex chars")
		assert.Equal(t, "eosctl-device-", k[:len("eosctl-device-")])
	})

	t.Run("HashSuffix_Is16HexChars", func(t *testing.T) {
		k := namespacedTokenKey("base", "https://auth.example.com", "cid")
		suffix := k[len("base-"):]
		assert.Len(t, suffix, 16)
		for _, c := range suffix {
			assert.True(t, (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f'), "suffix must be lowercase hex")
		}
	})
}

// TestIssuerMatchesAuthBaseURL verifies normalization and comparison rules.
func TestIssuerMatchesAuthBaseURL(t *testing.T) {
	tests := []struct {
		name        string
		issuer      string
		authBaseURL string
		want        bool
	}{
		{"exact match", "https://auth.example.com", "https://auth.example.com", true},
		{"trailing slash on issuer", "https://auth.example.com/", "https://auth.example.com", true},
		{"trailing slash on baseURL", "https://auth.example.com", "https://auth.example.com/", true},
		{"both trailing slash", "https://auth.example.com/", "https://auth.example.com/", true},
		{"case insensitive", "HTTPS://Auth.Example.Com", "https://auth.example.com", true},
		{"different host", "https://auth-a.example.com", "https://auth-b.example.com", false},
		{"empty issuer", "", "https://auth.example.com", false},
		{"empty baseURL", "https://auth.example.com", "", false},
		{"both empty", "", "", false},
		{"different scheme", "http://auth.example.com", "https://auth.example.com", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := issuerMatchesAuthBaseURL(tc.issuer, tc.authBaseURL)
			assert.Equal(t, tc.want, got)
		})
	}
}

// TestTokenIssuer verifies the helper extracts the "iss" claim without
// requiring the "eos" claim (unlike ExtractClaims).
func TestTokenIssuer(t *testing.T) {
	t.Run("ExtractsIssuer", func(t *testing.T) {
		token := &oauth2.Token{
			AccessToken: makeTestJWT(map[string]any{
				"iss": "https://auth.example.com",
				"sub": "user-1",
			}),
		}
		iss, err := tokenIssuer(token)
		require.NoError(t, err)
		assert.Equal(t, "https://auth.example.com", iss)
	})

	t.Run("NoIssClaim_ReturnsEmptyString", func(t *testing.T) {
		token := &oauth2.Token{
			AccessToken: makeTestJWT(map[string]any{"sub": "user-1"}),
		}
		iss, err := tokenIssuer(token)
		require.NoError(t, err)
		assert.Equal(t, "", iss)
	})

	t.Run("NoEosClaimRequired", func(t *testing.T) {
		// tokenIssuer must work even when the eos claim is absent (unlike ExtractClaims).
		token := &oauth2.Token{
			AccessToken: makeTestJWT(map[string]any{
				"iss": "https://auth.example.com",
			}),
		}
		_, err := tokenIssuer(token)
		require.NoError(t, err)
	})

	t.Run("InvalidJWT_ReturnsError", func(t *testing.T) {
		token := &oauth2.Token{AccessToken: "not.a.jwt"}
		_, err := tokenIssuer(token)
		assert.Error(t, err)
	})
}

// TestTokenKeyNamespacing verifies that two sources built with different
// authBaseURLs produce different primary token keys.
func TestTokenKeyNamespacing(t *testing.T) {
	t.Run("DeviceFlow_DifferentAuthURLs_DifferentKeys", func(t *testing.T) {
		storeA := newMemoryStore()
		cA := NewDeviceFlowConfig()
		cA.With(WithDeviceFlowAuthBaseURL("https://auth-a.example.com"))
		tsA := NewDeviceKeyringTokenSource(context.Background(), cA, storeA).(*deviceKeyringTokenSource)

		storeB := newMemoryStore()
		cB := NewDeviceFlowConfig()
		cB.With(WithDeviceFlowAuthBaseURL("https://auth-b.example.com"))
		tsB := NewDeviceKeyringTokenSource(context.Background(), cB, storeB).(*deviceKeyringTokenSource)

		assert.NotEqual(t, tsA.tokenKey, tsB.tokenKey, "different auth URLs must produce different token keys")
	})

	t.Run("ClientCredentials_DifferentAuthURLs_DifferentKeys", func(t *testing.T) {
		storeA := newMemoryStore()
		cA := NewClientCredentialsConfig("cid", "secret")
		cA.With(WithClientCredentialsAuthBaseURL("https://auth-a.example.com"))
		tsA := NewClientCredentialsKeyringTokenSource(context.Background(), cA, storeA).(*clientCredentialsKeyringTokenSource)

		storeB := newMemoryStore()
		cB := NewClientCredentialsConfig("cid", "secret")
		cB.With(WithClientCredentialsAuthBaseURL("https://auth-b.example.com"))
		tsB := NewClientCredentialsKeyringTokenSource(context.Background(), cB, storeB).(*clientCredentialsKeyringTokenSource)

		assert.NotEqual(t, tsA.tokenKey, tsB.tokenKey)
	})
}

// TestLegacyTokenMigration covers the read-only legacy fallback with upgrade-on-use.
func TestLegacyTokenMigration(t *testing.T) {
	const authBaseURL = "https://auth.example.com"
	const clientID = DefaultClientID

	// sameDomainLegacyToken builds a token whose access JWT has iss = authBaseURL.
	sameDomainLegacyToken := func() *oauth2.Token {
		return &oauth2.Token{
			AccessToken: makeTestJWTWithIss(authBaseURL, map[string]any{
				"eos": map[string]any{"realm": "realm-legacy", "org": "org-legacy"},
			}),
			RefreshToken: "legacy-refresh",
			TokenType:    "Bearer",
			Expiry:       time.Now().Add(10 * time.Minute),
		}
	}

	// crossDomainLegacyToken builds a token whose iss does NOT match authBaseURL.
	crossDomainLegacyToken := func() *oauth2.Token {
		return &oauth2.Token{
			AccessToken: makeTestJWTWithIss("https://other-domain.example.com", map[string]any{
				"eos": map[string]any{"realm": "realm-other", "org": "org-other"},
			}),
			RefreshToken: "other-refresh",
			TokenType:    "Bearer",
			Expiry:       time.Now().Add(10 * time.Minute),
		}
	}

	t.Run("Device_SameDomain_Adopted_MigratedToNamespacedKey_LegacyDeleted", func(t *testing.T) {
		legacy := sameDomainLegacyToken()
		legacyData, _ := json.Marshal(legacy)

		store := newMemoryStore()
		_ = store.Set(DefaultDeviceFlowTokenKey, string(legacyData))

		dts := &deviceKeyringTokenSource{
			tokenKey:    namespacedTokenKey(DefaultDeviceFlowTokenKey, authBaseURL, clientID),
			legacyKey:   DefaultDeviceFlowTokenKey,
			authBaseURL: authBaseURL,
			store:       store,
		}

		// loadFromKeyring must return the legacy token (adopted).
		got, err := dts.loadFromKeyring()
		require.NoError(t, err)
		require.NotNil(t, got, "same-domain legacy token must be adopted")
		assert.Equal(t, legacy.AccessToken, got.AccessToken)

		// Adoption must migrate the token to the namespaced key immediately.
		stored, err := store.Get(dts.tokenKey)
		require.NoError(t, err)
		assert.Equal(t, string(legacyData), stored, "adopted token must be written under the namespaced key")

		// The legacy entry must be deleted after a successful migration.
		_, err = store.Get(DefaultDeviceFlowTokenKey)
		assert.ErrorIs(t, err, ErrNotFound, "legacy entry must be deleted after adoption")
	})

	t.Run("Device_CrossDomain_NotAdopted_FreshAuthRequired", func(t *testing.T) {
		legacy := crossDomainLegacyToken()
		legacyData, _ := json.Marshal(legacy)

		store := newMemoryStore()
		_ = store.Set(DefaultDeviceFlowTokenKey, string(legacyData))

		dts := &deviceKeyringTokenSource{
			tokenKey:    namespacedTokenKey(DefaultDeviceFlowTokenKey, authBaseURL, clientID),
			legacyKey:   DefaultDeviceFlowTokenKey,
			authBaseURL: authBaseURL,
			store:       store,
		}

		got, err := dts.loadFromKeyring()
		require.NoError(t, err)
		assert.Nil(t, got, "cross-domain legacy token must NOT be adopted")

		// A non-adopted legacy entry must be left intact.
		legacyStored, err := store.Get(DefaultDeviceFlowTokenKey)
		require.NoError(t, err)
		assert.Equal(t, string(legacyData), legacyStored, "non-adopted legacy entry must not be deleted")
	})

	t.Run("Device_NoLegacyToken_NilReturned", func(t *testing.T) {
		store := newMemoryStore()

		dts := &deviceKeyringTokenSource{
			tokenKey:    namespacedTokenKey(DefaultDeviceFlowTokenKey, authBaseURL, clientID),
			legacyKey:   DefaultDeviceFlowTokenKey,
			authBaseURL: authBaseURL,
			store:       store,
		}

		got, err := dts.loadFromKeyring()
		require.NoError(t, err)
		assert.Nil(t, got, "no legacy token → nil, triggering fresh auth")
	})

	t.Run("Device_EmptyIssuer_NotAdopted", func(t *testing.T) {
		// A legacy token with no "iss" claim must not be adopted.
		legacy := &oauth2.Token{
			AccessToken: makeTestJWT(map[string]any{
				"eos": map[string]any{"realm": "realm-noiss", "org": "org-noiss"},
			}),
			TokenType: "Bearer",
			Expiry:    time.Now().Add(10 * time.Minute),
		}
		legacyData, _ := json.Marshal(legacy)

		store := newMemoryStore()
		_ = store.Set(DefaultDeviceFlowTokenKey, string(legacyData))

		dts := &deviceKeyringTokenSource{
			tokenKey:    namespacedTokenKey(DefaultDeviceFlowTokenKey, authBaseURL, clientID),
			legacyKey:   DefaultDeviceFlowTokenKey,
			authBaseURL: authBaseURL,
			store:       store,
		}

		got, err := dts.loadFromKeyring()
		require.NoError(t, err)
		assert.Nil(t, got, "token with empty iss must not be adopted")
	})
}
