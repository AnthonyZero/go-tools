package main

import (
	"flag"
	"fmt"
	"go-tools/cache"
	"log"
	"net/http"
)

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func createGroup() *cache.Group {
	return cache.NewGroup("scores", 2<<10, cache.GetterFunc(func(key string) ([]byte, error) {
		log.Println("[DB] search key", key)
		if v, ok := db[key]; ok {
			return []byte(v), nil
		}
		return nil, fmt.Errorf("%s not exist", key)
	}))
}

func startCacheServer(addr string) {
	peers := cache.NewHTTPPool(addr)
	log.Println("gocache is running at", addr)
	log.Fatal(http.ListenAndServe(addr[7:], peers))
}

func startAPIServer(apiAddr string, addrs []string, manager *cache.Group) {
	peers := cache.NewHTTPPool(apiAddr)
	peers.Set(addrs...)
	manager.RegisterPeers(peers)
	http.Handle("/api", http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		key := request.URL.Query().Get("key")
		view, err := manager.Get(key)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		writer.Header().Set("Content-Type", "application/octet-stream")
		writer.Write(view.ByteSlice())
	}))
	log.Println("server is running at", apiAddr)
	log.Fatal(http.ListenAndServe(apiAddr[7:], nil))
}

func main() {

	var port int
	var api bool
	flag.IntVar(&port, "port", 8001, "Gocache server port")
	flag.BoolVar(&api, "api", false, "Start a api server?")
	flag.Parse()

	apiAddr := "http://localhost:9999" //api服务 只负责用户交换
	addrMap := map[int]string{         //3个缓存远程节点
		8001: "http://localhost:8001",
		8002: "http://localhost:8002",
		8003: "http://localhost:8003",
	}

	var addrs []string
	for _, v := range addrMap {
		addrs = append(addrs, v)
	}

	manager := createGroup()
	if api {
		startAPIServer(apiAddr, addrs, manager)
	} else {
		startCacheServer(addrMap[port])
	}
}
