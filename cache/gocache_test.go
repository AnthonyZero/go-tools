package cache

import (
	"fmt"
	"log"
	"reflect"
	"testing"
)

func TestGetter(t *testing.T) {
	var f Getter = GetterFunc(func(key string) ([]byte, error) {
		return []byte(key), nil
	})
	expect := []byte("key")

	bytes, _ := f.Get("key")
	if reflect.DeepEqual(bytes, expect) {
		t.Logf("callback ok")
	}
}

func TestGet(t *testing.T) {
	//假DB数据
	var db = map[string]string{
		"Tom":  "630",
		"Jack": "589",
		"Sam":  "567",
	}

	loadCounts := make(map[string]int, len(db)) //每个key 的load次数

	gee := NewGroup("scores", 2<<10, GetterFunc(func(key string) ([]byte, error) {
		log.Printf("[DB] search key = %s \n", key)
		if v, ok := db[key]; ok {
			if _, ok := loadCounts[key]; !ok {
				loadCounts[key] = 1
			} else {
				loadCounts[key] += 1
			}
			return []byte(v), nil //返回
		}
		return nil, fmt.Errorf("%s not exist", key)
	}))

	for k, v := range db {
		if view, err := gee.Get(k); err != nil || view.String() != v {
			t.Fatal("failed to get value of Cache")
		} // load from callback function
		if _, err := gee.Get(k); err != nil || loadCounts[k] > 1 { //每个key 都只会回调数据源 加载一次，然后后续每次都命中cache
			t.Fatalf("cache %s miss", k)
		} // cache hit
	}

	if _, err := gee.Get("Jack"); err != nil {

	}

	//数据源中无unknown key这个数据 每次都回query 但无结果
	if view, err := gee.Get("unknown"); err == nil {
		t.Fatalf("the value of unknow should be empty, but %s got", view)
	}
	if view, err := gee.Get("unknown"); err == nil {
		t.Fatalf("the value of unknow should be empty, but %s got", view)
	}

	for s, i := range loadCounts {
		t.Logf("key = %s, count = %d", s, i)
	}
}
