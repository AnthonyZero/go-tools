package consistenthash

import (
	"strconv"
	"testing"
)

func TestHashing(t *testing.T) {
	hash := New(3, func(data []byte) uint32 {
		i, _ := strconv.Atoi(string(data))
		return uint32(i)
	})

	// 2, 4, 6, 12, 14, 16, 22, 24, 26  hash测试数据  6：6 16 26   4： 4 14 24  2：2 12 22 虚拟节点hash与真实节点key
	hash.Add("6", "4", "2")

	testCases := map[string]string{
		"2":  "2",
		"11": "2",
		"23": "4",
		"27": "2",
	}

	for k, v := range testCases {
		if hash.Get(k) == v {
			t.Logf("Asking for %s, should have yielded %s", k, v)
		}
	}

	// Adds 8, 18, 28
	hash.Add("8")

	// 27 should now map to 8.
	testCases["27"] = "8" //27之前是2 变成了hash到8了

	for k, v := range testCases {
		if hash.Get(k) == v {
			t.Logf("Asking for %s, should have yielded %s", k, v)
		}
	}
}
