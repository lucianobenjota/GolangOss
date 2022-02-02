package bot

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/lucianobenjota/go-oss-bot/m/pkg/afiliaciones"
	"github.com/lucianobenjota/go-oss-bot/m/pkg/compania"
	"github.com/lucianobenjota/go-oss-bot/m/pkg/convertidor"
	"github.com/lucianobenjota/go-oss-bot/m/pkg/descargas"
	"github.com/lucianobenjota/go-oss-bot/m/pkg/modoapp"
	"github.com/lucianobenjota/go-oss-bot/m/pkg/novedad"
	"github.com/lucianobenjota/go-oss-bot/m/pkg/pagomono"
	"github.com/lucianobenjota/go-oss-bot/m/pkg/procesonovedad"
	"github.com/lucianobenjota/go-oss-bot/m/pkg/webdriver"
	tb "gopkg.in/tucnak/telebot.v2"
)

var modo modoapp.ModoApp

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
		btnMonotributos         = menu.Text("🙈 Monotributistas")
		btnNovedades     = menu.Text("🙌 Generar Novedades")
		btnProcNovedades = menu.Text("👁 Procesar Novedad")
		btnAfiliaciones  = menu.Text("👥 Afiliaciónes")
	)

	menu.Reply(
		menu.Row(btnCompañia, btnMonotributos),
		menu.Row(btnNovedades, btnProcNovedades),
		menu.Row(btnAfiliaciones),
		menu.Row(btnAyuda),
	)

	// Comando para iniciar el bot
	b.Handle("/start", func(m *tb.Message) {
		if m.Chat.ID != tgUserId {
			log.Printf("El usuario %v inicio el bot", m.Chat.ID)
		}
		b.Send(m.Sender, "Bienvenido al megabot", menu)
	})

	b.Handle(&btnCompañia, func(m *tb.Message) {
		if m.Chat.ID != tgUserId {
			return
		}
		modo = modo.Set(modoapp.Compania)
		b.Delete(m)
		b.Send(m.Sender, "Envíe el archivo de reporte de compañia")
	})

	b.Handle(&btnAfiliaciones, func(m *tb.Message) {
		if m.Chat.ID != tgUserId {
			return
		}
		log.Println("Solicitar padron de afiliados en formato MICAM")

		b.Send(m.Sender, "Envie el archivo de padron de afiliados")
		modo = modo.Set(modoapp.Afiliaciones)
	})

	var (
		menumonos = &tb.ReplyMarkup{ResizeReplyKeyboard: true, OneTimeKeyboard: true, ForceReply: true}
		btnListaMonos = menumonos.Text("🐵 Lista de monotributos")
		btnGenerarPago = menumonos.Text("💰 Generar pago")
		btnFinalizarProc = menumonos.Text("💁 Finalizar proceso")
		btnBuscarMono = menumonos.Text("🔎 Consultar monotributo")
	)

	menumonos.Reply(
		menumonos.Row(btnBuscarMono, btnListaMonos),
		menumonos.Row(btnGenerarPago, btnFinalizarProc), 
	)



	b.Handle(&btnBuscarMono, func(m *tb.Message) {
		if m.Chat.ID != tgUserId {
			return
		}

		modo = modo.Set("consultamono")
		modo = modo.Set(modoapp.ConsultarMonotributo)

		b.Delete(m)
		b.Send(m.Sender, "Envia el CUIT del monotributista")
		
	})

	b.Handle(&btnMonotributos, func(m *tb.Message) {
		if m.Chat.ID != tgUserId {
			return
		}
		modo = modo.Set(modoapp.PagoMonotributo)
		b.Delete(m)
		b.Send(m.Sender, "Selecciona una opcion para los monotributos", menumonos)
	})

	b.Handle(&btnNovedades, func(m *tb.Message) {
		if m.Chat.ID != tgUserId {
			return
		}
		modo = modo.Set(modoapp.Novedades)
		b.Send(m.Sender, "Enviar un archivo de reporte con las novedades")
	})

	b.Handle(&btnProcNovedades, func(m *tb.Message) {
		if m.Chat.ID != tgUserId {
			return
		}
		modo = modo.Set(modoapp.ProcesarNovedades)
		b.Send(m.Sender, "Enviar archivo de novedades erroneas del FTP")
	})

	scrap := webdriver.Scrap{Estado: "idle"}
	var (
		cuit, cuitOriginal, captcha string
	)

	b.Handle(&btnFinalizarProc, func(m *tb.Message) {
		if m.Chat.ID != tgUserId {
			return
		}
		b.Send(m.Sender, "Finalizando proceso..")
		modo = modo.Set(modoapp.FinalizarWebDriver)
	})

	b.Handle(&btnGenerarPago, func(m *tb.Message) {
		if m.Chat.ID != tgUserId {
			return
		}
		modo = modo.Set(modoapp.GenerarPagos)
		b.Delete(m)
		b.Send(m.Sender, "CUIT a generar?")
	})

	// Maneja los textos enviados al bot que no sean los botones
	b.Handle(tb.OnText, func(m *tb.Message) {
		if m.Chat.ID != tgUserId {
			return
		}

		if modo.EsConsultaMono() {
			log.Println("consultamono: ", m.Text)
			cuit = m.Text
			b.Send(m.Sender, "consultando cuit "+ cuit)
		}

		if modo.EsFinalizarWebDriver() {
			scrap.FinalizarScrapping()
			cuit = ""
			captcha = ""
			modo.Set(modoapp.FinalizarWebDriver)
		}

		if modo.EsGenerarPagos() {
			if len(cuit) == 0 {
				cuit, err = pagomono.FormatoCuit(m.Text)
				if err != nil {
					cuit = ""
					captcha = ""
					modo.Set("start")
					b.Send(m.Sender, "Formato de cuit incorrecto")
				}
				cuitOriginal = m.Text
			}

			log.Println("Iniciando servidor de webdriver")

			if scrap.Estado != "iniciado"{
				b.Send(m.Sender, "Iniciando driver..")
				scrap.NuevoServicio()
				log.Println("Iniciando driver")
				scrap.IniciarDriver()
				log.Println("Navegando a ssssalud")
				b.Send(m.Sender, "Driver iniciado correctamente, rescatando captcha..")
			}

			// Url de superintendencia
			var urlSSS string = "https://www.sssalud.gob.ar/index.php?cat=consultas&page=mono_pagos"

			if !scrap.EsWeb("Superintendencia de Servicios de Salud") {
				scrap.NavegarA(urlSSS)	
			}

			log.Println("captcha:", captcha)
			log.Println("cuit:", cuit)
			log.Println(m.Text)

			if len(captcha) == 0 {
				if m.Text != cuitOriginal {
					captcha = m.Text
				} else {
					img := pagomono.ObtenerCaptcha(scrap)
					p := &tb.Photo{File: tb.FromReader(bytes.NewReader(img))}
					b.Send(m.Sender, p)
					b.Send(m.Sender, "Captcha?")
				}
			}

			if len(captcha) > 0 && len(cuit) > 0{
				log.Println("Submit de página")
				log.Println("captcha:", captcha, "largo:", len(captcha))
				log.Println("cuit:", cuit, "largo:", len(cuit))
				pagomono.RellenarCUIT(scrap, cuit)
				pagomono.RellenarCaptcha(scrap, captcha)
				pagomono.SubmitPagina(scrap)
				fuente := scrap.ObtenerFuente()

				if strings.Contains(fuente, "AAAAAA") {
					b.Send(m.Sender, "Registrando pagos..")
					cantidad, err := pagomono.ExtraerYRegistrarPago(fuente, cuit)
					if err != nil {
						b.Send(m.Sender, "Ocurrio un error al registrar el pago")
						log.Panicln(err)
					}
					if cantidad == 0 {
						b.Send(m.Sender, "No se encontraron nuevos pagos para el monotributo")
					} else {
						mensaje := fmt.Sprintf("Se registraron %d pagos", cantidad)
						b.Send(m.Sender, mensaje)
					}
				} else if strings.Contains(fuente, "CCCCCC") {
					b.Send(m.Sender, "Captcha incorrecto")
				}

				cuit = ""
				captcha = ""
				modo = modo.Set(modoapp.EsperandoWebdriver)
			}
		}

		if modo.EsEsperandoWebdriver() {
			if m.Text == "si" {
				modo = modo.Set(modoapp.GenerarPagos)
				b.Send(m.Sender, "CUIT a generar?")
			}
			if m.Text == "no" {
				scrap.FinalizarScrapping()
				modo = modo.Set(modoapp.FinalizarWebDriver)
			}
			if m.Text != "Nuevo pago" {
				b.Send(m.Sender, "Nuevo pago?")
			}
			log.Println("resp: ", m.Text)
		}

	})

// Maneja los archivos enviados al bot
	b.Handle(tb.OnDocument, func(m *tb.Message) {
		if modo.EsCompania() {
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
		if modo.EsNovedaes() {
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

		if modo.EsProcesarNovedades() {
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
			msga, _ := b.Send(m.Sender, "Procesando archivo...")
			data, err := procesonovedad.LeerCSVFTP(tsvFile)

			if err != nil {
				b.Send(m.Sender, "Ocurrio un error al procesar la novedad")
				log.Println("Error al procesar la novedad: ", err)
			}
			msgb, _ := b.Send(m.Sender, "Generando respuesta.. ")

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

		if modo.EsAfiliaciones() {
			log.Println("Modo afiliaciones")
			b.Send(m.Sender, "Verificando afiliaciones..")
			destFolder := os.Getenv("PROCESS_FOLDER")
			filename := destFolder + m.Document.FileName

			d := &descargas.Download{Bot: *b, Msg: *m}
			d.DescargarArchivo(filename)

			afiliaciones.ProcesarAfiliaciones(filename, *b, m)

			// if err != nil {
			// 	b.Send(m.Sender, "Ocurrio un error al procesar la novedad")
			// 	log.Println("Error al procesar la novedad: ", err)
			// }
			// msgb, err := b.Send(m.Sender, "Generando respuesta.. ")

			// // Creamos un archivo temporal para enviar..
			// temporaryFile, err := ioutil.TempFile("./", "*.csv")
			// if err != nil {
			// 	log.Panicln("Error al crear archivo temporal: ", err)
			// 	b.Send(m.Sender, "Error al procesar novedad..")
			// }
			// defer os.Remove(temporaryFile.Name())

			// if _, err := temporaryFile.Write(data); err != nil {
			// 	log.Panicln("Error al grabar datos en el temporal: ", err)
			// 	b.Send(m.Sender, "Error al procesar la novedad")
			// 	temporaryFile.Close()
			// }

			// res := &tb.Document{
			// 	File:     tb.FromDisk(temporaryFile.Name()),
			// 	FileName: "novedad.csv",
			// 	Caption:  "Archivo procesado",
			// 	MIME:     "text/csv",
			// }

			// _, err = b.Send(m.Sender, res)
			// if err != nil {
			// 	log.Fatal(err)
			// 	temporaryFile.Close()
			// }
			// temporaryFile.Close()
			// b.Delete(msga)
			// b.Delete(msgb)

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

// func doEvery(d time.Duration, f func(time.Time)) {
// 	// Ejecuta una funcion cada d time.*
// 	for x := range time.Tick(d) {
// 		f(x)
// 	}
// }
