package auth

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDefaultTokenStoreDir verifies XDG base directory conformance: the token
// directory lives under the user config directory in a non-hidden,
// application-named subdirectory, an absolute XDG_CONFIG_HOME is honored, and
// no unsafe fallback location is invented when nothing is resolvable.
func TestDefaultTokenStoreDir(t *testing.T) {
	t.Run("SubdirIsNotHidden", func(t *testing.T) {
		dir := DefaultTokenStoreDir()
		require.NotEmpty(t, dir, "test environment should have a user config dir")

		assert.Equal(t, filepath.Join("eosctl", "tokens"),
			filepath.Join(filepath.Base(filepath.Dir(dir)), filepath.Base(dir)),
			"token dir should end in eosctl/tokens")
		assert.NotContains(t, dir, ".eosctl",
			"app subdir must not be dot-prefixed inside an already-hidden config dir")
	})

	t.Run("UnderUserConfigDir", func(t *testing.T) {
		configDir, err := os.UserConfigDir()
		require.NoError(t, err)

		assert.Equal(t, filepath.Join(configDir, "eosctl", "tokens"), DefaultTokenStoreDir())
	})

	t.Run("HonorsAbsoluteXDGConfigHome", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("XDG_CONFIG_HOME is not the platform convention on windows")
		}

		xdg := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", xdg)
		// Unset HOME so darwin also falls through to the explicit XDG branch;
		// on linux os.UserConfigDir already prefers XDG_CONFIG_HOME.
		t.Setenv("HOME", "")

		assert.Equal(t, filepath.Join(xdg, "eosctl", "tokens"), DefaultTokenStoreDir())
	})

	t.Run("NoConfigDir_FallsBackToTempDir", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("config dir derives from %AppData% on windows")
		}

		t.Setenv("HOME", "")
		t.Setenv("XDG_CONFIG_HOME", "")

		assert.Equal(t, filepath.Join(os.TempDir(), "eosctl", "tokens"), DefaultTokenStoreDir(),
			"file-based storage should stay available when no config dir is resolvable")
	})

	t.Run("RelativeXDGConfigHome_IsIgnored", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("XDG_CONFIG_HOME is not the platform convention on windows")
		}

		t.Setenv("HOME", "")
		t.Setenv("XDG_CONFIG_HOME", "relative/config")

		dir := DefaultTokenStoreDir()
		assert.True(t, filepath.IsAbs(dir), "must never return a cwd-relative path, got %q", dir)
		assert.NotContains(t, dir, "relative/config",
			"relative XDG base dirs must be ignored, not resolved against the cwd")
		assert.Equal(t, filepath.Join(os.TempDir(), "eosctl", "tokens"), dir)
	})
}

// TestNewFileStoreEmptyDir asserts an empty tokenDir resolves to the default
// token directory rather than to a bare filename, which would put credentials
// in the process's current working directory. Callers can reach this via
// With{DeviceFlow,ClientCredentials}FallbackTokenStoreDir("") or by calling
// NewTokenStore directly.
func TestNewFileStoreEmptyDir(t *testing.T) {
	t.Run("ResolvesToDefaultDir", func(t *testing.T) {
		store := newFileStore("eosctl-device", "")

		assert.Equal(t, DefaultTokenStoreDir(), store.tokenDir)
		assert.True(t, filepath.IsAbs(store.path()),
			"token path must be absolute, got %q", store.path())
	})

	t.Run("WritesNothingToCwd", func(t *testing.T) {
		// Point the default dir at a temp HOME so the write cannot land in the
		// real user config directory. On darwin os.UserConfigDir derives from
		// $HOME; on XDG platforms XDG_CONFIG_HOME takes precedence.
		tmp := t.TempDir()
		t.Setenv("HOME", tmp)
		t.Setenv("XDG_CONFIG_HOME", tmp)

		cwd, err := os.Getwd()
		require.NoError(t, err)

		before, err := os.ReadDir(cwd)
		require.NoError(t, err)

		store := newFileStore("eosctl-device", "")
		require.NoError(t, store.Set("k", "v"))

		after, err := os.ReadDir(cwd)
		require.NoError(t, err)
		assert.Len(t, after, len(before), "no token file should be created in the cwd")

		assert.True(t, strings.HasPrefix(store.path(), tmp),
			"token should be written under the redirected config dir, got %q", store.path())
		_, err = os.Stat(store.path())
		assert.NoError(t, err, "token file should exist at the resolved default path")
	})
}

// TestFileStoreRoundTrip covers the happy path against a real directory,
// including the 0600/0700 permissions expected for credential storage.
func TestFileStoreRoundTrip(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "eosctl", "tokens")
	store := &fileStore{service: "eosctl-device", tokenDir: dir}

	require.NoError(t, store.Set("key-a", "token-a"))

	val, err := store.Get("key-a")
	require.NoError(t, err)
	assert.Equal(t, "token-a", val)

	dirInfo, err := os.Stat(dir)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0700), dirInfo.Mode().Perm(), "token dir must not be group/world readable")

	path := store.path()

	fileInfo, err := os.Stat(path)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0600), fileInfo.Mode().Perm(), "token file must not be group/world readable")

	require.NoError(t, store.Delete("key-a"))
	_, err = store.Get("key-a")
	assert.True(t, errors.Is(err, ErrNotFound), "got %v", err)

	require.NoError(t, store.DeleteAll())
	_, err = os.Stat(path)
	assert.True(t, os.IsNotExist(err))
}
