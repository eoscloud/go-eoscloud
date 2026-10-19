package api

import (
	"fmt"
	"runtime"
	"strings"
)

// buildUserAgent returns the User-Agent value in the form
//
//	go-eoscloud/VERSION [name/version] (goX.Y; os/arch)
//
// app is the optional, already validated "name/version" part.
func buildUserAgent(app string) string {
	ua := "go-eoscloud/" + Version
	if app != "" {
		ua += " " + app
	}
	return fmt.Sprintf("%s (%s; %s/%s)", ua, runtime.Version(), runtime.GOOS, runtime.GOARCH)
}

// WithUserAgent identifies the application using the library. name and version
// are reported in the User-Agent in addition to the go-eoscloud version:
//
//	go-eoscloud/VERSION name/version (goX.Y; os/arch)
//
// Neither may be empty. The name must not contain spaces or "/", the version
// must not contain spaces. Only printable ASCII is allowed.
func WithUserAgent(name, version string) (Option, error) {
	if name == "" || version == "" {
		return nil, fmt.Errorf("user agent: name and version must not be empty")
	}
	if err := validateUserAgentToken("name", name, "/"); err != nil {
		return nil, err
	}
	if err := validateUserAgentToken("version", version, ""); err != nil {
		return nil, err
	}
	app := name + "/" + version
	return func(c *client) { c.userAgentApp = app }, nil
}

func validateUserAgentToken(field, value, forbidden string) error {
	for _, r := range value {
		if r <= ' ' || r >= 0x7f {
			return fmt.Errorf("user agent: %s %q must only contain printable ASCII without spaces", field, value)
		}
		if forbidden != "" && strings.ContainsRune(forbidden, r) {
			return fmt.Errorf("user agent: %s %q must not contain %q", field, value, forbidden)
		}
	}
	return nil
}

// UserAgent returns the User-Agent value a Caller created by NewClient sends,
// or "" if the Caller is not one created by NewClient. It is exposed so
// transports outside Call (e.g. websocket dials) can report the same value.
func UserAgent(c Caller) string {
	if cl, ok := c.(*client); ok {
		return cl.userAgent
	}
	return ""
}

// appendUserAgent appends ours to an already present User-Agent value.
func appendUserAgent(existing, ours string) string {
	if existing == "" {
		return ours
	}
	return existing + " " + ours
}
