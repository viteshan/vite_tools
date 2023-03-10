package sbp

import (
	"math/big"

	"github.com/pkg/errors"
	"github.com/vitelabs/go-vite/v2/client"
	"github.com/vitelabs/go-vite/v2/common/types"
	"github.com/vitelabs/go-vite/v2/rpcapi/api"
)

var OnePercentVite, _ = big.NewInt(0).SetString("10000000000000000", 10)
var OneVite, _ = big.NewInt(0).SetString("1000000000000000000", 10)
var OneViteFloat, _ = big.NewFloat(0).SetString("1000000000000000000")

type SBPRewardVote struct {
	SbpReward   *api.Reward
	VoteTotal   *big.Int
	VoteDetails map[types.Address]*big.Int
	Idx         uint64
}

type SBPRewardDropDetails struct {
	ToAddr types.Address
	Amount *big.Int
}

func CalReward(startIdx uint64, endIdx uint64, sbp string, sbpRewardAddr types.Address, rpc client.RpcClient) (map[uint64]*SBPRewardVote, error) {
	result := make(map[uint64]*SBPRewardVote)
	for i := startIdx; i <= endIdx; i++ {
		tmp := &SBPRewardVote{}
		result[i] = tmp
		reward, err := rpc.GetRewardByIndex(i)
		if err != nil {
			return nil, err
		}

		tmp.SbpReward = reward.RewardMap[sbp]
		details, err := rpc.GetVoteDetailsByIndex(i)
		if err != nil {
			return nil, err
		}

		for _, v := range details {
			if v.Name == sbp {
				tmp.VoteTotal, tmp.VoteDetails = reCalReward(sbpRewardAddr, v.Balance, v.Addr)
			}
		}
	}
	return result, nil
}

func reCalReward(sbpRewardAddr types.Address, totalBalance *big.Int, voteDetails map[types.Address]*big.Int) (*big.Int, map[types.Address]*big.Int) {
	sbpFixVote := big.NewInt(1000000)
	sbpFixVote.Mul(sbpFixVote, OneVite)

	reTotalBalance := big.NewInt(0)
	reTotalBalance.Add(reTotalBalance, totalBalance)
	reTotalBalance.Add(reTotalBalance, sbpFixVote)

	reSbp, ok := voteDetails[sbpRewardAddr]
	if !ok {
		voteDetails[sbpRewardAddr] = sbpFixVote
	} else {
		reSbp.Add(reSbp, sbpFixVote)
	}
	return reTotalBalance, voteDetails
}

func MergeReward(rewardVoteMap map[uint64]*SBPRewardVote) (*big.Int, error) {
	totalReward := big.NewInt(0)
	for _, v := range rewardVoteMap {
		total, flag := big.NewInt(0).SetString(v.SbpReward.TotalReward, 10)
		if !flag {
			return nil, errors.Errorf("can't parse [%s] to bigInt.", v.SbpReward.TotalReward)
		}
		totalReward.Add(totalReward, total)
	}
	return totalReward, nil
}

func CalRewardDropDetails(rewardVoteMap map[uint64]*SBPRewardVote) (map[uint64]map[types.Address]*SBPRewardDropDetails, error) {
	result := make(map[uint64]map[types.Address]*SBPRewardDropDetails)

	for k, v := range rewardVoteMap {
		if v.VoteTotal.Sign() <= 0 {
			continue
		}
		m := make(map[types.Address]*SBPRewardDropDetails)
		total := big.NewInt(0)
		for kk, vv := range v.VoteDetails {
			details := &SBPRewardDropDetails{ToAddr: kk}
			total.Add(total, vv)
			totalReward, flag := big.NewInt(0).SetString(v.SbpReward.TotalReward, 10)
			if !flag {
				return nil, errors.Errorf("bigInt fail %s", v.SbpReward.TotalReward)
			}
			totalReward.Mul(totalReward, vv)
			totalReward.Div(totalReward, v.VoteTotal)

			details.Amount = totalReward
			m[kk] = details
		}
		if total.Cmp(v.VoteTotal) != 0 {
			return nil, errors.Errorf("total vote not equal")
		}
		result[k] = m
	}
	return result, nil
}

func MergeRewardDrop(details map[uint64]map[types.Address]*SBPRewardDropDetails) (map[types.Address]*SBPRewardDropDetails, error) {
	result := make(map[types.Address]*SBPRewardDropDetails)
	for _, v := range details {
		for kk, vv := range v {
			dropDetails, ok := result[kk]
			if !ok {
				dropDetails = &SBPRewardDropDetails{ToAddr: kk, Amount: big.NewInt(0)}
				result[kk] = dropDetails
			}
			dropDetails.Amount.Add(dropDetails.Amount, vv.Amount)
		}
	}
	return result, nil
}
