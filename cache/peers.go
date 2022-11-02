package cache

import pb "go-tools/gocachepb"

// PeerPicker 用于根据传入的 key 选择相应节点 PeerGetter
type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool)
}

// PeerGetter 模拟HTTP 客户端
type PeerGetter interface {
	//Get(group string, key string) ([]byte, error)
	Get(in *pb.Request, out *pb.Response) error //还是HTTP请求 只不过换了数据结构 用pb格式
}
