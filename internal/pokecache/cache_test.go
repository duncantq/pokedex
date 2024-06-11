package pokecache

import (
	"fmt"
	"testing"
	"time"
)

func TestAddGet(t *testing.T) {
	cases := []struct {
		key   string
		value []byte
	}{
		{
			key:   "https://example.com",
			value: []byte("testdata"),
		},
		{
			key:   "https://example.com/path",
			value: []byte("moretestdata"),
		},
	}

	for i, c := range cases {
		t.Run(fmt.Sprintf("Test case %d", i), func(t *testing.T) {
			cache := NewCache(time.Second * 5)
			cache.Add(c.key, c.value)
			value, ok := cache.Get(c.key)
			if !ok {
				t.Errorf("expected to find key")
				return
			}
			if string(value) != string(value) {
				t.Errorf("expected to find value")
				return
			}
		})
	}
}

func TestReapLoop(t *testing.T) {
	cache := NewCache(time.Millisecond * 5)
	cache.Add("https://example.com", []byte("testdata"))

	if _, ok := cache.Get("https://example.com"); !ok {
		t.Errorf("expected to find key")
	}

	time.Sleep(time.Millisecond * 5)

	if _, ok := cache.Get("https://example.com"); ok {
		t.Errorf("expected to not find key")
	}
}
