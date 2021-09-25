package compania

import (
	"encoding/csv"
	"io"
	"os"
	"strings"

	"github.com/gocarina/gocsv"
	"github.com/lucianobenjota/go-oss-bot/m/pkg/convertidor"
)

type Afiliado struct {
	Numero         string `csv:"Nro. Afiliado"`
	DNI            string `csv:"Nro. Doc."`
	Nombre         string `csv:"Afiliado"`
	Parentesco     string `csv:"Tipo Afiliado"`
	FechaNac       string `csv:"Fecha Nac."`
	FechaAlta      string `csv:"Fecha Alta"`
	Localidad      string `csv:"Localidad"`
	Provincia      string `csv:"Provincia"`
	Telefono       string `csv:"Teléfono"`
	Celular        string `csv:"Celular"`
	TipoDNI        string `csv:"Tipo Doc."`
	Edad           string `csv:"Edad"`
	EstadoAfiliado string `csv:"Estado Actual"`
	NuevoTelefono  string `csv:"-"`
}

type Salida struct {
	Numero     string `csv:"numero"`
	Nrodoc     string `csv:"nrodoc"`
	Nombre     string `csv:"nombre"`
	Parentesco string `csv:"parentesco"`
	FechaNac   string `csv:"fechanac"`
	FechaAlta  string `csv:"fecha_alta"`
	Localidad  string `csv:"localidad"`
	Provincia  string `csv:"prov_nombre"`
	Telefono   string `csv:"telefono"`
	TipoDNI    string `csv:"tipodoc"`
	Edad       string `csv:"edad"`
}

var interCSV string

func ReporteACSV(reportPath string, outCSVPath string) error {
	interCSV = "archivo.csv"
	// Crear un archivo intermedio,  podria ser reemplazado por uno temporal
	err := convertidor.CmdWrapper(reportPath, interCSV)
	csvInterFile, err := os.Open(interCSV)
	if err != nil {
		return err
	}

	defer csvInterFile.Close()
	err = ProcesarReporte(csvInterFile, outCSVPath)
	if err != nil {
		return err
	}

	err = DeleteIntermediateCSV(interCSV)
	if err != nil {
		return err
	}

	return nil
}

func ProcesarReporte(csvFile *os.File, csvOutPath string) error {
	// Procesar el reporte de compañia en formato csv

	empadronamiento := []*Afiliado{}

	if err := gocsv.UnmarshalFile(csvFile, &empadronamiento); err != nil {
		return err
	}

	salida := []*Salida{}

	for _, af := range empadronamiento {
		res := Salida{}
		af.Telefono = strings.TrimSpace(af.Telefono)
		af.Celular = strings.TrimSpace(af.Celular)
		af.NuevoTelefono = getTelefono(af.Telefono, af.Celular)
		af.EstadoAfiliado = strings.TrimSpace(af.EstadoAfiliado)
		if af.EstadoAfiliado == "Activo" {
			res.Numero = strings.TrimSpace(af.Numero)
			res.Nrodoc = strings.TrimSpace(af.DNI)
			res.Nombre = strings.TrimSpace(af.Nombre)
			res.Parentesco = strings.TrimSpace(af.Parentesco)
			res.FechaNac = strings.TrimSpace(af.FechaNac)
			res.FechaAlta = strings.TrimSpace(af.FechaAlta)
			res.Localidad = strings.TrimSpace(af.Localidad)
			res.Provincia = strings.TrimSpace(af.Provincia)
			res.Telefono = strings.TrimSpace(af.NuevoTelefono)
			res.TipoDNI = strings.TrimSpace(af.TipoDNI)
			res.Edad = strings.TrimSpace(af.Edad)
			salida = append(salida, &res)
		}
	}

	gocsv.SetCSVWriter(func(out io.Writer) *gocsv.SafeCSVWriter {
		writer := csv.NewWriter(out)
		writer.Comma = ';'
		writer.UseCRLF = true
		return gocsv.NewSafeCSVWriter(writer)
	})

	csvOutFile, err := os.OpenFile(csvOutPath, os.O_RDWR|os.O_CREATE, os.ModePerm)

	if err != nil {
		return err
	}

	err = gocsv.MarshalFile(&salida, csvOutFile)

	if err != nil {
		return err
	}

	return nil
}

func getTelefono(telefono string, celular string) string {
	// Si existe el celular devuelve el celular si no devuelvel el telefono
	if len(celular) > 0 {
		return celular
	}
	return telefono
}

func DeleteIntermediateCSV(csvPath string) error {
	err := os.Remove(csvPath)
	if err != nil {
		return err
	}
	return nil
}
