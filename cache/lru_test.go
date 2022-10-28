package lru

import (
	"reflect"
	"testing"
)

type String string

func (s String) Len() int {
	return len(s)
}

func TestGet(t *testing.T) {
	lru := New(0, nil) //初始化Cache
	lru.Add("key1", String("1234"))
	if v, ok := lru.Get("key1"); ok {
		t.Logf("cache hit key1 = %s ok", string(v.(String)))
	}
	if _, ok := lru.Get("key2"); !ok {
		t.Logf("cache miss key2 failed")
	}
}

func TestRemoveOldest(t *testing.T) {
	k1, k2, k3 := "key1", "key2", "k3"
	v1, v2, v3 := "value1", "value2", "v3"
	cap := len(k1 + k2 + v1 + v2) //4 + 4 + 6 + 6 = 20 bytes
	t.Logf("cap = %v", cap)
	lru := New(int64(cap), nil)
	lru.Add(k1, String(v1))
	lru.Add(k2, String(v2))
	lru.Add(k3, String(v3)) //k1淘汰

	if _, ok := lru.Get("key1"); !ok {
		t.Logf("Removeoldest key1 occur, cur len = %d， cur bytes = %v, cur data = %v", lru.Len(), lru.nbytes, lru.cache)
	}
}

func TestOnEvicted(t *testing.T) {
	keys := make([]string, 0)
	callback := func(key string, value Value) {
		t.Logf("keys has append key = %s", key)
		keys = append(keys, key)
	}
	lru := New(10, callback)
	lru.Add("key1", String("123456"))
	lru.Add("k2", String("k2"))
	lru.Add("k3", String("k3"))
	lru.Add("k4", String("k4"))
	expect := []string{"key1", "k2"}

	if reflect.DeepEqual(expect, keys) {
		t.Logf("Call OnEvicted ok, expect key equals to %s", keys)
	}
}
