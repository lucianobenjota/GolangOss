package compania

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-gota/gota/dataframe"
	"github.com/go-gota/gota/series"
)



func ArchivoACsv(rutaXLS string) (rutaCSV string, err error){
  nombreXLS := filepath.Base(rutaXLS)
  ext := filepath.Ext(rutaXLS)
  nombreCSV := strings.TrimSuffix(nombreXLS, ext) + ".csv"

  directorio := filepath.Dir(rutaXLS)
  rutaCSV = filepath.Join(directorio, nombreCSV)
  log.Printf("ruta del archivo a convertir: %v", rutaXLS)
  log.Printf("ruta de salida: %v", rutaCSV)
  cmd := exec.Command("ssconvert", rutaXLS, rutaCSV)
  log.Printf("Convirtiendo archivo a csv..")
  err = cmd.Run()
  if err != nil {
    return "", err
  }
  return rutaCSV, nil
}

func LeerCSV(rutaCSV string) (df dataframe.DataFrame, err error){
  csvFile, err := os.Open(rutaCSV)
  if err != nil {
    return dataframe.DataFrame{}, err
  }

  df = dataframe.ReadCSV(csvFile, dataframe.DefaultType(series.String))
  log.Println("df: ", df)
  return df, nil
}

func ProcesarXLS(rutaXLS string) (df dataframe.DataFrame, err error){
  var rutaCSV string
  rutaCSV, err = ArchivoACsv(rutaXLS )
  if err != nil {
    return dataframe.DataFrame{}, err
  }
  df, err = LeerCSV(rutaCSV)

  trim := func(s series.Series) series.Series {
    strs := s.String()
    var str string
    for _, f := range strs{
      str = strings.Trim(strconv.QuoteRune(f), " ") + " caca"
    }
    return series.Strings(str)
  }

  df = df.Capply(trim)
  df = df.Rapply(trim)

  return df, nil
}

func GrabarCSV(rutaCSV string, df dataframe.DataFrame) (err error){
  f, err := os.Open(rutaCSV)
  if err != nil {
    log.Fatal(err)
  }
  df.WriteCSV(f)
  return nil
}

func XLSaCSV(rutaXLS string, rutaCSV string) (err error){
  var df dataframe.DataFrame
  df, err = ProcesarXLS(rutaXLS)
  if err != nil {
    log.Fatal(err)
    return err
  }
  err = GrabarCSV(rutaCSV, df)
  if err != nil {
    log.Fatal(err)
    return err
  }
  return nil
}
