package main

import (
	"bytes"
	"crypto/aes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"./nebulas"
)

const mainnetURL = "https://mainnet.nebulas.io"
const chainID = 1

var errorNotInStorage = errors.New("account not in storage")
var errorUnexpectedLength = errors.New("unexpected length")

var client = &http.Client{
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

func accountInfo(a *core.Address) (parsed map[string]interface{}, err error) {
	url := fmt.Sprintf("%v/v1/user/accountstate", mainnetURL)
	data := fmt.Sprintf(`{"address": %q}`, a)
	resp, err := client.Post(url, "application/json", bytes.NewBuffer([]byte(data)))
	if err != nil {
		return parsed, err
	}

	body := readBody(resp)
	err = json.Unmarshal(body, &parsed)
	if err != nil {
		return parsed, err
	}
	return
}

func postRawTx(data string) (*http.Response, error) {
	url := fmt.Sprintf("%v/v1/user/rawtransaction", mainnetURL)
	resp, err := client.Post(url, "application/json", bytes.NewBuffer([]byte(data)))
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func getAddress(id int64) ([]byte, error) {
	url := fmt.Sprintf("%v/v1/user/call", mainnetURL)
	data := fmt.Sprintf(
		`{"from":%q, "to":%q, "value":"0", "nonce":0, "gasPrice":"1000000", "gasLimit":"2000000", "contract":{"function":"getAccount", "args":"[\"%v\"]"}}`,
		bot.addr,
		contractAddress,
		id,
	)

	resp, err := client.Post(url, "application/json", bytes.NewBuffer([]byte(data)))
	if err != nil {
		return nil, err
	}

	body := readBody(resp)
	var parsed map[string]interface{}
	err = json.Unmarshal(body, &parsed)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	switch parsed["result"].(type) {
	case map[string]interface{}:
		if e := parsed["result"].(map[string]interface{})["execute_err"].(string); e != "" {
			fmt.Println(err)
			return nil, errors.New(e)
		}

		r := parsed["result"].(map[string]interface{})["result"]
		if r == nil || r.(string) == "" || r.(string) == `{"account":null}` {
			fmt.Println(err)
			return nil, errorNotInStorage
		}

		result := make(map[string]interface{})
		err := json.Unmarshal([]byte(r.(string)), &result)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}

		key, err := decrypt(result["account"].(string))
		if err != nil {
			fmt.Println(result)
			fmt.Println(result["error"])
			fmt.Println(result["execute_err"])
			return nil, err
		}

		return key, nil
	}

	return nil, errors.New("no matching case")
}

func encrypt(acc account) (string, error) {
	bytes, err := getPrivateKeyByteArray(acc)
	if err != nil {
		return "", err
	}

	bc, err := aes.NewCipher([]byte(os.Getenv("secret")))
	if err != nil {
		return "", err
	}

	var dst = make([]byte, 32)
	for i := 0; i <= 32-16; i += 16 {
		bc.Encrypt(dst[i:], bytes[i:])
	}

	return hex.EncodeToString(dst), nil

}

func decrypt(d string) ([]byte, error) {
	if len(d) != 64 {
		return nil, errorUnexpectedLength
	}

	data, err := hex.DecodeString(d)
	if err != nil {
		return nil, err
	}

	bc, err := aes.NewCipher([]byte(os.Getenv("secret")))
	if err != nil {
		return nil, err
	}

	var dst = make([]byte, 32)
	for i := 0; i <= 32-16; i += 16 {
		bc.Decrypt(dst[i:], data[i:])
	}

	return dst, nil
}

func setAddress(acc account, id int64) error {
	accountInfo, err := accountInfo(bot.addr)
	if err != nil {
		return err
	} else if accountInfo["error"] != nil {
		return errors.New(accountInfo["error"].(string))
	}

	nonceRaw := accountInfo["result"].(map[string]interface{})["nonce"].(string)
	nonce, _ := strconv.ParseUint(nonceRaw, 10, 64)

	encrypted, err := encrypt(acc)
	if err != nil {
		return err
	}

	data := fmt.Sprintf(
		`{"function":"setAccount", "args":"[\"%v\",\"%s\"]"}`,
		id,
		encrypted,
	)

	payload := []byte(data)
	ca, err := core.AddressParse(contractAddress)
	if err != nil {
		return err
	}

	tx, err := newTx(txParams{bot.addr, ca, uint128(0), nonce + 1, uint128(1000000), uint128(2000000), core.TxPayloadCallType, payload})
	if err != nil {
		return err
	}

	err = signTransaction(bot, tx)
	if err != nil {
		return err
	}

	encoded, err := encodeRawTx(tx)
	err = tx.VerifyIntegrity(chainID)
	if err != nil {
		return err
	}

	resp, err := postRawTx(encoded)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		body := readBody(resp)
		var parsed map[string]string
		err := json.Unmarshal(body, &parsed)
		if err != nil {
			return err
		}

		if e := parsed["execute_err"]; e != "" {
			return errors.New(e)
		} else if e = parsed["error"]; e != "" {
			return errors.New(e)
		}
	}

	return nil
}
