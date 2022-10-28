package lru

import "container/list"

type Cache struct {
	ll    *list.List
	cache map[string]*list.Element
	//允许使用的最大内存
	maxBytes int64
	//当前已使用的内存
	nbytes int64
	//是某条记录被移除时的回调函数，可以为 nil
	OnEvicted func(key string, value Value)
}

// Value use Len to count how many bytes it takes
type Value interface {
	Len() int
}

//双向链表节点的数据类型
type entry struct {
	key   string
	value Value
}

func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

func (c *Cache) Add(key string, value Value) {
	if ele, ok := c.cache[key]; ok {
		//如果已经存在
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)                               //原先的value
		c.nbytes += int64(value.Len()) - int64(kv.value.Len()) //换了Value Key未变
		kv.value = value
	} else {
		//如果不存在 这里重点 Element 的Value就是entry
		ele := c.ll.PushFront(&entry{
			key:   key,
			value: value,
		})
		c.cache[key] = ele
		c.nbytes += int64(len(key)) + int64(value.Len()) //已使用内容 key value
	}
	for c.maxBytes != 0 && c.maxBytes < c.nbytes {
		c.RemoveOldest() //淘汰
	}
}

func (c *Cache) Get(key string) (Value, bool) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele) //约定front为最新最近访问的
		kv := ele.Value.(*entry)
		return kv.value, ok
	}
	return nil, false
}

func (c *Cache) RemoveOldest() {
	ele := c.ll.Back()
	if ele != nil {
		c.ll.Remove(ele)
		kv := ele.Value.(*entry)
		delete(c.cache, kv.key)
		c.nbytes -= int64(len(kv.key)) + int64(kv.value.Len())
		if c.OnEvicted != nil { //回调函数如果有的话 触发
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

func (c *Cache) Len() int {
	return c.ll.Len()
}
