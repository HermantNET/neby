package main

import (
	"encoding/base64"
	"fmt"

	"./nebulas"
	"./nebulas/crypto"
	"./nebulas/crypto/keystore"
	"./nebulas/util"
	"github.com/gogo/protobuf/proto"
)

type txParams struct {
	to       *core.Address
	from     *core.Address
	value    *util.Uint128
	nonce    uint64
	gasPrice *util.Uint128
	gasLimit *util.Uint128
	txtype   string
	payload  []byte
}

func newTx(p txParams) (*core.Transaction, error) {
	return core.NewTransaction(
		chainID,
		p.to,
		p.from,
		p.value,
		p.nonce,
		p.txtype,
		p.payload,
		p.gasPrice,
		p.gasLimit,
	)
}

func encodeRawTx(tx *core.Transaction) (string, error) {
	msg, err := tx.ToProto()
	if err != nil {
		return "", err
	}

	wired, err := proto.Marshal(msg)
	if err != nil {
		return "", err
	}

	data := base64.StdEncoding.EncodeToString(wired)
	val := fmt.Sprintf(`{"data": %q}`, data)

	return val, nil
}

func signTransaction(account account, tx *core.Transaction) error {
	sig, err := crypto.NewSignature(keystore.SECP256K1)
	if err != nil {
		return err
	}

	sig.InitSign(account.priv)
	return tx.Sign(sig)
}
