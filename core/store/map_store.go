package store

import "fmt"

type mapStore struct {
	store map[string]string
}

// A simple store that is based out of golang maps
func NewMapStore() Store {
	return &mapStore{
		store: make(map[string]string),
	}
}

func (ms *mapStore) Get(k string) (string, error) {
	v, ok := ms.store[k]
	if !ok {
		return "", fmt.Errorf("key not found")
	}
	return v, nil
}

func (ms *mapStore) Set(k string, v string) error {
	ms.store[k] = v
	return nil
}

func (ms *mapStore) Delete(k string) error {
	delete(ms.store, k)
	return nil
}
