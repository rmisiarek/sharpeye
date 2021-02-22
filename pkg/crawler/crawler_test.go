package crawler

import "testing"

func TestUrlCache(t *testing.T) {
	cache := newURLCache()

	_key := "test"
	exist := cache.Set(_key)
	if !exist {
		t.Errorf("set %s to cache first time; got %v, want %v", _key, exist, false)
	}

	exist = cache.Set(_key)
	if exist {
		t.Errorf("set %s to cache second time; got %v, want %v", _key, exist, true)
	}
}
