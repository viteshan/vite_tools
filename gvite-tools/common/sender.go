package core

import (
	"fmt"
	"math/big"
	"time"

	"github.com/vitelabs/go-vite/log15"

	"github.com/vitelabs/go-vite/client"
	"github.com/vitelabs/go-vite/common/types"
	"github.com/vitelabs/go-vite/interfaces/core"
	"github.com/vitelabs/go-vite/wallet/hd-bip/derivation"
)

func NewSender(rpc client.RpcClient, self types.Address, key *derivation.Key) (*Sender, error) {
	cli, err := client.NewClient(rpc)
	if err != nil {
		return nil, err
	}
	return &Sender{Rpc: rpc, Cli: cli, Self: self, key: key}, nil
}

type Sender struct {
	Rpc  client.RpcClient
	Cli  client.Client
	Self types.Address
	key  *derivation.Key
}

func (s Sender) Send(params client.RequestTxParams, prev *core.HashHeight) (*core.HashHeight, error) {
	block, err := s.Cli.BuildNormalRequestBlock(params, prev)
	if err != nil {
		return nil, err
	}

	err = s.Cli.SignDataWithPriKey(s.key, block)
	if err != nil {
		return nil, err
	}

	// fmt.Printf("%v\n", common.ToJson(block))
	err = s.Rpc.SendRawTx(block)
	if err != nil {
		return nil, err
	}
	return BlockToHashHeight(block)
}

type SimpleRequestTx struct {
	ToAddr  types.Address
	Amount  *big.Int
	TokenId types.TokenTypeId
	Data    []byte
}

func (s Sender) BatchSend(logs []*SimpleRequestTx, prev *core.HashHeight) ([]*core.HashHeight, error) {
	var tmpPrev = prev
	var result []*core.HashHeight

	for k, v := range logs {
		time.Sleep(time.Millisecond * 100)

		params := client.RequestTxParams{
			ToAddr:   v.ToAddr,
			SelfAddr: s.Self,
			Amount:   v.Amount,
			TokenId:  v.TokenId,
			Data:     v.Data,
		}

		for i := 0; ; i++ {
			// fmt.Printf("%s,%s,%s,%s,%s\n", params.ToAddr, params.SelfAddr, params.Amount, params.TokenId, params.Data)
			hashHeight, err := s.Send(params, tmpPrev)
			if err != nil {
				if i < 3 {
					sleepTime := time.Second * time.Duration(3*(i+1))
					log15.Error(fmt.Sprintf("[%d]submit request tx[%s-%s] fail, sleep %s", k, v.ToAddr, v.Amount, sleepTime), "err", err, "prev", tmpPrev)
					time.Sleep(sleepTime)
					continue
				}
				log15.Error(fmt.Sprintf("[%d]submit request tx[%s-%s] fail.", k, v.ToAddr, v.Amount), "err", err, "prev", tmpPrev)
				return nil, err
			}
			result = append(result, hashHeight)
			tmpPrev = hashHeight
			log15.Info(fmt.Sprintf("[%d]transfer to[%s-%s] success.", k, v.ToAddr, v.Amount))
			break
		}
	}
	return result, nil
}
