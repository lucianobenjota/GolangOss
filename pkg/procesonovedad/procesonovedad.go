package procesonovedad

import (
	"encoding/csv"
	"os"
	"strings"

	"github.com/gocarina/gocsv"
)

type NovedadFTP struct {
	Codigo      string `csv:"cod_dual"`
	Descripcion string `csv:"descripcion"`
	Resumen     string `csv:"resumen"`
	Accion      string `csv:"accion"`
}

type TSVNovedad struct {
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
	Movimiento       string `csv:"movimiento"`
	CodigoErr        string `csv:"codigo_err"`
	CodigoVal        string `csv:"codigo_val"`
	CUILNuevo        string `csv:"cuil_nuevo"`
	CodigoDual       string `csv:"cod_dual"`
	Descripcion      string `csv:"descripcion"`
	Resumen          string `csv:"resumen"`
	Accion           string `csv:"accion"`
}

func leerBase() (base []*NovedadFTP, err error) {
	b, err := os.Open("./codigos.csv")
	if err != nil {
		return nil, err
	}

	defer b.Close()

	err = gocsv.UnmarshalFile(b, &base)
	if err != nil {
		return nil, err
	}

	return base, nil
}

func LeerCSVFTP(file *os.File) (out []byte, err error) {
	base, err := leerBase()
	if err != nil {
		return nil, err
	}

	csvReader := csv.NewReader(file)
	csvReader.Comma = '|'
	csvReader.FieldsPerRecord = -1

	csvLines, err := csvReader.ReadAll()
	if err != nil {
		return nil, err
	}
	data := []TSVNovedad{}

	for _, line := range csvLines {
		// Verificamos que la 1ra linea sea el RNOS
		if line[0] != "128201" {
			break
		}

		row := TSVNovedad{
			RNOS:             line[0],
			CUIT:             line[1],
			CUILTitular:      line[2],
			Parentesco:       line[3],
			CUIL:             line[4],
			TipoDNI:          line[5],
			NroDNI:           line[6],
			Nombre:           line[7],
			Sexo:             line[8],
			EstadoCivil:      line[9],
			FechaNac:         line[10],
			Nacionalidad:     line[11],
			Calle:            line[12],
			NroPuerta:        line[13],
			Piso:             line[14],
			Depto:            line[15],
			Localidad:        line[16],
			CodPostal:        line[17],
			Provincia:        line[18],
			TipoDomicilio:    line[19],
			Telefono:         line[20],
			SituacionRevista: line[21],
			Incapacidad:      line[22],
			TipoBenTitular:   line[23],
			FechaAltaOS:      line[24],
			FechaCierre:      line[25],
			Movimiento:       line[26],
			CodigoErr:        line[27],
			CodigoVal:        line[28],
			CUILNuevo:        line[29],
			CodigoDual:       line[27] + line[28],
		}

		for _, k := range base {
			if strings.TrimSpace(row.CodigoDual) == strings.TrimSpace(k.Codigo) {
				row.Accion = k.Accion
				row.Descripcion = k.Descripcion
				row.Resumen = k.Resumen
			}
		}

		data = append(data, row)
	}

	out, err = gocsv.MarshalBytes(data)
	if err != nil {
		return nil, err
	}
	return out, nil
}
