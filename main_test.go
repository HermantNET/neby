package main

import (
	"encoding/json"
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

func TestDecrypt(t *testing.T) {
	_, err := decrypt("486b2395e856db1981bb611739193fee7e1e19f2402192bb83d3375b40655492")
	if err != nil {
		t.Error(err)
	}

	_, err = decrypt("")
	if err == nil {
		t.Error("Should fail.")
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

func TestReadResponse(t *testing.T) {
	body := []byte(`{"result":{"result":"\"432608e866e9833c118f9abc3b5c6a82a646f870f0fd1b49d332f81b58fbae5f\"","execute_err":"","estimate_gas":"20156"}}`)
	r := response{}
	err := json.Unmarshal(body, &r)
	if err != nil {
		t.Error(err)
	}

	if r.Result.Result == "" {
		t.Error("Result is empty.")
	}

	if r.Result.ExecuteErr != "" {
		t.Errorf("Execution error: %v", r.Result.ExecuteErr)
	}

	_, err = decrypt(r.Result.Result)
	if err != nil {
		t.Error(err)
	}
}

func TestParseTxResult(t *testing.T) {
	body := []byte(`{"result":{"txhash":"f37acdf93004f7a3d72f1b7f6e56e70a066182d85c186777a2ad3746b01c3b52"}}`)
	var parsed map[string]interface{}
	err := json.Unmarshal(body, &parsed)
	if err != nil {
		t.Error(err)
	}

	if parsed["result"].(map[string]interface{})["txhash"].(string) != "f37acdf93004f7a3d72f1b7f6e56e70a066182d85c186777a2ad3746b01c3b52" {
		t.Error("Error parsing result.")
	}
}

func TestReaction(t *testing.T) {
	for i := 0; i < 100; i++ {
		reaction()
	}
}
