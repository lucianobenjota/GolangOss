package afiliaciones

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocarina/gocsv"
	"github.com/joho/sqltocsv"
	"github.com/lucianobenjota/go-oss-bot/m/pkg/convertidor"
	"github.com/lucianobenjota/go-oss-bot/m/pkg/pagomono"
	"github.com/lucianobenjota/go-oss-bot/m/pkg/webdriver"
	"github.com/tebeka/selenium"
	tb "gopkg.in/tucnak/telebot.v2"
)

// Estructura para guardar los dni en modo csv
type PadronDNI struct {
	DNI            string `csv:"Nro. Doc."`
}

// Procesar afiliaciones
func ProcesarAfiliaciones(rutaReporte string, bot tb.Bot, msg *tb.Message) {
	err := ReporteMicamCSV(rutaReporte)
	if err != nil {
		log.Panicln("Error al procesar reportes")
	}
	
	padronDNIs, err := ExtraeDNIs()
	if err != nil {
		log.Panicln("Error al procesar DNIs")
	}

	mensaje, _ := bot.Send(msg.Sender, "Iniciando driver..")
	
	err = IniciarDriver()
	if err != nil {
		bot.Edit(mensaje, "Error al iniciar el driver")
		log.Panicln(err)
	}

	err = LoguearSuper()
	if err != nil {
		bot.Edit(mensaje, "Error al loguearse en la super")
		log.Panicln("Error al loguearse en superintendencia")
	}

	bot.Edit(mensaje, "Webdriver iniciado, iniciando scrapping")
	db := pagomono.ConnectDB()

	err = LimpiarDatosAfiliado(db)
	if err != nil {
		log.Panicln(err)
	}

	restante := len(padronDNIs)
	for _, d := range padronDNIs {
		dni := d.DNI
		bot.Edit(mensaje, fmt.Sprintf("verificando dni %s, restantes %d", dni, restante))
		f, err := AfiliacionWeb(dni)
		if err != nil {
			bot.Edit(mensaje, fmt.Sprintf("error al scrapear datos de dni %s", dni))
			log.Panicln(err)
		}
		grupo, err := ExtraeGrupoFamiliarWeb(f, dni)
		if err != nil {
			bot.Edit(mensaje, fmt.Sprintf("error al extraer datos de la pagina para el dni %s", dni))
			log.Panicln(err)
		}
		err = InsertarDatosAfiliado(db, grupo)
		if err != nil {
			bot.Edit(mensaje, fmt.Sprintf("error al insertar en la base para el dni %s", dni))
			log.Panicln(err)
		}

		log.Println(fmt.Sprintf("scrap del dni %s correcto, restantes %d", dni, restante))
		restante = restante - 1
	}
	
	// Finalizar la instancia del scrap
	scrap.FinalizarScrapping()
	bot.Edit(mensaje, "Instancia de webdriver finalizada")
	err = ExportarAfiliacion(db)

	res := &tb.Document{
		File:     tb.FromDisk("./salida.csv"),
		FileName: "afiliaciones.csv",
		MIME:     "text/csv",
	}

	bot.Send(msg.Sender, res)

	if err != nil {
		log.Panicln("Error de afiliacion")
	}

	defer db.Close()	

	err = limpiarTemporal(archivoTemporal)
	
	if err != nil {
		log.Panicln("Error al limpiar archivo temporal")
	}

	if err = limpiarTemporal("./salida.csv"); err != nil {
		log.Panicln("Error al limpiar el archivo de salida")
	}

	bot.Delete(mensaje)
	bot.Send(msg.Sender, "Proceso finalizado")
}

const archivoTemporal = "temp_afiliaciones.csv"

// Convierte a CSV el reporte de afiliados
func ReporteMicamCSV(reporte string) error {
	err := convertidor.CmdWrapper(reporte, archivoTemporal)
	if err != nil {
		return err
	}
	return nil
}

// Extrae la columna de DNI del reporte y la devuelve como un array de punteros
func ExtraeDNIs() (padron []*PadronDNI, err error){
	archivoCSV, err := os.Open(archivoTemporal)
	if err != nil {
		return []*PadronDNI{}, err
	}

	p := []*PadronDNI{}

	if err := gocsv.UnmarshalFile(archivoCSV, &p); err != nil {
		return []*PadronDNI{}, err
	}

	defer archivoCSV.Close()

	return p, nil
}

var scrap = webdriver.Scrap{}
var wb selenium.WebDriver
// Inicia el webdriver
func IniciarDriver() error{
	_, err := scrap.NuevoServicio()

	if err != nil {
		return err
	}

	wb, err = scrap.IniciarDriver()
	if err != nil {
		return err
	}

	scrap.Estado = "nologueado"
	return nil
}

// URL de la pagina de login de superintendencia
var urlLoginSuper string = "https://seguro.sssalud.gob.ar/login.php?b_publica=Acceso+Restringido+para+Obras+Sociales&opc=bus650&user=RNOS"

// Loguea el driver en superintendencia
func LoguearSuper() error {

	usuarioSSS := os.Getenv("USUARIO_SSS")
	passSSS := os.Getenv("PASS_SSS")

	err := wb.Get(urlLoginSuper)
	if err != nil {
		log.Panicln("Error al navegar a la página de login")
		return err
	}

	inputUsuario, err  := wb.FindElement(selenium.ByName, "_user_name_")
	
	if err != nil {
		log.Panicln("No se econtró el campo de usuario en la página de login")
		return err
	}

	inputPassword, err := wb.FindElement(selenium.ByName, "_pass_word_")
	if err != nil {
		log.Panicln("No se econtró el campo de password en la página de login")
		return err
	}

	inputUsuario.SendKeys(usuarioSSS)
	inputPassword.SendKeys(passSSS)
	
	forma, err := wb.FindElement(selenium.ByCSSSelector, ".formulario")
	
	if err != nil {
		log.Panicln("No se econtró la forma (clase incorrecta en login)")
		return err
	}
	
	err = forma.Submit()

	if err != nil {
		log.Panicln("Error al postear datos de login")
		return err
	}
	scrap.Estado = "logueado"
	return nil
}

// Extrae la fuente de la web de afiliacion del afiliado por dni
func AfiliacionWeb(dni string) (fuenteWeb string, err error) {
	// Enviamos los datos al webdriver
	inputCUIL, _ := wb.FindElement(selenium.ByName, "cuil_b")
	inputDNI, _ := wb.FindElement(selenium.ByName, "nro_doc")
	btnConsultar, _ := wb.FindElement(selenium.ByName, "B1")

	inputCUIL.SendKeys(dni)
	inputDNI.SendKeys(dni)
	btnConsultar.Click()

	fuenteWeb, err = wb.PageSource()
	if err != nil {
		return "", err
	}

	return fuenteWeb, nil
}

// Elimina los espacios y tabulaciones
func ReformatString(s string) (e string) {
	e = s
	e = strings.ReplaceAll(e, "\n", "")
	e = strings.ReplaceAll(e, "\t", "")
	e = strings.Trim(e, " ")
	return e
}


// Tipo afiliado web, campos de la tabla a extraer
type afWeb struct{
	dniScrap string
	parentesco string
	cuil string
	tipodni string
	nrodni string
	nombre string
	provincia string
	fecNac time.Time
	sexo string
	cuilTitular string
	cuitEmpleador string
	tipoTitular string
	codigoOS string
	denominacionOS string
	fechaAltaOS time.Time
	ultimoPeriodo time.Time
}

// Mapea los datos de la tabla personales en el struct de afiliado web
func mapPersonales(afiliado afWeb, valor string, campo string) (afWeb) {
	af := afiliado
	const fechaLayout = "02-01-2006"
	switch campo {
		case "Parentesco":
			af.parentesco = valor
		case "CUIL":
			af.cuil = valor
		case "Tipo de documento":
			af.tipodni = valor
		case "Número de documento":
			af.nrodni = valor
		case "Apellido y nombre":
			af.nombre = valor
		case "Provincia":
			af.provincia = valor
		case "Fecha de nacimiento":
			af.fecNac, _= time.Parse(fechaLayout, valor)
		case "Sexo":
			af.sexo = valor
	}
	return af
} 

// Mapea los datos de la tabla afiliaciones en el struct de afiliado web
func mapAfiliatorios(afiliado afWeb, valor string, campo string) (afWeb) {
 af := afiliado
 const fechaLayout = "02-01-2006"
 switch campo {
 	case "CUIL titular":
		af.cuilTitular = valor
	case "CUIT de empleador":
		af.cuitEmpleador = valor
	case "Tipo de beneficiario":
		af.tipoTitular = valor
	case "Código de Obra Social":
		af.codigoOS = valor
	case "Denominación Obra Social":
		af.denominacionOS = valor
	case "Fecha Alta Obra Social":
		af.fechaAltaOS, _ = time.Parse(fechaLayout, valor)
 }
 return af
}

func mapEmpleador(grupo []afWeb, valor string, campo string) []afWeb{
	const periodoLayout = "01-2006"
	for i := range grupo {
		switch campo {
			case "Ultimo Período Declarado":
				grupo[i].ultimoPeriodo, _ = time.Parse(periodoLayout, valor)
		}
	}
	return grupo
}

// Extrae las tablas en formato JSON dese la fuente de la página
func ExtraeGrupoFamiliarWeb(fuente string, dniScrap string) (grupoFamiliar []afWeb, err error) {

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(fuente))
	if err != nil {
		return nil, err
	}

	doc.Find("table[summary=\"Esta tabla muestra los datos personales\"]").Each(func(index int, tablapersonal *goquery.Selection) {
		af := afWeb{}
		af.dniScrap = dniScrap
		tablapersonal.Find("tr").Each(func(indextr int, rowhtml *goquery.Selection) {
			var campo string
			var valor string
			rowhtml.Find("th").Each(func(indexth int, tHead *goquery.Selection) {
				campo = ReformatString(tHead.Text())
			})
			rowhtml.Find("td").Each(func(indextd int, td *goquery.Selection) {
				valor = ReformatString(td.Text())
			})
			af = mapPersonales(af, valor, campo)
		})
		grupoFamiliar = append(grupoFamiliar, af)
	})

	doc.Find("table[summary=\"Esta tabla muestra los datos de afiliación\"]").Each(func(index int, tablaafiliacion *goquery.Selection) {
		var campo string
		var valor string
		afiliado := grupoFamiliar[index]

		tablaafiliacion.Find("tr").Each(func(indextr int, rowhtml *goquery.Selection) {
			rowhtml.Find("th").Each(func(indexth int, th *goquery.Selection) {
				campo = ReformatString(th.Text())
			})
			rowhtml.Find("td").Each(func(indextd int, td *goquery.Selection) {
				valor = ReformatString(td.Text())
			})
			afiliado = mapAfiliatorios(afiliado, valor, campo)
		})
		grupoFamiliar[index] = afiliado
	})

	doc.Find("table[summary=\"Esta tabla muestra los datos declarados por el empleador\"]").Each(func(i int, tablaempleador *goquery.Selection) {
		var campo string
		var valor string
		tablaempleador.Find("tr").Each(func(indextr int, rowhtml *goquery.Selection) {
			rowhtml.Find("th").Each(func(indexth int, th *goquery.Selection) {
				campo = ReformatString(th.Text())
			})
			rowhtml.Find("td").Each(func(indextd int, td *goquery.Selection) {
				valor = ReformatString(td.Text())
			})
			grupoFamiliar = mapEmpleador(grupoFamiliar, valor, campo)
		})

	})

	return grupoFamiliar, nil
}

// Inserta los datos en el padron de afiliados
func InsertarDatosAfiliado(db *sql.DB, grupoFamiliar []afWeb) error {
	for _, af := range grupoFamiliar {
		q := `
		INSERT INTO afiliaciones (
			dni_scrap,
			parentesco, 
			cuil, 
			tipo_dni, 
			nro_dni, 
			nombre, 
			provincia, 
			fec_nac, 
			sexo, 
			cuil_titular, 
			cuit_empleador, 
			tipo_titular, 
			codigo_os, 
			denominacion_os, 
			fecha_alta_os, 
			ultimo_periodo
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);
		`	
		stmt, _ := db.Prepare(q)
		_, err := stmt.Exec(
			af.dniScrap,
			af.parentesco, 
			af.cuil, 
			af.tipodni, 
			af.nrodni, 
			af.nombre, 
			af.provincia, 
			af.fecNac, 
			af.sexo, 
			af.cuilTitular, 
			af.cuitEmpleador, 
			af.tipoTitular, 
			af.codigoOS, 
			af.denominacionOS, 
			af.fechaAltaOS, 
			af.ultimoPeriodo,
		)		

		if err != nil {
			return err
		}

		defer stmt.Close()
	}
	return nil
}

// Limpiar la bse de afiliados
func LimpiarDatosAfiliado(db *sql.DB) error {
	q := `DELETE FROM afiliaciones;`
	_, err := db.Exec(q)
	if err != nil {
		return err
	}
	return nil
}

// Genera el archivo salida.csv con los datos del padrón
func ExportarAfiliacion(db *sql.DB) error {
	q := `
		SELECT 
			dni_scrap, 
			parentesco, 
			cuil, 
			tipo_dni, 
			nro_dni, 
			nombre, 
			provincia, 
			strftime("%Y/%m/%d",fec_nac) as fec_nac, 
			sexo, 
			cuil_titular, 
			cuit_empleador, 
			tipo_titular, 
			codigo_os, 
			denominacion_os, 
			strftime("%Y/%m/%d", fecha_alta_os) as fecha_alta_os, 
			strftime("%Y-%m", ultimo_periodo) as ultimo_periodo
		FROM afiliaciones;
	`
	rows, _ := db.Query(q)

	err := sqltocsv.WriteFile("./salida.csv", rows)
	if err != nil {
    return err
	}
	
	return nil
}

// Limpia el archivo temporal
func limpiarTemporal(ruta string) error {
	if _, err := os.Stat(ruta); err == nil {
		e := os.Remove(ruta)
		if e != nil {
			return err
		}
	}
	return nil
}
