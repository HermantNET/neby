package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"./nebulas"
	"github.com/ChimeraCoder/anaconda"
)

const botID int64 = 997554387227684865

type waiter struct {
	StatusID            int64
	SenderID            int64
	SenderScreenName    string
	RecipientID         int64
	RecipientScreenName string
	Amount              float64
}

var api = anaconda.NewTwitterApiWithCredentials(
	os.Getenv("accessToken"),
	os.Getenv("accessSecret"),
	os.Getenv("consumerKey"),
	os.Getenv("consumerSecret"),
)

var waitingForConfirmation = sync.Map{}
var waitingForAddress = sync.Map{}

// Wait for @bot mentions to instigate a transaction
func stream() {
	userStream := api.UserStream(nil)

	for t := range userStream.C {
		switch status := t.(type) {
		case anaconda.Tweet:
			if status.User.Id != botID {
				amount, err := parseStatus(status)
				if err == nil && amount != 0 && status.InReplyToStatusID != 0 {
					go confirmUserTx(status, amount)
				}
			}
		case anaconda.DirectMessage:
			if confirmed := confirmUserTxResponse(status); !confirmed {
				err := parseChatCmds(status)
				if err != nil {
					api.PostDMToUserId(fmt.Sprintf("Sorry, something went wrong. Error: %v", err), status.SenderId)
				}
			}
		default:
		}
	}
}

func clean(s string) string {
	return strings.TrimSpace(s)
}

func cleanLower(s string) string {
	return strings.TrimSpace(strings.ToLower(s))
}

func parseChatCmds(msg anaconda.DirectMessage) error {
	m := cleanLower(msg.Text)

	if len(m) >= 46 && strings.HasPrefix(m, "transfer ") {
		var nonce uint64
		senderAcc, err := getAcc(msg.SenderId, msg.SenderId, &nonce)
		if err != nil {
			return err
		}

		addr, err := core.AddressParse(clean(msg.Text)[9:44])
		if err != nil {
			return err
		}

		amount, err := strconv.ParseFloat(m[45:], 64)
		if err != nil {
			return err
		}

		amt := uint64(amount * 1000000000000000000)

		tx, err := newTx(txParams{
			senderAcc.addr,
			addr,
			uint128(amt),
			nonce + 1,
			uint128(1000000),
			uint128(2000000),
			core.TxPayloadBinaryType,
			nil,
		})
		if err != nil {
			return err
		}

		err = signTransaction(senderAcc, tx)
		if err != nil {
			return err
		}

		encoded, err := encodeRawTx(tx)
		if err != nil {
			return err
		}

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
			return errors.New(parsed["error"])
		}

		api.PostDMToUserId("Transaction sent. View your pending transactions at https://explorer.nebulas.io/", msg.SenderId)
		return nil
	}

	switch m {
	case "help":
		api.PostDMToUserId("Available commands: help, address, transfer", msg.SenderId)
	case "address":
		go func(msg anaconda.DirectMessage) {
			var nonce uint64
			a, err := getAcc(msg.SenderId, msg.SenderId, &nonce)
			if err != nil {
				fmt.Println(err)
				api.PostDMToUserId("Sorry, something went wrong.", msg.SenderId)
			} else {
				api.PostDMToUserId(fmt.Sprintf("Your NAS address is: %s", a.addr), msg.SenderId)
			}
		}(msg)
	case "transfer":
		api.PostDMToUserId(`To transfer NAS to another address, type "transfer your_address_here amount"`, msg.SenderId)
	}
	return nil
}

func confirmUserTx(status anaconda.Tweet, amount float64) {
	waitingForConfirmation.Store(status.User.Id, waiter{
		status.Id,
		status.User.Id,
		status.User.ScreenName,
		status.InReplyToUserID,
		status.InReplyToScreenName,
		amount,
	})
	msg := fmt.Sprintf("CONFIRMATION: Send %f NAS to @%v? (yes/NO)", amount, status.InReplyToScreenName)
	_, err := api.PostDMToUserId(msg, status.User.Id)
	if err != nil {
		fmt.Println(err)
		return
	}

	go confirmTxTimeout(status.User.Id)
}

func confirmUserTxResponse(dm anaconda.DirectMessage) bool {
	if len(dm.Text) < 2 || len(dm.Text) > 4 {
		return false
	}
	if response := cleanLower(dm.Text); response == "yes" {
		raw, ok := waitingForConfirmation.Load(dm.SenderId)
		if ok == true {
			w, _ := raw.(waiter)
			waitingForConfirmation.Delete(w.SenderID)
			hash, err := startTx(w)
			if err == nil {
				api.PostDMToUserId("Transaction sent. View your pending transactions at https://explorer.nebulas.io/", w.SenderID)
				tweetTransactionSuccess(w, hash)
			} else {
				api.PostDMToUserId(fmt.Sprintf("Transaction failed.\nReason: %v", err), w.SenderID)
			}
			return true
		} else if response == "no" {
			cancelTx(dm.SenderId)
			return true
		}
	}
	return false
}

func getAcc(id int64, senderID int64, nonce *uint64) (acc account, err error) {
	address, err2 := getAddress(id)
	if err2 != nil {
		if err2 == errorDecodeJSON {
			err = err2
			return
		}

		if _, ok := waitingForAddress.Load(id); ok {
			err = errors.New("generating address, please wait")
			return
		}

		waitingForAddress.Store(senderID, true)
		go time.AfterFunc(time.Second*90, func() { waitingForAddress.Delete(senderID) })
		acc, err = newAccount(nil)
		if err != nil {
			return
		}

		err = setAddress(acc, senderID)
		if err == nil {
		}

	} else {
		acc, err = newAccount(address)
		if err != nil {
			return
		}

		var accInfo map[string]interface{}
		accInfo, err = accountInfo(acc.addr)
		if err != nil {
			return
		} else if accInfo["error"] != nil {
			return acc, fmt.Errorf("Remote error: %s", accInfo["error"].(string))
		}

		nonceRaw := accInfo["result"].(map[string]interface{})["nonce"].(string)
		*nonce, err = strconv.ParseUint(nonceRaw, 10, 64)
	}
	return
}

func startTx(w waiter) (string, error) {
	api.PostDMToUserId("Starting transaction...", w.SenderID)

	var nonce uint64
	senderAcc, err := getAcc(w.SenderID, w.SenderID, &nonce)
	if err != nil {
		return "", err
	}

	recipientAcc, err := getAcc(w.RecipientID, w.SenderID, &nonce)
	if err != nil {
		return "", err
	}

	amt := uint64(w.Amount * 1000000000000000000)

	tx, err := newTx(txParams{
		senderAcc.addr,
		recipientAcc.addr,
		uint128(amt),
		nonce + 1,
		uint128(1000000),
		uint128(2000000),
		core.TxPayloadBinaryType,
		nil,
	})
	if err != nil {
		return "", err
	}

	err = signTransaction(senderAcc, tx)
	if err != nil {
		return "", err
	}

	encoded, err := encodeRawTx(tx)
	err = tx.VerifyIntegrity(chainID)
	if err != nil {
		return "", err
	}

	resp, err := postRawTx(encoded)
	if err != nil {
		return "", err
	}

	body := readBody(resp)
	var parsed map[string]interface{}
	err = json.Unmarshal(body, &parsed)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", errors.New(parsed["error"].(string))
	}

	return parsed["result"].(map[string]interface{})["txhash"].(string), nil
}

func cancelTx(senderID int64) {
	api.PostDMToUserId("Transaction not sent.", senderID)
	waitingForConfirmation.Delete(senderID)
}

func confirmTxTimeout(userID int64) {
	time.Sleep(5 * time.Minute)
	if _, ok := waitingForConfirmation.Load(userID); ok {
		api.PostDMToUserId("TIMEOUT: Defaulted to NO. Transaction not sent.", userID)
		waitingForConfirmation.Delete(userID)
	}
}

func parseStatus(status anaconda.Tweet) (amount float64, err error) {
	r, _ := regexp.Compile("@NebBot (send|gift|give|wire|grant|drop|donate) ")
	if r.MatchString(status.Text) {
		match := r.FindStringIndex(status.Text)
		end := strings.Index(status.Text, " NAS")

		amount, err := strconv.ParseFloat(status.Text[match[len(match)-1]:end], 64)
		if err != nil {
			return 0, err
		}
		return amount, nil
	}
	return 0, errors.New("does not match")
}

func reaction() string {
	w := []string{
		"How wonderful!",
		"Awesome,",
		"Now that's generous,",
		"Rock on!",
		"Marvelous,",
		"The one and only",
		"Really? Really.",
		"Wow,",
	}

	return fmt.Sprintf("%s", w[rand.Intn(len(w))])
}

// Send a tweet in reply to the instigating tweet to confirm the transaction succeeded.
func tweetTransactionSuccess(w waiter, hash string) {
	v := url.Values{}

	v.Add("in_reply_to_status_id", string(w.StatusID))
	api.PostTweet(fmt.Sprintf("%v @%v sent %f NAS to @%v. TX: %v", reaction(), w.SenderScreenName, w.Amount, w.RecipientScreenName, hash), v)
}
