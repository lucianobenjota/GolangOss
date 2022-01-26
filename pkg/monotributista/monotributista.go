package monotributista

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Pago de monotributista
type Pago struct {
	Id int
	Cuit string
	Fecha time.Time
	Periodo time.Time
	Concepto string
	Nro_secuencia string
	Credito string
	Debito string 
	Rnos string
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
func ListarMonotributistas(db *sql.DB, ) []Mononotributista {
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

// Registramos el pago del monotributista en la base datos
func RegistrarPago(db *sql.DB, pago Pago) (err error) {
	q := `INSERT INTO 
					pagos(cuit, fecha, periodo, concepto, nro_secuencia, credito, debito, rnos)
				VALUES(?, ?, ?, ?, ?, ?, ?, ?);`

	stmt, _ := db.Prepare(q)
	
	log.Println(
		pago.Cuit, 
		pago.Fecha, 
		pago.Periodo, 
		pago.Concepto, 
		pago.Nro_secuencia,
		pago.Credito,
		pago.Nro_secuencia,
		pago.Credito,
		pago.Debito,
		pago.Rnos)

	_, err = stmt.Exec(
		pago.Cuit,
		pago.Fecha, 
		pago.Periodo, 
		pago.Concepto, 
		pago.Nro_secuencia, 
		pago.Credito, 
		pago.Debito, 
		pago.Rnos)
	
	if err != nil {
		return err
	}

	defer stmt.Close()

	log.Printf("Se registro correctamente el pago del cuit %s", pago.Cuit)
	return
}