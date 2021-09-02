package novedad

import (
	"os"
	"strconv"
	"strings"

	"github.com/gocarina/gocsv"
)

// Estructura de tags de novedad a exportar como CSV
type Novedad struct {
	RNOS             string `csv:"rnos"`
	CUIT             string `csv:"cuit"`
	CUILTitular      string `csv:"cuil_titular"`
	Parentesco       string `csv:"parentesco"`
	CUIL             string `csv:"cuil"`
	TipoDNI          string `csv:"tipo_dni"`
	NroDNI           string `csv:"nro_dni"`
	Nombre           string `csv:"nombre"`
	Sexo             string `csv:"sexo"`
	EstadoCivil      string `csv:"estado_civil"`
	FechaNac         string `csv:"fecha_nac"`
	Nacionalidad     string `csv:"nacionalidad"`
	Calle            string `csv:"calle"`
	NroPuerta        string `csv:"nro_puerta"`
	Piso             string `csv:"piso"`
	Depto            string `csv:"depto"`
	Localidad        string `csv:"localidad"`
	CodPostal        string `csv:"cod_postal"`
	Provincia        string `csv:"provincia"`
	TipoDomicilio    string `csv:"tipo_domicilio"`
	Telefono         string `csv:"telefono"`
	SituacionRevista string `csv:"situacion_revista"`
	Incapacidad      string `csv:"incapacidad"`
	TipoBenTitular   string `csv:"tipo_ben_titular"`
	FechaAltaOS      string `csv:"fecha_alta_os"`
	FechaCierre      string `csv:"fecha_cierre"`
	Grupo            string `csv:"-"`
}

// Estructura del afiliado del reporte de MICAM
type AfReporteMICAM struct {
	Grupo         string `csv:"Nro. Grupo"`
	NroAf         string `csv:"Nro. Afiliado"`
	CUIL          string `csv:"CUIL"`
	DNI           string `csv:"Nro. Doc."`
	Nombre        string `csv:"Afiliado"`
	Categoria     string `csv:"Categoría"`
	Cobertura     string `csv:"Cobertura"`
	Plan          string `csv:"Plan Cob."`
	EstadoActual  string `csv:"Estado Actual"`
	Observaciones string `csv:"Observaciones"`
	Barrio        string `csv:"Barrio"`
	Celular       string `csv:"Celular"`
	CodPostal     string `csv:"Código Postal"`
	Condicion     string `csv:"Condición"`
	Convenio      string `csv:"Convenio"`
	Documento     string `csv:"Documento"`
	Domicilio     string `csv:"Domicilio"`
	Edad          string `csv:"Edad"`
	Empresa       string `csv:"Empresa"`
	EstadoCivil   string `csv:"Estado Civil"`
	FechaAlta     string `csv:"Fecha Alta"`
	FechaEgreso   string `csv:"Fecha Egreso"`
	FechaNac      string `csv:"Fecha Nac."`
	Ingreso       string `csv:"Ingreso"`
	Legajo        string `csv:"Legajo"`
	Localidad     string `csv:"Localidad"`
	Mail          string `csv:"Mail"`
	Nacionalidad  string `csv:"Nacionalidad"`
	PlanPropio    string `csv:"Plan Propio"`
	Provincia     string `csv:"Provincia"`
	Sexo          string `csv:"Sexo"`
	TelefonoFijo  string `csv:"Teléfono"`
	TipoAf        string `csv:"Tipo Afiliado"`
	TipoDNI       string `csv:"Tipo Doc."`
	Zona          string `csv:"Zona"`
}

// Obtener las novedades desde el archivo csv de reporte
func CSVANovedad(archivoCSV *os.File, rutaSalida string) error {
	reporte := []*AfReporteMICAM{}

	salida := []*Novedad{}

	if err := gocsv.UnmarshalFile(archivoCSV, &reporte); err != nil {
		return err
	}

	for _, v := range reporte {
		nv := Novedad{}
		nv = nv.NuevaNovedad(v)
		salida = append(salida, &nv)
	}

	archivoSalida, err := os.OpenFile(rutaSalida, os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		return err
	}
	err = gocsv.MarshalFile(&salida, archivoSalida)
	if err != nil {
		return err
	}
	return nil
}

// Crear una nueva novedad desde el afiliado
func (n Novedad) NuevaNovedad(af *AfReporteMICAM) Novedad {
	n.RNOS = "128201"
	n.Grupo = strings.TrimSpace(af.Grupo)
	n.CUIL = strings.ReplaceAll(af.CUIL, "-", "")
	n.TipoDNI = "DU"
	n.NroDNI = af.DNI
	n.Nombre = formatNombre(af.Nombre)
	n.Sexo = string(af.Sexo[0])
	n.FechaNac = strings.ReplaceAll(af.FechaNac, "/", "")
	n.Localidad = formatLocalidad(af.Localidad)
	n.CodPostal = formatCPA(af.CodPostal)
	n.Telefono = obtenerTelefono(af)
	n.EstadoCivil = obtenerEstadoCivil(af.EstadoCivil)
	n.Parentesco = obtenerParentesco(af.TipoAf, af.Edad)
	n.Calle, n.NroPuerta = obtenerDomicilio(af.Domicilio)
	n.Piso = "    "
	n.Depto = "    "
	n.Provincia = "04"
	return n
}

// Obtiene el telefono desde el reporte de MICAM
func obtenerTelefono(a *AfReporteMICAM) string {
	res := a.TelefonoFijo
	if len(a.Celular) > 0 {
		res = a.Celular
	}
	res = strings.ReplaceAll(res, "-", "")

	if len(res) > 20 {
		res = string(res[0:20])
	}
	res = PadRight(res, " ", 20)
	return res
}

// Formatea el cod postal a 8 digitos con espacios
func formatCPA(cpa string) string {
	res := cpa
	if len(res) == 0 {
		res = "5000"
	}
	res = PadRight(res, " ", 8)
	return res
}

// Formatea la localidad al estilo SSSalud
func formatLocalidad(localidad string) string {
	res := localidad
	if len(localidad) > 20 {
		res = res[0:20]
	}
	res = PadRight(res, " ", 20)
	return res
}

// Agrega caracteres a la derecha
func PadRight(str, pad string, lenght int) string {
	for {
		str += pad
		if len(str) > lenght {
			return str[0:lenght]
		}
	}
}

// Agrega caracteres a la izquierda
func PadLeft(str, pad string, lenght int) string {
	for {
		str = pad + str
		if len(str) > lenght {
			return str[0:lenght]
		}
	}
}

// Formatea el nombre del afiliado del reporte de MICAM al
// requerido por SSS
func formatNombre(nombre string) string {
	res := strings.ReplaceAll(nombre, ",", "")
	if len(res) >= 20 {
		res = res[0:20]
	}
	res = PadRight(res, " ", 20)
	return res
}

// Obtiene el codigo de estado civil desde el reporte
func obtenerEstadoCivil(rcivil string) string {
	in := strings.TrimSpace(rcivil)
	var res string
	switch in {
	case "Soltero":
		res = "01"
	case "Casado":
		res = "02"
	case "Viudo":
		res = "03"
	case "Divorciado":
		res = "06"
	case "Legal":
		res = "04"
	case "De hecho":
		res = "05"
	case "Convivencia":
		res = "07"
	case "No definido":
		res = "01"
	default:
		res = "01"
	}
	return res
}

// Obtener el codigo de parentesco desde el reporte
func obtenerParentesco(parentesco string, edad string) string {
	in := strings.TrimSpace(parentesco)
	var res string
	switch in {
	case "Adherente":
		res = "08"
	case "Concubino/a":
		res = "02"
	case "Conyuge":
		res = "01"
	case "Familiar a Cargo":
		res = "08"
	case "Hijastro/a a cargo menor de 21 años":
		res = "05"
	case "Hijastro/a edad 21 a 25 años que estudien":
		res = "06"
	case "Hijo/a Conyugue":
		res = "05"
	case "Hijo/a edad 21 a 25 que estudien":
		res = "04"
	case "Hijo/a incapacitados":
		e, err := strconv.Atoi(edad)
		if err != nil {
			res = "ER"
		}
		if e >= 21 {
			res = "04"
		} else {
			res = "03"
		}
	case "Hijo/a Menor 21":
		res = "03"
	case "Hijos/as":
		res = "03"
	case "Mayor de 25 años Discapacitado":
		res = "09"
	case "Menor en guarda hasta 21 años":
		res = "07"
	case "Sin Dato":
		res = "SD"
	case "Sin definir":
		res = "SD"
	case "Titular":
		res = "00"
	default:
		res = "ER"
	}
	return res
}

// Obtener la direccion y el numero de puerta desde el campo de reporte
func obtenerDomicilio(direccion string) (calle string, numero string) {
	res := strings.TrimSpace(direccion)
	palabras := strings.Fields(res)
	if res == "0" {
		return "S/D", "0"
	}
	i_nro := len(palabras) - 1
	numero = palabras[i_nro]
	calle = strings.Join(palabras[:i_nro], " ")

	if len(calle) > 20 {
		calle = string(calle[0:20])
	}

	if len(numero) > 5 {
		numero = string(numero[0:5])
	}

	calle = PadRight(calle, " ", 20)
	numero = PadRight(numero, " ", 5)
	return calle, numero
}
