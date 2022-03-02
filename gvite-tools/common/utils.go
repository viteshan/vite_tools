package core

import (
	"strconv"

	"github.com/go-errors/errors"
	bip39 "github.com/tyler-smith/go-bip39"
	"github.com/vitelabs/go-vite/v2/common/types"
	ledger "github.com/vitelabs/go-vite/v2/interfaces/core"
	"github.com/vitelabs/go-vite/v2/rpcapi/api"
	"github.com/vitelabs/go-vite/v2/wallet/entropystore"
	"github.com/vitelabs/go-vite/v2/wallet/hd-bip/derivation"
)

func BlockToHashHeight(block *api.AccountBlock) (*ledger.HashHeight, error) {
	u, err := strconv.ParseUint(block.Height, 10, 64)
	if err != nil {
		return nil, err
	}
	return &ledger.HashHeight{Hash: block.Hash, Height: u}, err
}

func DerivationKey(mnemonic string, addr types.Address) (key *derivation.Key, index uint32, e error) {
	seed := bip39.NewSeed(mnemonic, "")
	return entropystore.FindAddrFromSeed(seed, addr, 100)
}

func RandomMnemonic() (string, error) {
	entropy, err := bip39.NewEntropy(256)
	if err != nil {
		return "", err
	}
	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return "", err
	}
	seed := bip39.NewSeed(mnemonic, "")
	derivation.DeriveWithIndex(0, seed)
	return mnemonic, nil
}

func EntropystoreToMnemonic(filename, pwd string) (string, error) {
	mayValid, _, e := entropystore.IsMayValidEntropystoreFile(filename)
	if e != nil {
		return "", e
	}
	if !mayValid {
		return "", errors.New("not valid entropy store file")
	}
	ks := entropystore.CryptoStore{EntropyStoreFilename: filename}

	entropy, e := ks.ExtractEntropy(pwd)
	if e != nil {
		return "", e
	}
	return bip39.NewMnemonic(entropy)

}

func Derivation(mnemonic string, index uint32, err error) (*derivation.Key, *types.Address, error) {
	if err != nil {
		return nil, nil, err
	}
	seed := bip39.NewSeed(mnemonic, "")
	key, err := derivation.DeriveWithIndex(index, seed)
	if err != nil {
		return nil, nil, err
	}
	addr, err := key.Address()
	if err != nil {
		return nil, nil, err
	}
	return key, addr, nil
}

func DervationFromEntropystore(filename, pwd string) (*derivation.Key, *types.Address, error) {
	mayValid, _, e := entropystore.IsMayValidEntropystoreFile(filename)
	if e != nil {
		return nil, nil, e
	}
	if !mayValid {
		return nil, nil, errors.New("not valid entropy store file")
	}
	ks := entropystore.CryptoStore{EntropyStoreFilename: filename}

	seed, _, e := ks.ExtractSeed(pwd)
	if e != nil {
		return nil, nil, e
	}
	key, err := derivation.DeriveWithIndex(0, seed)
	if err != nil {
		return nil, nil, err
	}
	addr, err := key.Address()
	if err != nil {
		return nil, nil, err
	}
	return key, addr, nil
}

func RandomEntropystore(dir, pwd string) (string, error) {
	mnemonic, err := RandomMnemonic()
	if err != nil {
		return "", err
	}
	entropy, e := bip39.EntropyFromMnemonic(mnemonic)
	if e != nil {
		return "", e
	}
	primaryAddress, e := entropystore.MnemonicToPrimaryAddr(mnemonic)
	if e != nil {
		return "", e
	}
	filename := entropystore.FullKeyFileName(dir, *primaryAddress)
	e = entropystore.CryptoStore{EntropyStoreFilename: filename}.StoreEntropy(entropy, *primaryAddress, pwd)
	if e != nil {
		return "", e
	}
	return filename, nil
}
