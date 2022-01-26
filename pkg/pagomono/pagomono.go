package pagomono

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/PuerkitoBio/goquery"
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

type PagoSSS struct {
	periodo string
	fecha_pago string
	concepto string
	nro_secuencia string
	credito string
	debito string 
	rnos string
}

func GeneradorPago(rows [][]string) {
	var pagos []PagoSSS
	for _, row := range rows {
		v := PagoSSS{}
		v.periodo = row[0]
		v.fecha_pago = row[1]
		v.concepto = strings.TrimSpace(row[2])
		v.nro_secuencia = row[3]
		v.credito = row[4]
		v.debito = row[5]
		v.rnos = row[6]
		pagos = append(pagos, v)
	}
	log.Println(pagos)
}

// Extractor de tablas
func ExtractorTablas(fuente string) {
	var headings, row []string
	var rows [][]string
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(fuente))
	if err != nil {
		log.Panicln("Error al extraer las tablas:", err)
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
	log.Println("Headings:")
	PrettyPrint(headings)
	log.Println("Rows:")
	PrettyPrint(rows)
	
	GeneradorPago(rows)
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