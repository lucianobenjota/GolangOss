package pagomono

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/lucianobenjota/go-oss-bot/m/pkg/monotributista"
	"github.com/lucianobenjota/go-oss-bot/m/pkg/webdriver"
	"github.com/tebeka/selenium"
)

// Formatea el CUIT
func FormatoCuit(cuit string) (resCuit string, err error) {

	if len(cuit) == 13 {
		return cuit, nil
	}

	if len(cuit) == 11 {
		cuit = strings.Trim(cuit, " ")
		
		a := cuit[0:2]
		b := cuit[2:10]
		c := cuit[10:]
		resCuit = a + "-" + b + "-" + c
		return resCuit, nil
	}
	err = errors.New("CUIT con formato incorrecto")
	return "", err
}

// Imprime las variables como tablas json
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

func ConnectDB() *sql.DB {
	db, err := sql.Open("sqlite3", "./monotributos.db")
	checkErr(err)
	return db
}

func GeneradorPago(rows [][]string, cuit string) (cantidadPagos int, err error) {
	const periodoLayout = "200601"
	const fechaLayout = "02-01-2006"
	db := ConnectDB()
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
			cantidadPagos++
		}
	}

	defer db.Close()
	if err != nil {
		return 0, err
	}
	return cantidadPagos, nil
}

// Extrae los datos de la fuente y registra el pago, devuelve la cantidad de
// pagos registrados
func ExtraerYRegistrarPago(fuente string, cuit string) (cantidad int, err error){
	var headings, row []string
	var rows [][]string
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(fuente))
	if err != nil {
		return 0, err
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

	cantidad, err = GeneradorPago(rows, cuit)
	if err != nil {
		return 0, err
	}
	return

}


// Obtiene la imagen del captcha desde el webdriver
func ObtenerCaptcha(s webdriver.Scrap) ([]byte) {
	log.Printf("Obteniendo captcha..")
	element, err := s.Driver.FindElement(selenium.ByID, "siimage")
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
func RellenarCaptcha(s webdriver.Scrap, captcha string) (){
	element, err := s.Driver.FindElement(selenium.ByName, "code")
	if err != nil {
		log.Panicln("no se encontro el campo para responder captcha, reintentar")
	}
	err = element.SendKeys(captcha)
	if err != nil {
		log.Panicln("error: ", err)
	}
}

// Rellenar el campo de cuit con el valor
func RellenarCUIT(s webdriver.Scrap, cuit string) () {
	element, err := s.Driver.FindElement(selenium.ByName, "nro_cuil")
	if err != nil {
		log.Panicln("no se encontro el campo para responder al cuit, reintentar")
	}
	err = element.SendKeys(cuit)
	if err != nil {
		log.Panicln("error: ", err)
	}
}

// Click en el boton buscar
func SubmitPagina(s webdriver.Scrap) {
	element, err := s.Driver.FindElement(selenium.ByName, "buscar")
	if err != nil {
		log.Panicln("No esta el boton buscar")
	}
	if err := element.Click(); err != nil {
		log.Panicln(err)
	}
}
