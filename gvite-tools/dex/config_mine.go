package dex

import (
	"fmt"
	"math/big"

	"github.com/urfave/cli/v2"
	"github.com/vitelabs/go-vite/v2/client"
	"github.com/vitelabs/go-vite/v2/common/types"
	"github.com/vitelabs/go-vite/v2/vm/contracts/abi"
	"github.com/vitelabs/go-vite/v2/vm/contracts/dex"
	comm "github.com/viteshan/vite_tools/gvite-tools/common"
)

var ConfigMineFlags = []cli.Flag{
	&cli.StringFlag{
		Name:     "tradeToken",
		Required: true,
	},
	&cli.StringFlag{
		Name:     "quoteToken",
		Required: true,
	},
	&cli.BoolFlag{
		Name: "enable",
	},
	&cli.StringFlag{
		Name:  "toAddress",
		Value: "vite_0000000000000000000000000000000000000006e82b8ba657",
	},
}

var viteTokenId = types.TokenTypeId{'V', 'I', 'T', 'E', ' ', 'T', 'O', 'K', 'E', 'N'}

func ConfigMineAction(c *cli.Context) error {
	data, err := configMine(c.String("tradeToken"), c.String("quoteToken"), c.Bool("enable"))

	if err != nil {
		return err
	}
	sender, err := comm.NewSenderFromCli(c)
	if err != nil {
		return err
	}
	toAddress, err := types.HexToAddress(c.String("toAddress"))
	if err != nil {
		return err
	}
	hashHeight, err := sender.Send(client.RequestTxParams{
		ToAddr:   toAddress,
		SelfAddr: sender.Self,
		Amount:   big.NewInt(0),
		TokenId:  viteTokenId,
		Data:     data,
	}, nil)

	if err != nil {
		return err
	}
	fmt.Printf("block send success, hash: %s, height:%d\n", hashHeight.Hash, hashHeight.Height)
	return nil
}

func configMine(tradeToken, quoteToken string, enable bool) ([]byte, error) {
	fmt.Println(tradeToken, quoteToken, enable)
	tt, err := types.HexToTokenTypeId(tradeToken)
	if err != nil {
		return nil, err
	}
	qt, err := types.HexToTokenTypeId(quoteToken)
	if err != nil {
		return nil, err
	}
	data, err := abi.ABIDexFund.PackMethod(abi.MethodNameDexFundTradeAdminConfig, uint8(dex.TradeAdminConfigMineMarket), tt, qt, enable, viteTokenId, uint8(1), uint8(1), big.NewInt(0), uint8(1), big.NewInt(0))
	if err != nil {
		return nil, err
	}
	return data, nil
}
