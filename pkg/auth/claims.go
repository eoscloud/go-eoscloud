package auth

import (
	"fmt"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
)

// TokenClaims holds the realm and organization IDs extracted from a JWT token.
type TokenClaims struct {
	RealmID        string
	OrganizationID string
}

// TokenSourceWithClaims extends oauth2.TokenSource with direct access to the EOS
// token claims. Implementations extract and cache claims inside Token() so that
// TokenClaims() never needs to re-parse the JWT.
type TokenSourceWithClaims interface {
	oauth2.TokenSource
	TokenClaims() (*TokenClaims, error)
	ClearTokenStore() error
}

type eosTokenClaims struct {
	EOS struct {
		Realm string `json:"realm"`
		Org   string `json:"org"`
	} `json:"eos"`
	jwt.RegisteredClaims
}

// tokenIssuer extracts the "iss" claim from a JWT without signature
// verification. It is intentionally permissive — unlike ExtractClaims it does
// not require the "eos" claim — so it can be used on legacy tokens whose
// payload may only have standard JWT fields.
func tokenIssuer(token *oauth2.Token) (string, error) {
	var claims jwt.RegisteredClaims
	if _, _, err := jwt.NewParser().ParseUnverified(token.AccessToken, &claims); err != nil {
		return "", fmt.Errorf("parsing JWT: %w", err)
	}
	return claims.Issuer, nil
}

// issuerMatchesAuthBaseURL reports whether the token issuer identifies the
// same auth server as authBaseURL. Both strings are trimmed of trailing slashes
// and compared case-insensitively so minor URL variations do not cause false
// mismatches. An empty issuer never matches.
func issuerMatchesAuthBaseURL(issuer, authBaseURL string) bool {
	if issuer == "" || authBaseURL == "" {
		return false
	}
	return strings.EqualFold(
		strings.TrimRight(issuer, "/"),
		strings.TrimRight(authBaseURL, "/"),
	)
}

// ExtractClaims parses the JWT access token without signature verification
// (the token is already validated by the OAuth2 flow) and extracts realm and
// organization IDs from the "eos" claim.
func ExtractClaims(token *oauth2.Token) (*TokenClaims, error) {
	var claims eosTokenClaims
	if _, _, err := jwt.NewParser().ParseUnverified(token.AccessToken, &claims); err != nil {
		return nil, fmt.Errorf("parsing JWT: %w", err)
	}

	if claims.EOS.Realm == "" && claims.EOS.Org == "" {
		return nil, fmt.Errorf("JWT token missing eos.realm or eos.org claim")
	}

	return &TokenClaims{
		RealmID:        claims.EOS.Realm,
		OrganizationID: claims.EOS.Org,
	}, nil
}
