package lru

import "testing"

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
