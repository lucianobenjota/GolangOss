package main

import (
	"log"
	"os"
	"strconv"
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
  tgUser := os.Getenv("TELEGRAM_USER_ID")

  userId, err := strconv.ParseInt(tgUser, 10, 64)
  if err != nil {
    log.Fatal("Error al leer variables de entorno")
  }

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

  var (
    menu = &tb.ReplyMarkup{ResizeReplyKeyboard: true, OneTimeKeyboard: true}
    btnCompañia = menu.Text("🏢 Compañia")
    btnAyuda = menu.Text(" Ayuda")
  )

  menu.Reply(
    menu.Row(btnCompañia),
    menu.Row(btnAyuda),
  )

  b.Handle("/start", func(m *tb.Message){

    if m.Chat.ID != userId{
      log.Printf("El id %v envió un mensaje", m.Chat.ID)
      return
    }

    b.Send(m.Sender, "Bienvenido al megabot", menu)
    log.Printf("id: %v", m.Chat.ID)
  })


  b.Handle(&btnCompañia, func(m *tb.Message){
    if m.Chat.ID != userId{
      return
    }
    b.Send(m.Sender, "Envie el archivo de Reporte de compañia")
    b.Delete(m)
  })

	b.Start()
}
