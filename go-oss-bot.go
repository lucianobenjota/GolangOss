package main

import (
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	tb "gopkg.in/tucnak/telebot.v2"
)

func main() {
  err := godotenv.Load()
  if err != nil {
    log.Fatal("Error al leer variables de entorno")
  }

  tgApi := os.Getenv("TELEGRAM_API_KEY")

	b, err := tb.NewBot(tb.Settings{
		// You can also set custom API URL.
		// If field is empty it equals to "https://api.telegram.org".
		Token:  tgApi,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})

	if err != nil {
		log.Fatal(err)
		return
	}

  b.Handle("/start", func(m *tb.Message){
    b.Send(m.Sender, "Bienvenido al megabot")
  })

	b.Handle("/hello", func(m *tb.Message) {
		b.Send(m.Sender, "Hello World!")
	})

	b.Start()
}
