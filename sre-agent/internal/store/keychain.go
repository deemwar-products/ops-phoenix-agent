package store

import (
	"fmt"
	"os/exec"
	"strings"
)

// keychainStore uses macOS `security` CLI for keychain access.
// Falls back to no-op store if `security` is unavailable.
type keychainStore struct{}

func NewKeychainStore() (Store, error) {
	// Verify `security` exists
	if _, err := exec.LookPath("security"); err != nil {
		return &noopStore{}, nil
	}
	return &keychainStore{}, nil
}

func (k *keychainStore) Get(key string) (string, error) {
	cmd := exec.Command("security", "find-generic-password",
		"-s", key,
		"-a", "sre-agent",
		"-w")
	out, err := cmd.Output()
	if err != nil {
		if strings.Contains(err.Error(), "errSecItemNotFound") {
			return "", ErrKeyNotFound
		}
		return "", fmt.Errorf("keychain read: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func (k *keychainStore) Put(key, value string) error {
	// -U = update if exists, add if not
	cmd := exec.Command("security", "add-generic-password",
		"-s", key,
		"-a", "sre-agent",
		"-w", value,
		"-U",
		"-T", "",
	)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("keychain write: %w", err)
	}
	return nil
}

func (k *keychainStore) Delete(key string) error {
	cmd := exec.Command("security", "delete-generic-password",
		"-s", key,
		"-a", "sre-agent",
	)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("keychain delete: %w", err)
	}
	return nil
}

func (k *keychainStore) Available() bool {
	return true
}

// noopStore is used when keychain is unavailable (e.g., CI, Linux without secret service).
type noopStore struct{}

func (n *noopStore) Get(key string) (string, error) {
	return "", ErrKeyNotFound
}

func (n *noopStore) Put(key, value string) error {
	return ErrStoreLocked
}

func (n *noopStore) Delete(key string) error {
	return ErrStoreLocked
}

func (n *noopStore) Available() bool {
	return false
}
