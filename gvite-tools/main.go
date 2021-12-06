package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"

	"github.com/urfave/cli/v2"
	"github.com/vitelabs/go-vite/client"
	"github.com/vitelabs/go-vite/common/types"
	comm "github.com/viteshan/gvite-tools/common"
	tools_dex "github.com/viteshan/gvite-tools/dex"
)

func main() {
	app := &cli.App{
		Usage: "vite tools",
		Commands: []*cli.Command{
			{
				Name:   "configMine",
				Usage:  "configMine",
				Action: tools_dex.ConfigMineAction,
				Flags:  append(senderFlags, tools_dex.ConfigMineFlags...),
			},
			{
				Name:   "batchSend",
				Usage:  "batch send with csv file",
				Action: batchSend,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "targetFile",
						Value: "received.csv",
					},
					&cli.StringFlag{
						Name:  "rpcUrl",
						Value: "https://node.vite.net/gvite",
					},
					&cli.StringFlag{
						Name: "tokenId",
					},
					&cli.BoolFlag{
						Name:  "prePrint",
						Value: false,
					},
					&cli.StringFlag{
						Name:     "mnemonic",
						Usage:    "mnemonic for the sender",
						FilePath: "vite.mnemonic",
					},
					&cli.StringFlag{
						Name: "fromAddr",
					},
				},
			}, {
				Name:   "autoReceive",
				Usage:  "receive all tx",
				Action: autoReceive,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "rpcUrl",
						Value: "https://node.vite.net/gvite",
					},
					&cli.StringFlag{
						Name:     "mnemonic",
						Usage:    "mnemonic for the sender",
						FilePath: "vite.mnemonic",
					},
					&cli.StringFlag{
						Name: "address",
					},
				},
			},
		},
	}
	err := app.Run(os.Args)

	if err != nil {
		log.Fatal(err)
	}
}

func autoReceive(c *cli.Context) error {
	url := c.String("rpcUrl")
	addrS := c.String("address")
	mnemonic := strings.TrimSpace(c.String("mnemonic"))

	fmt.Printf("url: %s\n", url)
	fmt.Printf("address: %s\n", addrS)

	address, err := types.HexToAddress(addrS)
	if err != nil {
		return err
	}

	cli := comm.GetCli(url)

	key, _, err := comm.DerivationKey(mnemonic, address)
	if err != nil {
		return err
	}
	for {
		blocks, err := cli.GetOnroadBlocksByAddress(address, 0, 100)
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
				SelfAddr:    address,
				RequestHash: v.Hash,
			})
			fmt.Printf("hash: %s, amount:%s \n", v.Hash, *v.Amount)
		}
		receiver, err := comm.NewReceiver(cli, address, key)
		if err != nil {
			panic(err)
		}
		_, err = receiver.BatchReceive(logs, nil)

		if err != nil {
			return err
		}
	}
	return nil
}

func batchSend(c *cli.Context) error {
	url := c.String("rpcUrl")
	tokenId := c.String("tokenId")
	targetFile := c.String("targetFile")
	prePrint := c.Bool("prePrint")
	fromAddrS := c.String("fromAddr")
	mnemonic := strings.TrimSpace(c.String("mnemonic"))

	fmt.Printf("url: %s\n", url)
	fmt.Printf("tokenId: %s\n", tokenId)
	fmt.Printf("targetFile: %s\n", targetFile)
	fmt.Printf("prePrint: %v\n", prePrint)

	receives, err := readCsv(targetFile)
	if err != nil {
		return err
	}
	tti, err := types.HexToTokenTypeId(tokenId)
	if err != nil {
		return err
	}
	for _, receive := range receives {
		receive.TokenId = tti
		receive.Data = []byte{}
	}

	total := 0
	sum := big.NewInt(0)
	for _, v := range receives {
		sum.Add(sum, v.Amount)
		total = total + 1
	}

	fmt.Printf("total: %v\n", total)
	fmt.Printf("sum: %s\n", sum.String())
	if len(receives) == 0 {
		fmt.Println("zero receiver")
		return nil
	}

	if prePrint {
		return nil
	}
	var RawUrl = url
	cli := comm.GetCli(RawUrl)

	fromAddr, err := types.HexToAddress(fromAddrS)
	if err != nil {
		return err
	}
	{
		viteCli, err := client.NewClient(cli)
		if err != nil {
			return err
		}
		balance, _, err := viteCli.GetBalance(fromAddr, tti)
		if err != nil {
			return err
		}
		if balance.Cmp(sum) < 0 {
			return fmt.Errorf("balance is not enough")
		}
	}

	// 	t.Log(balance, big.NewInt(0).SetBytes(balance.Bytes()).Div(balance, big.NewInt(1e18)).String(), onroad)
	if mnemonic == "" {
		return fmt.Errorf("empty mnemonic, please check vite.mnemonic file")
	}
	key, _, err := comm.DerivationKey(mnemonic, fromAddr)
	if err != nil {
		return err
	}

	sender, err := comm.NewSender(cli, fromAddr, key)
	if err != nil {
		return err
	}
	blocks, err := sender.BatchSend(receives, nil)
	if err != nil {
		return err
	}

	for _, v := range blocks {
		fmt.Println(fromAddr, v.Height, v.Hash)
	}
	fmt.Println("done")
	return nil
}

func readCsv(targetFile string) ([]*comm.SimpleRequestTx, error) {
	csvFile, err := os.Open(targetFile)
	if err != nil {
		return nil, err
	}
	defer csvFile.Close()

	csvLines, err := csv.NewReader(csvFile).ReadAll()
	if err != nil {
		return nil, err
	}
	var result []*comm.SimpleRequestTx
	for _, line := range csvLines {
		amount, flag := big.NewInt(0).SetString(line[1], 10)
		if !flag {
			return nil, fmt.Errorf("error amount: %s", line[1])
		}
		addr, err := types.HexToAddress(line[0])
		if err != nil {
			return nil, err
		}
		a := &comm.SimpleRequestTx{
			ToAddr: addr,
			Amount: amount,
		}
		result = append(result, a)
	}
	return result, nil
}

var senderFlags = []cli.Flag{
	&cli.StringFlag{
		Name:  "rpcUrl",
		Value: "https://node.vite.net/gvite",
	},
	&cli.StringFlag{
		Name:     "mnemonic",
		Usage:    "mnemonic for the sender",
		FilePath: "vite.mnemonic",
	},
	&cli.StringFlag{
		Name:     "privateKey",
		Usage:    "private key for the sender",
		FilePath: "vite.key",
	},
	&cli.StringFlag{
		Name: "fromAddr",
	},
}
