package pagomono

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/lucianobenjota/go-oss-bot/m/pkg/monotributista"
	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
)

const (
	seleniumPath = "/home/luciano/go/pkg/mod/github.com/tebeka/selenium@v0.9.9/vendor/selenium-server.jar"
	chromeDriver = "/home/luciano/go/pkg/mod/github.com/tebeka/selenium@v0.9.9/vendor/chromedriver"
	port         = 8080
)

type Scrap struct {
	servicio *selenium.Service
	driver selenium.WebDriver
	Estado string
}

var urlSSS string = "https://www.sssalud.gob.ar/index.php?cat=consultas&page=mono_pagos"

// Inicia el servicio de scraping
func (s *Scrap) NuevoServicio() *selenium.Service {
	opts := []selenium.ServiceOption{
		selenium.StartFrameBuffer(),
		selenium.ChromeDriver(chromeDriver),
		//selenium.Output(os.Stderr),
	}	
	selenium.SetDebug(false)
	service, err := selenium.NewSeleniumService(seleniumPath, port, opts...)
	if err != nil {
		fmt.Println("error al iniciar el servicio de chromedriver: ", err.Error())
		log.Panicln(err)
	}
	s.servicio = service
	s.Estado = "iniciado"
	return service
}

func (s *Scrap) FinalizarScrapping() {
	log.Println("Deteniendo el servicio de scrapping")
	s.driver.Close()
	s.servicio.Stop()
	s.Estado = "finalizado"
}

func (s *Scrap) IniciarDriver() selenium.WebDriver {
	c := chrome.Capabilities{Path: "/usr/bin/google-chrome-stable"}
	caps := selenium.Capabilities{"browserName": "chrome"}
	caps.AddChrome(c)

	url := fmt.Sprintf("http://localhost:%d/wd/hub", port)

	driver, err := selenium.NewRemote(caps, url)
	if err != nil {
		log.Panicln(err)
	}
	s.driver = driver
	return driver
}

func (s Scrap) NavegarASSS() {
	err := s.driver.Get(urlSSS) 
	
	if err != nil {
		log.Panicln("Error de navegación: ", err)
	}

	titulo, err := s.driver.Title()
	if err != nil {
		log.Panicln("Error de navegación: ", err)
	}

	log.Println("El driver se encuentra en la página ", titulo)
}

// Formatea el CUIT
func FormatoCuit(cuit string) (resCuit string) {

	if len(cuit) == 11 {
		cuit = strings.Trim(cuit, " ")
		
		a := cuit[0:2]
		b := cuit[2:10]
		c := cuit[10:]
		resCuit = a + "-" + b + "-" + c
	}

	return resCuit
}

func PrettyPrint(v interface{}) (err error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err == nil {
					fmt.Println(string(b))
	}
	return
}

func checkErr(err error) {
	if err != nil {
		log.Panicln(err)
	}
}

func connectDB() *sql.DB {
	db, err := sql.Open("sqlite3", "./monotributos.db")
	checkErr(err)
	return db
}

func GeneradorPago(rows [][]string, cuit string) (err error) {
	const periodoLayout = "200601"
	const fechaLayout = "02-01-2006"
	db := connectDB()
	for _, row := range rows {
		var esUnico bool
		v := monotributista.Pago{}
		v.Cuit = cuit
		v.Periodo, _ = time.Parse(periodoLayout, row[0])
		v.Fecha, _ = time.Parse(fechaLayout, row[1])
		v.Concepto = strings.TrimSpace(row[2])
		v.Nro_secuencia = row[3]
		v.Credito = row[4]
		v.Debito = row[5]
		v.Rnos = row[6]
		// Verificamos que el pago no exista en la base
		// para insertar uno nuevo
		esUnico, err = monotributista.VerificarPago(db, v)
		if esUnico {
			err = monotributista.RegistrarPago(db, v)
		}
	}

	defer db.Close()
	if err != nil {
		return err
	}
	return nil
}

// Extractor de tablas
func ExtraerYRegistrarPago(fuente string, cuit string) (err error){
	var headings, row []string
	var rows [][]string
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(fuente))
	if err != nil {
		return err
	}
	doc.Find("table").Each(func(index int, tablehtml *goquery.Selection) {
		tablehtml.Find("tr").Each(func(indextr int, rowhtml *goquery.Selection) {
			rowhtml.Find("th").Each(func(indexth int, tableheading *goquery.Selection) {
				headings = append(headings, tableheading.Text())
			})
			rowhtml.Find("td").Each(func(indexth int, tablecell *goquery.Selection) {
				row = append(row, tablecell.Text())
			})
			if row != nil {
				if len(row) == 7{
					rows = append(rows, row)
				}
			}
			row = nil
		})
	})

	err = GeneradorPago(rows, cuit)
	if err != nil {
		return err
	}
	return
}

// Verifica que el driver se encuentre en la pag de pagos de mono
func (s Scrap) EsSuper() bool{
	curr_tit, err := s.driver.Title()
	if err != nil {
		log.Panicln(err)
	}
	return curr_tit == "Superintendencia de Servicios de Salud"
}

// Obtiene la imagen del captcha desde el webdriver
func (s Scrap) ObtenerCaptcha() ([]byte) {
	log.Printf("Obteniendo captcha..")
	element, err := s.driver.FindElement(selenium.ByID, "siimage")
	if err != nil {
		log.Panicln("Error, no se encontro el captcha: ", err)
	}
	imgByte, err := element.Screenshot(true)
	if err != nil {
		log.Panicln("Error al capturar captcha")
	}
	log.Printf("Captcha recuperado, enviando a tg")
	return imgByte
}

// Rellenar el campo de captcha
func (s Scrap) RellenarCaptcha(captcha string) (){
	element, err := s.driver.FindElement(selenium.ByName, "code")
	if err != nil {
		log.Panicln("no se encontro el campo para responder captcha, reintentar")
	}
	err = element.SendKeys(captcha)
	if err != nil {
		log.Panicln("error: ", err)
	}
}

// Rellenar el campo de cuit con el valor
func (s Scrap) RellenarCUIT(cuit string) () {
	element, err := s.driver.FindElement(selenium.ByName, "nro_cuil")
	if err != nil {
		log.Panicln("no se encontro el campo para responder al cuit, reintentar")
	}
	err = element.SendKeys(cuit)
	if err != nil {
		log.Panicln("error: ", err)
	}
}

// Click en el boton buscar
func (s Scrap) SubmitPagina() {
	element, err := s.driver.FindElement(selenium.ByName, "buscar")
	if err != nil {
		log.Panicln("No esta el boton buscar")
	}
	if err := element.Click(); err != nil {
		log.Panicln(err)
	}
}

func (s Scrap) ObtenerFuente() (fuente string) {
	fuente, err := s.driver.PageSource()
	if err != nil {
		log.Panicln(err)
	}
	return fuente
}