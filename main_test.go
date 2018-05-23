package main

import (
	"os"
	"testing"

	"github.com/ChimeraCoder/anaconda"

	"./nebulas"
)

var acc, _ = newAccount(nil)
var tweet = anaconda.Tweet{Text: "@NebBot please send 5 NAS, thanks"}
var tx, _ = newTx(txParams{
	acc.addr,
	acc.addr,
	uint128(0),
	1,
	uint128(1000000),
	uint128(2000000),
	core.TxPayloadBinaryType,
	nil,
})

func TestParseStatus(t *testing.T) {
	amount, err := parseStatus(tweet)
	if err != nil {
		t.Error(err)
	} else if amount != 5 {
		t.Errorf("Amount was incorrect, got: %v, want: %v.\n", amount, 5)
	}

	amount, err = parseStatus(anaconda.Tweet{Text: "This shouldn't work"})
	if err == nil {
		t.Errorf("Invalid argument didn't throw an error: %v\n", amount)
	} else if amount != 0 {
		t.Errorf("Amount was incorrect, got: %v, want: %v.\n", amount, 0)
	}

	amount, err = parseStatus(anaconda.Tweet{Text: "@NebBot send five NAS"})
	if err == nil {
		t.Errorf("Invalid argument didn't throw an error: %v\n", amount)
	} else if amount != 0 {
		t.Errorf("Amount was incorrect, got: %v, want: %v.\n", amount, 0)
	}
}

func TestEncyption(t *testing.T) {
	os.Setenv("secret", "123456789abcdefg")

	enc, err := encrypt(acc)
	if err != nil {
		t.Error(err)
	}

	_, err = decrypt(enc)
	if err != nil {
		t.Error(err)
	}
}

func TestGetAcc(t *testing.T) {
	var nonce uint64
	_, err := getAcc(123456, 123456, &nonce)
	if err != nil && err != errorNotInStorage {
		t.Error(err)
	}
}

func TestSignTx(t *testing.T) {
	err := signTransaction(acc, tx)
	if err != nil {
		t.Error(err)
	}
}

func TestEncodeRawTx(t *testing.T) {
	_, err := encodeRawTx(tx)
	if err != nil {
		t.Error(err)
	}
}
