package convertidor

import (
	"log"
	"os"
	"os/exec"
)

func CmdWrapper(xlsFile string, outCSVFile string) error {
	// LLamar al comando de rust para convertir el xls a csv
	if _, err := os.Stat(outCSVFile); os.IsExist(err) {
		log.Println("Eliminando xls..")
		os.Remove(xlsFile)
	}

	cmd := exec.Command("xls2csv", xlsFile)
	csvFile, err := os.Create(outCSVFile)
	if err != nil {
		return err
	}
	defer csvFile.Close()

	cmd.Stdout = csvFile
	err = cmd.Run()
	if err != nil {
		return err
	}
	return nil
}
