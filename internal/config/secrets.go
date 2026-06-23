package config

import (
	"errors"
	"fmt"

	"github.com/zalando/go-keyring"
)

const keyringService = "notion-cli"

// ErrSecretNotFound is returned when a stored credential is missing from the
// underlying secret store.
var ErrSecretNotFound = errors.New("secret not found")

// SecretStore abstracts OS credential storage so tests can use an in-memory
// store while production uses the system keychain.
type SecretStore interface {
	Get(key string) (string, error)
	Set(key, value string) error
	Delete(key string) error
}

type keyringSecretStore struct{}

func (keyringSecretStore) Get(key string) (string, error) {
	value, err := keyring.Get(keyringService, key)
	if err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return "", ErrSecretNotFound
		}
		return "", err
	}
	return value, nil
}

func (keyringSecretStore) Set(key, value string) error {
	return keyring.Set(keyringService, key, value)
}

func (keyringSecretStore) Delete(key string) error {
	if err := keyring.Delete(keyringService, key); err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return nil
		}
		return err
	}
	return nil
}

var secretStore SecretStore = keyringSecretStore{}

// SetSecretStoreForTest swaps the secret store and returns a restore function.
// It is intended for tests and deliberately lives in production code so command
// tests in other packages can avoid touching the user's real keychain.
func SetSecretStoreForTest(store SecretStore) func() {
	prev := secretStore
	secretStore = store
	return func() { secretStore = prev }
}

// MemorySecretStore is a simple in-memory SecretStore for tests.
type MemorySecretStore struct {
	Values map[string]string
}

func NewMemorySecretStore() *MemorySecretStore {
	return &MemorySecretStore{Values: map[string]string{}}
}

func (m *MemorySecretStore) Get(key string) (string, error) {
	value, ok := m.Values[key]
	if !ok {
		return "", ErrSecretNotFound
	}
	return value, nil
}

func (m *MemorySecretStore) Set(key, value string) error {
	if m.Values == nil {
		m.Values = map[string]string{}
	}
	m.Values[key] = value
	return nil
}

func (m *MemorySecretStore) Delete(key string) error {
	delete(m.Values, key)
	return nil
}

func secretKey(slug, name string) string {
	return fmt.Sprintf("workspace:%s:%s", slug, name)
}
