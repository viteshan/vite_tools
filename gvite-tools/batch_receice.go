package main

import (
	"fmt"

	"github.com/urfave/cli/v2"
	"github.com/vitelabs/go-vite/v2/client"
	comm "github.com/viteshan/gvite-tools/common"
)

func autoReceive(c *cli.Context) error {
	sender, err := comm.NewSenderFromCli(c)
	if err != nil {
		return err
	}

	receiver, err := comm.NewReceiverFromSender(sender)
	if err != nil {
		return err
	}
	selfAddr := sender.Self
	for {
		blocks, err := sender.Rpc.GetOnroadBlocksByAddress(selfAddr, 0, 100)
		if err != nil {
			panic(err)
		}

		fmt.Println(len(blocks))
		if len(blocks) == 0 {
			break
		}

		var logs []client.ResponseTxParams

		for _, v := range blocks {
			logs = append(logs, client.ResponseTxParams{
				SelfAddr:    selfAddr,
				RequestHash: v.Hash,
			})
			fmt.Printf("from:%s ,hash: %s, amount:%s \n", v.AccountAddress, v.Hash, *v.Amount)
		}

		hashHeights, err := receiver.BatchReceive(logs, nil)

		if err != nil {
			return err
		}

		for _, hashHeight := range hashHeights {
			fmt.Printf("receive %s, hash:%s, height:%d\n", selfAddr, hashHeight.Hash, hashHeight.Height)
		}
	}
	return nil
}
