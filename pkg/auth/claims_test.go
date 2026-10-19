package auth

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func makeTestJWT(payload any) string {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"RS256","typ":"JWT"}`))
	payloadBytes, _ := json.Marshal(payload)
	body := base64.RawURLEncoding.EncodeToString(payloadBytes)
	return header + "." + body + ".fakesig"
}

func TestExtractClaims(t *testing.T) {
	t.Run("ValidToken_ReturnsRealmAndOrg", func(t *testing.T) {
		token := &oauth2.Token{
			AccessToken: makeTestJWT(map[string]any{
				"sub": "user-1",
				"eos": map[string]any{
					"realm": "019c2e2b-324c-7dfb-81b3-32eded01af08",
					"org":   "019c2e2b-2d2c-7265-9652-bf0e9fc528a5",
				},
			}),
		}

		claims, err := ExtractClaims(token)
		require.NoError(t, err)
		require.NotNil(t, claims)
		assert.Equal(t, "019c2e2b-324c-7dfb-81b3-32eded01af08", claims.RealmID)
		assert.Equal(t, "019c2e2b-2d2c-7265-9652-bf0e9fc528a5", claims.OrganizationID)
	})

	t.Run("MissingOrgClaim_StillReturnsRealm", func(t *testing.T) {
		token := &oauth2.Token{
			AccessToken: makeTestJWT(map[string]any{
				"eos": map[string]any{
					"realm": "realm-only",
				},
			}),
		}

		claims, err := ExtractClaims(token)
		require.NoError(t, err)
		assert.Equal(t, "realm-only", claims.RealmID)
		assert.Equal(t, "", claims.OrganizationID)
	})

	t.Run("MissingEosClaim_ReturnsError", func(t *testing.T) {
		token := &oauth2.Token{
			AccessToken: makeTestJWT(map[string]any{
				"sub": "user-1",
			}),
		}

		claims, err := ExtractClaims(token)
		assert.Error(t, err)
		assert.Nil(t, claims)
		assert.Contains(t, err.Error(), "eos.realm")
	})

	t.Run("InvalidJWT_NotThreeParts_ReturnsError", func(t *testing.T) {
		token := &oauth2.Token{AccessToken: "notajwt"}

		claims, err := ExtractClaims(token)
		assert.Error(t, err)
		assert.Nil(t, claims)
		assert.Contains(t, err.Error(), "parsing JWT")
	})

	t.Run("InvalidBase64Payload_ReturnsError", func(t *testing.T) {
		token := &oauth2.Token{AccessToken: "header.!!!invalid!!!.sig"}

		claims, err := ExtractClaims(token)
		assert.Error(t, err)
		assert.Nil(t, claims)
	})

	t.Run("InvalidJSONPayload_ReturnsError", func(t *testing.T) {
		badPayload := base64.RawURLEncoding.EncodeToString([]byte(`not json`))
		token := &oauth2.Token{AccessToken: "header." + badPayload + ".sig"}

		claims, err := ExtractClaims(token)
		assert.Error(t, err)
		assert.Nil(t, claims)
	})
}
