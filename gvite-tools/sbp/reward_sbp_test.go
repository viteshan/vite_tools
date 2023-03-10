package sbp

import (
	"fmt"
	"math/big"
	"math/rand"
	"testing"

	"github.com/vitelabs/go-vite/v2/client"
	core "github.com/vitelabs/go-vite/v2/interfaces/core"
	comm "github.com/viteshan/vite_tools/gvite-tools/common"

	"github.com/vitelabs/go-vite/v2/common/types"
)

// var RawUrl = "http://116.63.158.55:48132"
var RawUrl = "https://node.vite.net/gvite"

// 发送奖励地址
var selfAddr = types.HexToAddressPanic("vite_xxx")

// 发送奖励助记词
var mnemonic = "xxxxx"

// sbp注册100w的算份额
var sbpFixAddr = types.HexToAddressPanic("vite_xxx")

var sbpName = "V.Morgen"

func TestCalReward(t *testing.T) {
	cli := comm.GetCli(RawUrl)
	startIdx := uint64(668)
	endIdx := uint64(683)
	for i := startIdx; i <= endIdx; i++ {
		fmt.Printf("start:%d, end: %d\n", startIdx, startIdx+6)
		oneP(startIdx, startIdx+6, cli, t)
		startIdx = startIdx + 7
	}
}
func oneP(startIdx, endIdx uint64, cli client.RpcClient, t *testing.T) {

	// startIdx := uint64(143)
	// endIdx := uint64(149)

	if endIdx < startIdx {
		t.Fatal("error for idx")
	}

	//todo reward address
	votes, err := CalReward(startIdx, endIdx, sbpName, sbpFixAddr, cli)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	//t.Log(votes)

	rewardTotal, err := MergeReward(votes)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	avg := big.NewInt(0).Set(rewardTotal)

	t.Log("total reward", big.NewInt(0).Div(rewardTotal, OnePercentVite).String(), avg.Div(avg, big.NewInt(int64(endIdx-startIdx+1))).String(), int64(endIdx-startIdx+1))

	details, err := CalRewardDropDetails(votes)

	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	// t.Log(details)

	{ // print to csv
		var logs = ""
		for i := startIdx; i <= endIdx; i++ {
			details := details[i]
			for kk, vv := range details {
				amount := big.NewInt(0).Set(vv.Amount)
				amountStr := amount.String()
				logs += fmt.Sprintf("%d,%s,%f,%s\n", i, kk, float64(amount.Div(vv.Amount, OnePercentVite).Int64())/100.0, amountStr)
			}
		}
		fmt.Println(logs)
	}
	return

	mergedDetails, err := MergeRewardDrop(details)

	{ // 打印一共发送了多少钱
		sendTotal := big.NewInt(0)
		for k, v := range mergedDetails {
			amount := big.NewInt(0).Set(v.Amount)
			t.Log(k, amount.Div(amount, OnePercentVite))
			sendTotal.Add(sendTotal, amount)
		}

		//t.Log("all send", sendTotal.Div(sendTotal, oneVite))
		t.Log("all send", sendTotal)
	}
	{ // 打印每万份投票每天大概是多少收益
		fixAvgVote := big.NewInt(0)
		allAvgVote := big.NewInt(0)
		{
			for k, v := range votes {
				s := big.NewInt(0).Div(v.VoteDetails[sbpFixAddr], OnePercentVite)
				t.Log(k, s.String())
				fixAvgVote.Add(fixAvgVote, s)
				allAvgVote.Add(allAvgVote, v.VoteTotal)
			}
			fixAvgVote.Div(fixAvgVote, big.NewInt(int64(len(votes))))
		}
		avgReward := big.NewInt(0).Set(mergedDetails[sbpFixAddr].Amount)
		avgReward.Div(avgReward, big.NewInt(int64(len(votes))))
		avgReward.Div(avgReward, OnePercentVite)
		fixAvgVote.Div(fixAvgVote, big.NewInt(100))

		allAvgVote.Div(allAvgVote, big.NewInt(int64(len(votes))))
		t.Log("sbpFixAddr", allAvgVote.Div(allAvgVote, OneVite), avgReward.Div(avgReward, fixAvgVote.Div(fixAvgVote, big.NewInt(10000))))
	}

	txId := fmt.Sprintf("%d", rand.Intn(10000000))

	var finalLog []*comm.SimpleRequestTx

	var tokenId = core.ViteTokenId
	{
		for _, v := range mergedDetails {
			amount := big.NewInt(0).Set(v.Amount)
			if amount.Cmp(OnePercentVite) < 0 {
				t.Log("ignore amount", v.ToAddr, v.Amount)
				continue
			}
			finalLog = append(finalLog, &comm.SimpleRequestTx{ToAddr: v.ToAddr, Amount: amount, TokenId: tokenId, Data: []byte(txId)})
		}

		for k, v := range finalLog {
			t.Log(k, v.ToAddr, v.Amount.String())
		}
	}

	return

	// key, _, err := comm.DerivationKey(mnemonic, selfAddr)
	// if err != nil {
	// 	t.Fatal(err)
	// }

	// if err != nil {
	// 	t.Error(err)
	// 	t.FailNow()
	// }

	// sender, err := comm.NewSender(cli, selfAddr, key)
	// if err != nil {
	// 	t.Error(err)
	// 	t.FailNow()
	// }

	// all, err := sender.BatchSend(finalLog, nil)
	// if err != nil {
	// 	t.Error(err)
	// 	t.FailNow()
	// }

	// for _, v := range all {
	// 	t.Log(selfAddr, v.Height, v.Hash)
	// }
}
