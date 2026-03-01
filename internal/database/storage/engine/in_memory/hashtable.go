package inmemory

import "sync"

type Hashtable interface {
	Set(string, string)
	Get(string) (string, bool)
	Del(string)
}

type hashtable struct {
	data map[string]string
	mu   sync.RWMutex
}

func NewHashtable() Hashtable {
	data := make(map[string]string)
	return &hashtable{data: data, mu: sync.RWMutex{}}
}

func (ht *hashtable) Set(key, value string) {
	ht.mu.Lock()
	defer ht.mu.Unlock()
	ht.data[key] = value
}

func (ht *hashtable) Get(key string) (string, bool) {
	ht.mu.RLock()
	defer ht.mu.RUnlock()
	value, found := ht.data[key]
	return value, found
}

func (ht *hashtable) Del(key string) {
	ht.mu.Lock()
	defer ht.mu.Unlock()
	delete(ht.data, key)
}
