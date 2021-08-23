package bot

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/lucianobenjota/go-oss-bot/m/pkg/compania"
	tb "gopkg.in/tucnak/telebot.v2"
)

var modo string

func StartBot() (err error){
  err = godotenv.Load()

  if err != nil {
    log.Fatal("Error de configuración")
  }

  tgApiKey := os.Getenv("TELEGRAM_API_KEY")
  tgUserId, err := strconv.ParseInt(os.Getenv("TELEGRAM_USER_ID"), 10, 64)

  if err != nil {
    log.Fatal("Error de configuración")
  }

  b, err := tb.NewBot(tb.Settings{
    Token : tgApiKey,
    Poller: &tb.LongPoller{Timeout: 10 * time.Second},
  })

  if err != nil {
    log.Fatal(err)
  }

  var (
    menu = &tb.ReplyMarkup{ResizeReplyKeyboard: true, OneTimeKeyboard: true}
    btnCompañia = menu.Text("🏢 Compañia")
    btnAyuda = menu.Text("⚙ Ayuda")
  )

  menu.Reply(
    menu.Row(btnCompañia),
    menu.Row(btnAyuda),
  )

  b.Handle("/start", func(m *tb.Message){

    if m.Chat.ID != tgUserId{
      log.Printf("El id %v envió un mensaje", m.Chat.ID)
      return
    }

    b.Send(m.Sender, "Bienvenido al megabot", menu)
    log.Printf("id: %v", m.Chat.ID)
  })


  b.Handle(&btnCompañia, func(m *tb.Message){
    if m.Chat.ID != tgUserId{
      return
    }
    modo = "Compañia"
    b.Delete(m)
    b.Send(m.Sender, "Envie el archivo de Reporte de compañia")
  })

  b.Handle(tb.OnDocument, func(m *tb.Message){
    if modo == "Compañia"{
      log.Println("Archivo recivido")
      b.Download(&m.Document.File, "./file.xls")
      compania.XLSaCSV("./file.xls", "./file.csv")
      b.Send(m.Sender, "Proceso finalizao")
    }

  })

  b.Start()
  return err
}
