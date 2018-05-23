package main

import (
	"fmt"

	"./nebulas"
	"./nebulas/crypto"
	"./nebulas/crypto/keystore"
)

type account struct {
	priv keystore.PrivateKey
	addr *core.Address
}

func newAccount(privKey []byte) (account, error) {
	priv, err := crypto.NewPrivateKey(keystore.SECP256K1, privKey)
	if err != nil {
		return account{}, err
	}

	pub, err := priv.PublicKey().Encoded()
	if err != nil {
		return account{}, err
	}

	addr, err := core.NewAddressFromPublicKey(pub)
	if err != nil {
		return account{}, err
	}

	return account{priv, addr}, nil
}

func getPrivateKeyByteArray(account account) ([]byte, error) {
	privBytes, err := account.priv.Encoded()
	if err != nil {
		return nil, err
	}

	return privBytes, nil
}

func printBytes(bytes []byte) {
	for _, b := range bytes {
		fmt.Printf("%v, ", b)
	}
	fmt.Println()
}
