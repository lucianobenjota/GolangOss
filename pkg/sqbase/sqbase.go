package sqbase

import (
	"database/sql"
	"log"
)

// Para loguear error directamente
func checkErr (err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func InitDb () (db *sql.DB) {
	db, err := sql.Open("sqlite3", "./monotributos.db")
	checkErr(err)
	defer db.Close()
	return db
}
