package bot

import (
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/lucianobenjota/go-oss-bot/m/pkg/compania"
	"github.com/lucianobenjota/go-oss-bot/m/pkg/convertidor"
	"github.com/lucianobenjota/go-oss-bot/m/pkg/descargas"
	"github.com/lucianobenjota/go-oss-bot/m/pkg/novedad"
	"github.com/lucianobenjota/go-oss-bot/m/pkg/pagomono"
	"github.com/lucianobenjota/go-oss-bot/m/pkg/procesonovedad"
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
		menu             = &tb.ReplyMarkup{ResizeReplyKeyboard: true, OneTimeKeyboard: true}
		btnCompañia      = menu.Text("🏢 Compañia")
		btnAyuda         = menu.Text("⚙ Ayuda")
		btnScrap         = menu.Text("🤖 Pagos Monotributo")
		btnNovedades     = menu.Text("🙌 Generar Novedades")
		btnProcNovedades = menu.Text("👁 Procesar Novedad")
	)

	menu.Reply(
		menu.Row(btnCompañia, btnScrap),
		menu.Row(btnNovedades, btnProcNovedades),
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
		modo = "compañia"
		b.Delete(m)
		b.Send(m.Sender, "Envie el archivo de Reporte de compañia")
	})

	b.Handle(&btnScrap, func(m *tb.Message) {
		if m.Chat.ID != tgUserId {
			return
		}
		modo = "pagomonotributo"
		b.Delete(m)
		b.Send(m.Sender, "Iniciando scrap de monotributo")
		err := pagomono.IniciarScrap()
		if err != nil {
			log.Panicln(err)
		}
	})

	b.Handle(&btnNovedades, func(m *tb.Message) {
		if m.Chat.ID != tgUserId {
			return
		}
		modo = "novedades"
		b.Send(m.Sender, "Enviar un archivo de reporte con las novedades")
	})

	b.Handle(&btnProcNovedades, func(m *tb.Message) {
		if m.Chat.ID != tgUserId {
			return
		}
		modo = "procnov"
		b.Send(m.Sender, "Enviar archivo de novedades erroneas del FTP")
	})

	b.Handle(tb.OnDocument, func(m *tb.Message) {
		if modo == "compañia" {
			b.Send(m.Sender, "🤖 Modo compañia")
			destFolder := os.Getenv("PROCESS_FOLDER")
			filename := destFolder + m.Document.FileName
			csvfilename := destFolder + descargas.FileNameWithoutExt(m.Document.FileName) + ".csv"
			d := &descargas.Download{Bot: *b, Msg: *m}
			d.DescargarArchivo(filename)
			err := compania.ReporteACSV(filename, csvfilename)
			if err != nil {
				log.Println("Error al recibir reporte")
				log.Panic(err)
			}

			resDoc := &tb.Document{
				File:     tb.FromDisk(csvfilename),
				FileName: csvfilename,
				MIME:     "text/csv",
			}
			b.Send(m.Sender, resDoc)
		}
		if modo == "novedades" {
			b.Send(m.Sender, "🤖 Modo novedades")

			destFolder := os.Getenv("PROCESS_FOLDER")

			filename := destFolder + m.Document.FileName
			csvfilename := destFolder + descargas.FileNameWithoutExt(m.Document.FileName) + ".csv"

			d := &descargas.Download{Bot: *b, Msg: *m}
			d.DescargarArchivo(filename)

			err := convertidor.CmdWrapper(filename, csvfilename)
			if err != nil {
				log.Panicln(err)
			}

			csvF, err := os.Open(csvfilename)
			if err != nil {
				log.Panicln(err)
			}

			defer csvF.Close()

			filedest := destFolder + "novedad.csv"
			err = os.Remove(filedest)
			if err != nil {
				log.Println("No existe el archivo, procediendo")
			}

			err = novedad.CSVANovedad(csvF, filedest)
			if err != nil {
				log.Panicln(err)
			}

			resDoc := &tb.Document{
				File:     tb.FromDisk(filedest),
				FileName: filedest,
				MIME:     "text/csv",
			}

			b.Send(m.Sender, resDoc)
		}

		if modo == "procnov" {
			b.Send(m.Sender, "🤖 Modo de proceso de novedades")

			destFolder := os.Getenv("PROCESS_FOLDER")

			filename := destFolder + m.Document.FileName

			d := &descargas.Download{Bot: *b, Msg: *m}
			d.DescargarArchivo(filename)
			tsvFile, err := os.Open(filename)
			if err != nil {
				log.Panicln("Error al abrir el TSV: ", err)
			}
			defer tsvFile.Close()
			msga, err := b.Send(m.Sender, "Procesando archivo...")
			data, err := procesonovedad.LeerCSVFTP(tsvFile)

			if err != nil {
				b.Send(m.Sender, "Ocurrio un error al procesar la novedad")
				log.Println("Error al procesar la novedad: ", err)
			}
			msgb, err := b.Send(m.Sender, "Generando respuesta.. ")

			// Creamos un archivo temporal para enviar..
			temporaryFile, err := ioutil.TempFile("./", "*.csv")
			if err != nil {
				log.Panicln("Error al crear archivo temporal: ", err)
				b.Send(m.Sender, "Error al procesar novedad..")
			}
			defer os.Remove(temporaryFile.Name())

			if _, err := temporaryFile.Write(data); err != nil {
				log.Panicln("Error al grabar datos en el temporal: ", err)
				b.Send(m.Sender, "Error al procesar la novedad")
				temporaryFile.Close()
			}

			res := &tb.Document{
				File:     tb.FromDisk(temporaryFile.Name()),
				FileName: "novedad.csv",
				Caption:  "Archivo procesado",
				MIME:     "text/csv",
			}

			_, err = b.Send(m.Sender, res)
			if err != nil {
				log.Fatal(err)
				temporaryFile.Close()
			}
			temporaryFile.Close()
			b.Delete(msga)
			b.Delete(msgb)

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
