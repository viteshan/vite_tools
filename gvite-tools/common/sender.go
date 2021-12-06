package core

import (
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/vitelabs/go-vite/crypto/ed25519"

	"github.com/urfave/cli/v2"
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
	pk, err := key.PrivateKey()
	if err != nil {
		return nil, err
	}
	return &Sender{Rpc: rpc, Cli: cli, Self: self, key: pk}, nil
}

func NewSenderWithHexPrivateKey(rpc client.RpcClient, self types.Address, key string) (*Sender, error) {
	cli, err := client.NewClient(rpc)
	if err != nil {
		return nil, err
	}
	pk, err := ed25519.HexToPrivateKey(key)
	if err != nil {
		return nil, err
	}
	return &Sender{Rpc: rpc, Cli: cli, Self: self, key: pk}, nil
}

type Sender struct {
	Rpc  client.RpcClient
	Cli  client.Client
	Self types.Address
	key  ed25519.PrivateKey
}

func (s Sender) Send(params client.RequestTxParams, prev *core.HashHeight) (*core.HashHeight, error) {
	block, err := s.Cli.BuildNormalRequestBlock(params, prev)
	if err != nil {
		return nil, err
	}
	err = s.Cli.SignDataWithEd25519Key(s.key, block)
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

func NewSenderFromCli(c *cli.Context) (*Sender, error) {
	url := c.String("rpcUrl")
	fromAddrS := c.String("fromAddr")
	mnemonic := strings.TrimSpace(c.String("mnemonic"))
	privateKey := strings.TrimSpace(c.String("privateKey"))

	fmt.Printf("url: %s\n", url)
	fmt.Printf("from Addr:%s\n", fromAddrS)

	var RawUrl = url
	cli := GetCli(RawUrl)

	fromAddr, err := types.HexToAddress(fromAddrS)
	if err != nil {
		return nil, err
	}

	// 	t.Log(balance, big.NewInt(0).SetBytes(balance.Bytes()).Div(balance, big.NewInt(1e18)).String(), onroad)
	if mnemonic == "" && privateKey == "" {
		return nil, fmt.Errorf("empty mnemonic and privateKey, please check vite.mnemonic or vite.key file")
	}
	if mnemonic != "" {
		key, _, err := DerivationKey(mnemonic, fromAddr)
		if err != nil {
			return nil, err
		}

		sender, err := NewSender(cli, fromAddr, key)
		if err != nil {
			return nil, err
		}
		return sender, nil
	}
	if privateKey != "" {
		sender, err := NewSenderWithHexPrivateKey(cli, fromAddr, privateKey)
		if err != nil {
			return nil, err
		}
		return sender, nil
	}
	return nil, fmt.Errorf("new sender fail")
}
