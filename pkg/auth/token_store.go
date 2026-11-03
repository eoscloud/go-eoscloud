package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/zalando/go-keyring"
)

// TokenStore abstracts credential storage for a fixed service name.
type TokenStore interface {
	Get(key string) (string, error)
	Set(key, value string) error
	Delete(key string) error
	DeleteAll() error
}

// keyringStore delegates to the OS keyring.
type keyringStore struct {
	service string
}

func (k *keyringStore) Get(key string) (string, error) {
	return keyring.Get(k.service, key)
}

func (k *keyringStore) Set(key, value string) error {
	return keyring.Set(k.service, key, value)
}

func (k *keyringStore) Delete(key string) error {
	return keyring.Delete(k.service, key)
}

func (k *keyringStore) DeleteAll() error {
	return keyring.DeleteAll(k.service)
}

// ErrNoTokenDir is returned by the file-based token store when no token
// directory was configured and none can be derived from the environment
// ($HOME / XDG_CONFIG_HOME unset). There is intentionally no fallback to a
// shared location such as the temp directory.
var ErrNoTokenDir = errors.New("no token store directory available: set $HOME or $XDG_CONFIG_HOME, or configure a fallback token store directory explicitly")

// errCorruptTokenFile marks a token file whose content is not valid JSON.
var errCorruptTokenFile = errors.New("token file is corrupt")

// fileStore persists credentials as JSON files under a directory.
// Layout: <tokenDir>/<service>.json containing a map[key]value.
type fileStore struct {
	service  string
	tokenDir string
	mu       sync.Mutex
}

// newFileStore builds a fileStore, substituting the default token directory
// when tokenDir is empty. The default may itself be empty (no HOME), in which
// case every operation fails with ErrNoTokenDir instead of resolving against
// the current working directory.
func newFileStore(service, tokenDir string) *fileStore {
	if tokenDir == "" {
		tokenDir = DefaultTokenStoreDir()
	}

	return &fileStore{service: service, tokenDir: tokenDir}
}

func (f *fileStore) path() string {
	return filepath.Join(f.tokenDir, f.service+".json")
}

// checkDir verifies the token directory is a directory owned by the current
// user and not writable by group/others. When create is set a missing
// directory is created with mode 0700. It reports whether the directory exists.
func (f *fileStore) checkDir(create bool) (bool, error) {
	if f.tokenDir == "" {
		return false, ErrNoTokenDir
	}

	fi, err := os.Stat(f.tokenDir)
	if errors.Is(err, fs.ErrNotExist) {
		if !create {
			return false, nil
		}
		if err := os.MkdirAll(f.tokenDir, 0o700); err != nil {
			return false, err
		}
		fi, err = os.Stat(f.tokenDir)
	}
	if err != nil {
		return false, err
	}

	if !fi.IsDir() {
		return false, fmt.Errorf("token directory %s is not a directory", f.tokenDir)
	}
	if runtime.GOOS == "windows" {
		return true, nil
	}
	if !ownedByCurrentUser(fi) {
		return false, fmt.Errorf("token directory %s is not owned by the current user", f.tokenDir)
	}
	if fi.Mode().Perm()&0o022 != 0 {
		return false, fmt.Errorf("token directory %s must not be group- or world-writable (mode %04o)", f.tokenDir, fi.Mode().Perm())
	}

	return true, nil
}

// checkFile Lstat's the token file. A missing file is reported as exists=false.
// An existing file must be a regular file (no symlink), owned by the current
// user with mode 0600.
func (f *fileStore) checkFile() (bool, error) {
	fi, err := os.Lstat(f.path())
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	if !fi.Mode().IsRegular() {
		return false, fmt.Errorf("token file %s is not a regular file (symlinks are refused)", f.path())
	}
	if runtime.GOOS == "windows" {
		return true, nil
	}
	if !ownedByCurrentUser(fi) {
		return false, fmt.Errorf("token file %s is not owned by the current user", f.path())
	}
	if fi.Mode().Perm() != 0o600 {
		return false, fmt.Errorf("token file %s must have mode 0600, has %04o", f.path(), fi.Mode().Perm())
	}

	return true, nil
}

func (f *fileStore) load() (map[string]string, error) {
	dirExists, err := f.checkDir(false)
	if err != nil {
		return nil, err
	}
	if !dirExists {
		return map[string]string{}, nil
	}

	exists, err := f.checkFile()
	if err != nil {
		return nil, err
	}
	if !exists {
		return map[string]string{}, nil
	}

	data, err := os.ReadFile(f.path())
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return map[string]string{}, nil
		}
		return nil, err
	}

	var entries map[string]string
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("%w: %s: %w", errCorruptTokenFile, f.path(), err)
	}
	if entries == nil {
		entries = map[string]string{}
	}

	return entries, nil
}

// loadForWrite is like load but tolerates a corrupt file: it warns on stderr
// and starts from an empty set so the next save replaces the broken file.
func (f *fileStore) loadForWrite() (map[string]string, error) {
	entries, err := f.load()
	if errors.Is(err, errCorruptTokenFile) {
		_, _ = fmt.Fprintf(os.Stderr, "WARNING: %v; discarding its content\n", err)
		return map[string]string{}, nil
	}

	return entries, err
}

// save atomically replaces the token file: it writes a 0600 temp file in the
// same directory and renames it over the target.
func (f *fileStore) save(entries map[string]string) (err error) {
	if _, err := f.checkDir(true); err != nil {
		return err
	}
	if _, err := f.checkFile(); err != nil {
		return err
	}

	data, err := json.Marshal(entries)
	if err != nil {
		return err
	}

	tmp, err := os.CreateTemp(f.tokenDir, f.service+".*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer func() {
		if err != nil {
			_ = os.Remove(tmpName)
		}
	}()

	if cerr := tmp.Chmod(0o600); cerr != nil && runtime.GOOS != "windows" {
		_ = tmp.Close()
		return cerr
	}
	if _, err = tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err = tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err = tmp.Close(); err != nil {
		return err
	}

	return os.Rename(tmpName, f.path())
}

func (f *fileStore) Get(key string) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	entries, err := f.load()
	if errors.Is(err, errCorruptTokenFile) {
		_, _ = fmt.Fprintf(os.Stderr, "WARNING: %v; ignoring it\n", err)
		return "", ErrNotFound
	}
	if err != nil {
		return "", err
	}

	val, ok := entries[key]
	if !ok {
		return "", ErrNotFound
	}

	return val, nil
}

func (f *fileStore) Set(key, value string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	entries, err := f.loadForWrite()
	if err != nil {
		return err
	}

	entries[key] = value
	if err := f.save(entries); err != nil {
		return err
	}

	// Warn only once the token has actually been written to disk, so a refused
	// or failed write does not claim to have stored anything.
	_, _ = os.Stderr.Write([]byte("WARNING: storing token insecure in file. \n"))
	_, _ = os.Stderr.Write([]byte("Consider installing keyring software like gnome-keyring if possible. \n"))

	return nil
}

func (f *fileStore) Delete(key string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	entries, err := f.loadForWrite()
	if err != nil {
		return err
	}

	if _, ok := entries[key]; !ok {
		return nil
	}

	delete(entries, key)
	return f.save(entries)
}

func (f *fileStore) DeleteAll() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.tokenDir == "" {
		return ErrNoTokenDir
	}

	return os.Remove(f.path())
}

// ErrNotFound is returned when a key is not found in the file store.
var ErrNotFound = keyring.ErrNotFound

// NewTokenStore returns a TokenStore bound to the given service name, backed by
// the OS keyring if available, otherwise falling back to file-based storage under
// tokenDir. An empty tokenDir means "use DefaultTokenStoreDir".
func NewTokenStore(service, tokenDir string) TokenStore {
	if err := keyring.Set(DefaultKeyringServiceName+"-probe", "_probe", "ok"); err == nil {
		_ = keyring.Delete(DefaultKeyringServiceName+"-probe", "_probe")
		return &keyringStore{service: service}
	}

	return newFileStore(service, tokenDir)
}
