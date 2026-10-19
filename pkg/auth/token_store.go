package auth

import (
	"encoding/json"
	"os"
	"path/filepath"
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

// fileStore persists credentials as JSON files under a directory.
// Layout: <tokenDir>/<service>.json containing a map[key]value.
type fileStore struct {
	service  string
	tokenDir string
	mu       sync.Mutex
}

// newFileStore builds a fileStore, substituting the default token directory
// when tokenDir is empty. This keeps tokenDir absolute, so path never produces
// a bare filename that would resolve against the current working directory.
func newFileStore(service, tokenDir string) *fileStore {
	if tokenDir == "" {
		tokenDir = DefaultTokenStoreDir()
	}

	return &fileStore{service: service, tokenDir: tokenDir}
}

func (f *fileStore) path() string {
	return filepath.Join(f.tokenDir, f.service+".json")
}

func (f *fileStore) load() (map[string]string, error) {
	data, err := os.ReadFile(f.path())
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]string{}, nil
		}
		return nil, err
	}

	var entries map[string]string
	if err := json.Unmarshal(data, &entries); err != nil {
		return map[string]string{}, nil
	}

	return entries, nil
}

func (f *fileStore) save(entries map[string]string) error {
	if err := os.MkdirAll(f.tokenDir, 0700); err != nil {
		return err
	}

	data, err := json.Marshal(entries)
	if err != nil {
		return err
	}

	return os.WriteFile(f.path(), data, 0600)
}

func (f *fileStore) Get(key string) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	entries, err := f.load()
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

	entries, err := f.load()
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

	entries, err := f.load()
	if err != nil {
		return err
	}

	delete(entries, key)
	return f.save(entries)
}

func (f *fileStore) DeleteAll() error {
	f.mu.Lock()
	defer f.mu.Unlock()

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
