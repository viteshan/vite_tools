package core

import (
	"fmt"
	"sync"

	"github.com/vitelabs/go-vite/client"
)

var globalCli client.RpcClient
var globalUrl string

var mu sync.Mutex

func GetCli(url string) client.RpcClient {
	if globalCli == nil {
		err := initCli(url)
		if err != nil {
			panic(err)
		}
	}
	if globalUrl != url {
		panic(fmt.Sprintf("Multiple url, %s, %s", globalUrl, url))
	}
	return globalCli
}

func initCli(url string) error {
	mu.Lock()
	defer mu.Unlock()
	if globalCli == nil {
		rpc, err := client.NewRpcClient(url)
		if err != nil {
			return err
		}
		globalCli = rpc
		globalUrl = url
	}
	return nil
}
