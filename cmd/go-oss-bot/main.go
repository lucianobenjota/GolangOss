package main

import (
	"log"

	"github.com/lucianobenjota/go-oss-bot/m/pkg/bot"
)

func main() {
  err := bot.StartBot()
  if err != nil {
    log.Fatal(err)
  }
}
