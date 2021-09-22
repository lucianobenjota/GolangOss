package procesonovedad

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

type NovedadesFTP struct {
	NovedadesFTP []NovedadFTP `json:"codigos"`
}

type NovedadFTP struct {
	Codigo      string `json:"cod_dual"`
	Descripcion string `json:"descripcion"`
	Resumen     string `json:"resumen"`
	Accion      string `json:"accion"`
}

func LeerCSVFTP(archivo *os.File) {
	base, err := os.Open("codigos.json")
	if err != nil {
		log.Panic(err)
	}

	log.Println("JSON leido correctamente")

	defer base.Close()
	byteValue, _ := ioutil.ReadAll(base)

	var Base NovedadesFTP

	json.Unmarshal(byteValue, &Base)

	for i := 0; i < len(Base.NovedadesFTP); i++ {
		fmt.Println(Base.NovedadesFTP[i].Codigo)
		fmt.Println(Base.NovedadesFTP[i].Accion)
		fmt.Println(Base.NovedadesFTP[i].Resumen)
		fmt.Println(Base.NovedadesFTP[i].Descripcion)
	}
}
