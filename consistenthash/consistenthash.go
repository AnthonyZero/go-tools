package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

type Hash func(data []byte) uint32

type Map struct {
	hash     Hash           //Hash 函数 hash
	replicas int            //虚拟节点倍数
	keys     []int          //sorted //哈希环(数组) 记录虚拟节点本身的hash值
	hashMap  map[int]string //虚拟节点与真实节点的映射表 hashMap 键是虚拟节点的哈希值，值是真实节点的名称。
}

func New(replicas int, fn Hash) *Map {
	m := &Map{
		replicas: replicas,
		hash:     fn,
		hashMap:  make(map[int]string),
	}
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

// Add 添加真实节点/机器
func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ {
			//虚拟节点的哈希值
			hash := int(m.hash([]byte(strconv.Itoa(i) + key))) //对每一个真实节点 key，对应创建 m.replicas 个虚拟节点
			m.keys = append(m.keys, hash)                      //添加到环上
			m.hashMap[hash] = key
		}
	}
	sort.Ints(m.keys)
}

func (m *Map) Get(key string) string {
	if len(m.keys) == 0 {
		return ""
	}
	hash := int(m.hash([]byte(key)))                   //key的哈希值
	idx := sort.Search(len(m.keys), func(i int) bool { //顺时针从小到大 找到第一个匹配的虚拟节点的下标 idx
		return m.keys[i] >= hash //当hash比m.keys都大的时候 idx = len(m.keys)
	})

	return m.hashMap[m.keys[idx%len(m.keys)]] //idx = len(m.keys) 的时候 取环第一个节点
}
