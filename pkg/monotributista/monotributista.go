package monotributista

import (
	"database/sql"
	"log"
	"time"
)

// Pago de monotributista
type Pago struct {
	id int
	cuit string
	fecha time.Time
	periodo time.Time
}

// Monotributista
type Mononotributista struct {
	id int
	cuit string
	nombre string
	estado string
	categoria string
}

// Inserta un nuevo monotributista en la base de datos
func NuevoMonotributista(db *sql.DB, mono Mononotributista) {
	stmt, _ := db.Prepare("INSERT INTO monotributistas (id, cuit, nombre, categoria, estado) VALUES (?, ?, ?, ?, ?)")
	stmt.Exec(nil, mono.cuit, mono.nombre, mono.categoria, mono.estado)
	defer stmt.Close()
	log.Printf("se agrego monotributita con el cuit %s", mono.cuit)
}

// Listar monotributistas
func ListarMonotributistas(db *sql.DB) []Mononotributista {
	rows, _ := db.Query("SELECT * FROM monotributistas")
	defer rows.Close()

	err := rows.Err()
	if err != nil {
		log.Panic("Error al consultar monotributistas")
	}

	monos := make([]Mononotributista, 0)

	for rows.Next() {
		mono := Mononotributista{}
		err = rows.Scan(&mono.id, &mono.cuit, &mono.nombre, &mono.categoria, &mono.estado)
		if err != nil {
			log.Panic("Error al consultar monotributista")
		}

		monos = append(monos, mono)
	}

	err =  rows.Err()
	if err != nil {
		log.Panicln("Error al consultar monotributistas")
	}
	
	return monos
}