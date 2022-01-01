package almacenamiento

import (
	"fmt"
	"raft/internal/raft"
	"sync"
)

//Base de datos clave valor en RAM
type Almacen struct {
	Mux sync.Mutex
	DB  map[string]string
}

func NewAlmacen() *Almacen {
	return &Almacen{
		DB: make(map[string]string),
	}
}

func (al *Almacen) Leer(clave string) (string, bool) {
	al.Mux.Lock()
	defer al.Mux.Unlock()
	v, found := al.DB[clave]
	return v, found
}

func (al *Almacen) Escribir(clave string, valor string) {
	al.Mux.Lock()
	defer al.Mux.Unlock()
	al.DB[clave] = valor
}

func (al *Almacen) HasData() bool {
	al.Mux.Lock()
	defer al.Mux.Unlock()
	return len(al.DB) > 0
}

func (al *Almacen) gestionAlmacen(op raft.TipoOperacion) {
	if op.Operacion == "leer" {
		al.Leer(op.Clave)
	} else if op.Operacion == "escribir" {
		al.Escribir(op.Clave, op.Valor)
	}
}

func (al *Almacen) DumpData() {
	fmt.Println("-------------- Base Datos ----------------")
	for clave, valor := range al.DB {
		fmt.Println("Clave: ", clave, " Valor: ", valor)
	}
	fmt.Println("-------------------------------------------")
}
