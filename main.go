package main

import (
	"encoding/hex"
	"fmt"
	"os"
	"time"

	"./nebulas/util"
	_ "github.com/joho/godotenv/autoload"
)

var contractAddress = "n1kWMooMHDAnLQXe6pLdZ6xDRBXCVJHuPcJ"
var botPriv, _ = hex.DecodeString(os.Getenv("bot"))
var bot, _ = newAccount(botPriv)

func uint128(i uint64) *util.Uint128 {
	return util.NewUint128FromUint(i)
}

func main() {
	defer persist()
	fmt.Println("Nastwitter v1")
	go stream()
}

func persist() {
	t := time.NewTicker(time.Hour * 24)

	for i := range t.C {
		fmt.Printf("Process has been running for %v days.\n", i)
	}
}
