package cache

import (
	"fmt"
	"go-tools/consistenthash"
	pb "go-tools/gocachepb"
	"google.golang.org/protobuf/proto"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

const (
	defaultBasePath = "/_gocache/"
	defaultReplicas = 50
)

var _ PeerGetter = (*httpGetter)(nil)
var _ PeerPicker = (*HTTPPool)(nil)

type HTTPPool struct {
	self        string
	basePath    string
	mu          sync.Mutex
	peers       *consistenthash.Map    //一致性哈希算法的 Map
	httpGetters map[string]*httpGetter //每一个远程节点对应一个 httpGetter
}

type httpGetter struct {
	baseUrl string //baseURL 表示将要访问的远程节点的地址
}

func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self:     self,
		basePath: defaultBasePath,
	}
}

func (p *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[Service %s] %s", p.self, fmt.Sprintf(format, v...))
}

// HTTP服务的能力
func (p *HTTPPool) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	if !strings.HasPrefix(request.URL.Path, p.basePath) {
		panic("HTTPPool serving unexpected path: " + request.URL.Path)
	}
	p.Log("%s %s %s", "Receive Request:", request.Method, request.URL.Path)
	// /<basepath>/<groupname>/<key> required
	parts := strings.SplitN(request.URL.Path[len(p.basePath):], "/", 2)
	if len(parts) != 2 {
		http.Error(writer, "bad request", http.StatusBadRequest)
		return
	}

	groupName := parts[0]
	key := parts[1]

	group := GetGroup(groupName)
	if group == nil {
		http.Error(writer, "no such group : "+groupName, http.StatusNotFound)
		return
	}

	view, err := group.Get(key)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	//writer.Header().Set("Content-Type", "application/octet-stream")
	//writer.Write(view.ByteSlice())
	// Write the value to the response body as a proto message.
	body, err := proto.Marshal(&pb.Response{Value: view.ByteSlice()}) //ServeHTTP() 中使用 proto.Marshal() 编码 HTTP 响应。
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/octet-stream")
	writer.Write(body)
}

func (p *HTTPPool) Set(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.peers = consistenthash.New(defaultReplicas, nil)
	p.peers.Add(peers...) //节点初始化
	p.httpGetters = make(map[string]*httpGetter, len(peers))
	for _, peer := range peers {
		p.httpGetters[peer] = &httpGetter{
			baseUrl: peer + p.basePath,
		}
	}
}

func (p *HTTPPool) PickPeer(key string) (PeerGetter, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if peer := p.peers.Get(key); peer != "" && peer != p.self {
		p.Log("Ready to Pick peer %s", peer)
		return p.httpGetters[peer], true
	}
	return nil, false
}

//func (h *httpGetter) Get(group string, key string) ([]byte, error) {
//	u := fmt.Sprintf("%v%v/%v", h.baseUrl, url.QueryEscape(group), url.QueryEscape(key))
//
//	res, err := http.Get(u)
//	if err != nil {
//		return nil, err
//	}
//	defer res.Body.Close()
//
//	if res.StatusCode != http.StatusOK {
//		return nil, fmt.Errorf("server returned : %v", res.StatusCode)
//	}
//
//	bytes, err := ioutil.ReadAll(res.Body)
//	if err != nil {
//		return nil, fmt.Errorf("reading response body: %v", err)
//	}
//	log.Printf("HTTP Get response data = %v \n", string(bytes))
//	return bytes, nil
//}

//Get 将 HTTP 通信的中间载体替换成了 protobuf
func (h *httpGetter) Get(in *pb.Request, out *pb.Response) error {
	u := fmt.Sprintf(
		"%v%v/%v",
		h.baseUrl,
		url.QueryEscape(in.GetGroup()),
		url.QueryEscape(in.GetKey()),
	)
	res, err := http.Get(u)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned: %v", res.Status)
	}

	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %v", err)
	}

	if err = proto.Unmarshal(bytes, out); err != nil { //Get() 中使用 proto.Unmarshal() 解码 HTTP 响应
		return fmt.Errorf("decoding response body: %v", err)
	}
	log.Printf("HTTP Get response data = %v \n", string(bytes))
	return nil
}
