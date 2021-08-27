package bot

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/lucianobenjota/go-oss-bot/m/pkg/descargas"
	tb "gopkg.in/tucnak/telebot.v2"
)

var modo string
var tgApiKey string
var tgUserId string

func StartBot() (err error) {
	// Inicia el Bot de telegram
	tg, err := getTgApiKey()
	if err != nil {
		log.Panic(err)
	}

	tgApiKey := tg["api"]
	tgUserId, err := strconv.ParseInt(tg["user"], 10, 64)

	if err != nil {
		log.Fatal("Error de configuración")
	}

	b, err := tb.NewBot(tb.Settings{
		Token:  tgApiKey,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})

	if err != nil {
		log.Fatal(err)
	}

	var (
		menu        = &tb.ReplyMarkup{ResizeReplyKeyboard: true, OneTimeKeyboard: true}
		btnCompañia = menu.Text("🏢 Compañia")
		btnAyuda    = menu.Text("⚙ Ayuda")
	)

	menu.Reply(
		menu.Row(btnCompañia),
		menu.Row(btnAyuda),
	)

	b.Handle("/start", func(m *tb.Message) {

		if m.Chat.ID != tgUserId {
			log.Printf("El id %v envió un mensaje", m.Chat.ID)
			return
		}

		b.Send(m.Sender, "Bienvenido al megabot", menu)
		log.Printf("id: %v", m.Chat.ID)
	})

	b.Handle(&btnCompañia, func(m *tb.Message) {
		if m.Chat.ID != tgUserId {
			return
		}
		modo = "Compañia"
		b.Delete(m)
		b.Send(m.Sender, "Envie el archivo de Reporte de compañia")
	})

	b.Handle(tb.OnDocument, func(m *tb.Message) {
		if modo == "Compañia" {
			log.Println("Archivo recivido")
			d := &descargas.Download{Bot: *b, Msg: *m}
			d.DescargarArchivo("chivo.xls")
			// b.Download(&m.Document.File, "./file.xls")
			// err := compania.ReporteACSV("file.xls", "file.csv")
			// if err != nil {
			// 	log.Println("Error al recibir reporte")
			// 	log.Panic(err)
			// }
			// file, err := os.Open("file.csv")
			// if err != nil {
			// 	log.Panic(err)
			// }
			// b.Send(m.Sender, "Proceso finalizao")
			// b.Send(m.Sender, file)
		}

	})

	b.Start()
	return err
}

func getTgApiKey() (res map[string]string, err error) {
	// Obtiene las credenciales del archivo de enotrno
	err = godotenv.Load()
	if err != nil {
		return nil, err
	}

	var res2 = make(map[string]string)
	res2["api"] = os.Getenv("TELEGRAM_API_KEY")
	res2["user"] = os.Getenv("TELEGRAM_USER_ID")
	return res2, nil
}

func doEvery(d time.Duration, f func(time.Time)) {
	// Ejecuta una funcion cada d time.*
	for x := range time.Tick(d) {
		f(x)
	}
}
