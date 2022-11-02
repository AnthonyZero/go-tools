package cache

import (
	"fmt"
	pb "go-tools/gocachepb"
	"go-tools/lru"
	"go-tools/singleflight"
	"log"
	"sync"
)

type Getter interface {
	Get(key string) ([]byte, error)
}

type GetterFunc func(key string) ([]byte, error) //接口型函数

func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

//负责与用户的交互，并且控制缓存值存储和获取的流程
type Group struct {
	name      string
	getter    Getter //缓存未命中时获取源数据的回调
	mainCache cache
	peers     PeerPicker
	loader    *singleflight.Group //fetch once
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil Getter")
	}
	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:   name,
		getter: getter,
		mainCache: cache{
			cacheBytes: cacheBytes,
			lru:        lru.New(cacheBytes, nil),
			mu:         sync.Mutex{},
		},
		loader: &singleflight.Group{},
	}
	groups[name] = g
	return g
}

func GetGroup(name string) *Group {
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}

	//从 mainCache 中查找缓存，如果存在则返回缓存值。
	if v, ok := g.mainCache.get(key); ok {
		log.Printf("gocache hit key = %s \n", key)
		return v, nil
	}
	return g.load(key)
}

// 使用 PickPeer() 方法选择节点，若非本机节点，则调用 getFromPeer() 从远程获取。若是本机节点或失败，则回退到 getLocally()
func (g *Group) load(key string) (value ByteView, err error) {
	viewi, err := g.loader.Do(key, func() (interface{}, error) { //将原来的 load 的逻辑，使用 g.loader.Do 包裹起来即可，这样确保了并发场景下针对相同的 key，load 过程只会调用一次。
		if g.peers != nil {
			log.Printf("consistent hash choose\n")
			if peer, ok := g.peers.PickPeer(key); ok {
				if value, err := g.getFromPeer(peer, key); err == nil {
					log.Printf("[gocache] success get value from peer, %v \n", value)
					return value, nil
				}
				log.Printf("[gocache] Failed to get from peer, %s %v \n", key, err)
			}
		}
		return g.getLocally(key)
	})
	if err == nil {
		return viewi.(ByteView), err //没有err 强转类型 返回数据
	}
	return ByteView{}, nil
}

//获取源数据，并且将源数据添加到缓存 mainCache 中（通过 populateCache 方法）
func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		log.Printf("[gocache] Get Local data fail, the err: %v \n", err)
		return ByteView{}, err

	}
	value := ByteView{b: cloneBytes(bytes)}
	log.Printf("[gocache] Get Local data ok, put value to Cache, %v \n", value.String())
	g.populateCache(key, value)
	return value, nil
}

func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}

func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peers
}

func (g *Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	//bytes, err := peer.Get(g.name, key)
	req := &pb.Request{
		Group: g.name,
		Key:   key,
	}
	res := &pb.Response{}
	err := peer.Get(req, res)

	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: res.Value}, nil
}
