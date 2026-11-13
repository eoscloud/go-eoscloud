package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
)

// DefaultKeyringServiceName is the default service name used to store tokens in the system keyring.
const DefaultKeyringServiceName = "go-eoscloud"

// DefaultAuthBaseURL is the default base URL used for the eoscloud auth service.
const DefaultAuthBaseURL = "https://auth.eoscloud.io"

// DefaultClientID is the default client ID used for the eoscloud device flow.
const DefaultClientID = "019cd72c-c183-7773-a5ef-bbfd960c0e89"

// DefaultDeviceFlowTokenKey is the fixed keyring key used to store device flow tokens.
const DefaultDeviceFlowTokenKey = "eosctl-device"

// DefaultClientCredentialsFlowTokenKey is the fixed keyring key used to store client credentials tokens.
const DefaultClientCredentialsFlowTokenKey = "eosctl-client-credentials"

// defaultConfigDirName is the name of the directory used for storing configuration files.
const defaultConfigDirName = "eosctl"

// defaultTokenStoreFileName is the name of the directory, below the config
// directory, that holds the per-service token files.
const defaultTokenStoreFileName = "tokens"

// namespacedTokenKey derives a stable, per-server token key by appending a
// short hex digest of "authBaseURL\x00clientID" to the legacy base key. This
// ensures tokens from different auth servers never share a keyring entry while
// preserving the human-readable prefix for debuggability.
//
// Example: "eosctl-device" + authBaseURL "https://auth.a.example" + clientID
// "abc" → "eosctl-device-<first 16 hex chars of SHA-256>".
func namespacedTokenKey(legacyBase, authBaseURL, clientID string) string {
	h := sha256.Sum256([]byte(authBaseURL + "\x00" + clientID))
	return legacyBase + "-" + hex.EncodeToString(h[:])[:16]
}

// DeviceFlowTokenKey returns the key under which the device flow token source
// stores the token for the given auth server and client ID. Tools that only
// read a cached login should use LoadDeviceFlowToken instead, which also
// falls back to the legacy key.
func DeviceFlowTokenKey(authBaseURL, clientID string) string {
	return namespacedTokenKey(DefaultDeviceFlowTokenKey, authBaseURL, clientID)
}

// DefaultTokenStoreDir returns the platform-appropriate directory for token
// storage: $XDG_CONFIG_HOME/eosctl/tokens (falling back to ~/.config) on
// XDG platforms, ~/Library/Application Support/eosctl/tokens on darwin and
// %AppData%\eosctl\tokens on Windows.
//
// A relative XDG_CONFIG_HOME is rejected rather than resolved against the
// current working directory, as the spec requires.
//
// When no user config directory can be determined at all ($HOME and
// XDG_CONFIG_HOME both unset) it returns the empty string. There is
// deliberately no fallback to a shared, predictable location such as the
// temp directory: the file-based token store reports an error in that case
// unless a directory is configured explicitly (see ErrNoTokenDir).
func DefaultTokenStoreDir() string {
	// os.UserConfigDir returns XDG_CONFIG_HOME verbatim on XDG platforms
	// without requiring it to be absolute; the spec says relative base
	// directories must be ignored, hence the explicit check.
	if dir, err := os.UserConfigDir(); err == nil && filepath.IsAbs(dir) {
		return filepath.Join(dir, defaultConfigDirName, defaultTokenStoreFileName)
	}

	// os.UserConfigDir only fails when $HOME (or %AppData% on Windows) is
	// unset. Per the XDG base directory spec XDG_CONFIG_HOME is authoritative
	// on its own and does not depend on $HOME, so honor it before giving up.
	if dir := os.Getenv("XDG_CONFIG_HOME"); filepath.IsAbs(dir) {
		return filepath.Join(dir, defaultConfigDirName, defaultTokenStoreFileName)
	}

	return ""
}
