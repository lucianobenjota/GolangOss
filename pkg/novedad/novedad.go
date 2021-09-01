package novedad

import (
	"strings"
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

// Crear una nueva novedad desde el afiliado
func (n Novedad) NuevaNovedad(af AfReporteMICAM) {
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
}

// Obtiene el cuil del titular desde el reporte de micam
func (n Novedad) ObtenerCUILTitular(emp []AfReporteMICAM) {
	nroTit := n.Grupo + "00"
	var res string
	for _, v := range emp {
		if v.NroAf == nroTit {
			res = strings.ReplaceAll(v.CUIL, "-", "")
		}
	}
	n.CUILTitular = res
}

// Obtiene el telefono desde el reporte de MICAM
func obtenerTelefono(a AfReporteMICAM) string {
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
