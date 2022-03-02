package main

import (
	"encoding/csv"
	"fmt"
	"math/big"
	"os"

	"github.com/urfave/cli/v2"
	"github.com/vitelabs/go-vite/v2/common/types"
	comm "github.com/viteshan/gvite-tools/common"
)

var batchSendFlags = []cli.Flag{
	&cli.StringFlag{
		Name:  "targetFile",
		Value: "received.csv",
	},
	&cli.StringFlag{
		Name: "tokenId",
	},
	&cli.BoolFlag{
		Name:  "prePrint",
		Value: false,
	},
}

func batchSend(c *cli.Context) error {
	tokenId := c.String("tokenId")
	targetFile := c.String("targetFile")
	prePrint := c.Bool("prePrint")

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

	sender, err := comm.NewSenderFromCli(c)
	if err != nil {
		return err
	}
	fromAddr := sender.Self

	{
		balance, _, err := sender.Cli.GetBalance(fromAddr, tti)
		if err != nil {
			return err
		}
		if balance.Cmp(sum) < 0 {
			return fmt.Errorf("balance is not enough")
		}
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
