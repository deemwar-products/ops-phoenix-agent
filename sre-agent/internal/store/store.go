package store

import (
	"errors"
	"fmt"
)

var (
	ErrKeyNotFound = errors.New("key not found in store")
	ErrStoreLocked = errors.New("store is locked")
)

// Store is the interface for persistent secret storage (OS keychain).
type Store interface {
	// Get retrieves a value by key.
	Get(key string) (string, error)
	// Put stores a value under a key.
	Put(key, value string) error
	// Delete removes a key.
	Delete(key string) error
	// Available reports whether the backing store is usable.
	Available() bool
}

// StoreFuncs adapts plain functions into a Store.
type StoreFuncs struct {
	GetFn    func(key string) (string, error)
	PutFn    func(key, value string) error
	DeleteFn func(key string) error
}

func (f *StoreFuncs) Get(key string) (string, error) {
	if f.GetFn == nil {
		return "", ErrStoreLocked
	}
	return f.GetFn(key)
}

func (f *StoreFuncs) Put(key, value string) error {
	if f.PutFn == nil {
		return ErrStoreLocked
	}
	return f.PutFn(key, value)
}

func (f *StoreFuncs) Delete(key string) error {
	if f.DeleteFn == nil {
		return ErrStoreLocked
	}
	return f.DeleteFn(key)
}

func (f *StoreFuncs) Available() bool {
	return f.GetFn != nil && f.PutFn != nil
}

// New creates a Store backed by the OS keychain.
// Falls back to a no-op store if the keychain is unavailable.
func New() (Store, error) {
	return NewKeychainStore()
}

func newUnavailableStore() Store {
	return &StoreFuncs{
		GetFn: func(key string) (string, error) { return "", ErrStoreLocked },
		PutFn: func(key, value string) error { return ErrStoreLocked },
		DeleteFn: func(key string) error { return ErrStoreLocked },
	}
}

// mustBeAvailable is a helper used in commands that require the store.
func mustBeAvailable(s Store) error {
	if !s.Available() {
		return fmt.Errorf("keychain is not available — check your OS keychain setup (%w)", ErrStoreLocked)
	}
	return nil
}
