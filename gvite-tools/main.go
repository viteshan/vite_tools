package main

import (
	"log"
	"os"

	"github.com/urfave/cli/v2"
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
				Flags:  append(append([]cli.Flag(nil), senderFlags...), tools_dex.ConfigMineFlags...),
			}, {
				Name:   "batchSend",
				Usage:  "batch send with csv file",
				Action: batchSend,
				Flags:  append(append([]cli.Flag(nil), senderFlags...), batchSendFlags...),
			}, {
				Name:   "autoReceive",
				Usage:  "receive all tx",
				Action: autoReceive,
				Flags:  append([]cli.Flag(nil), senderFlags...),
			},
		},
	}
	err := app.Run(os.Args)

	if err != nil {
		log.Fatal(err)
	}
}

var senderFlags = []cli.Flag{
	&cli.StringFlag{
		Name:    "rpcUrl",
		Aliases: []string{"rpc"},
		Value:   "https://node.vite.net/gvite",
	},
	&cli.StringFlag{
		Name:     "mnemonic",
		Usage:    "mnemonic for the sender",
		FilePath: "vite.mnemonic",
	},
	&cli.StringFlag{
		Name:     "privateKey",
		Aliases:  []string{"key", "private"},
		Usage:    "private key for the sender",
		FilePath: "vite.key",
	},
	&cli.StringFlag{
		Name:    "accountAddress",
		Aliases: []string{"address", "addr", "account"},
	},
}
