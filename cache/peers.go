package cache

// PeerPicker 用于根据传入的 key 选择相应节点 PeerGetter
type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool)
}

// PeerGetter 模拟HTTP 客户端
type PeerGetter interface {
	Get(group string, key string) ([]byte, error)
}
