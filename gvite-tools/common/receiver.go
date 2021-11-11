package core

import (
	"fmt"
	"time"

	"github.com/go-errors/errors"

	"github.com/vitelabs/go-vite/client"
	"github.com/vitelabs/go-vite/common/types"
	ledger "github.com/vitelabs/go-vite/interfaces/core"
	"github.com/vitelabs/go-vite/log15"
	"github.com/vitelabs/go-vite/wallet/hd-bip/derivation"
)

func NewReceiver(rpc client.RpcClient, self types.Address, key *derivation.Key) (*Receiver, error) {
	cli, err := client.NewClient(rpc)
	if err != nil {
		return nil, err
	}
	return &Receiver{Rpc: rpc, Cli: cli, Self: self, key: key}, nil
}

type Receiver struct {
	Rpc  client.RpcClient
	Cli  client.Client
	Self types.Address
	key  *derivation.Key
}

func (s Receiver) Receive(params client.ResponseTxParams, prev *ledger.HashHeight) (*ledger.HashHeight, error) {
	if params.SelfAddr != s.Self {
		return nil, errors.New("addr not match")
	}
	block, err := s.Cli.BuildResponseBlock(params, prev)
	if err != nil {
		return nil, err
	}

	err = s.Cli.SignDataWithPriKey(s.key, block)
	if err != nil {
		return nil, err
	}

	err = s.Rpc.SendRawTx(block)
	if err != nil {
		return nil, err
	}
	return BlockToHashHeight(block)
}

func (s Receiver) BatchReceive(logs []client.ResponseTxParams, prev *ledger.HashHeight) ([]*ledger.HashHeight, error) {
	var tmpPrev = prev
	var result []*ledger.HashHeight

	for k, v := range logs {
		time.Sleep(time.Millisecond * 100)

		for i := 0; ; i++ {
			hashHeight, err := s.Receive(v, tmpPrev)
			if err != nil {
				if i < 3 {
					sleepTime := time.Second * time.Duration(3*(i+1))
					log15.Error(fmt.Sprintf("[%d]submit response tx[%s-%s] fail, sleep %s", k, v.SelfAddr, v.RequestHash, sleepTime), "err", err, "prev", tmpPrev)
					time.Sleep(sleepTime)
					continue
				}
				log15.Error(fmt.Sprintf("[%d]submit response tx[%s-%s] fail.", k, v.SelfAddr, v.RequestHash), "err", err, "prev", tmpPrev)
				return nil, err
			}
			result = append(result, hashHeight)
			tmpPrev = hashHeight
			log15.Info(fmt.Sprintf("[%d]recive response addr[%s] reqHash[%s] response[%s-%d] success.", k, v.SelfAddr, v.RequestHash, hashHeight.Hash, hashHeight.Height))
			break
		}
	}
	return result, nil
}
