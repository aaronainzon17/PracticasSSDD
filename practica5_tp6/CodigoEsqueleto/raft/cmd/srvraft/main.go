package main

import (
	//"errors"
	//"fmt"
	//"log"
	"fmt"
	"net"
	"net/rpc"
	"os"
	"raft/internal/almacenamiento"
	"raft/internal/comun/check"
	"raft/internal/comun/rpctimeout"
	"raft/internal/raft"
	"strconv"
	"strings"
	//"time"
)

func main() {
	// obtener entero de indice de este nodo
	meIni := os.Args[1]
	me1 := strings.Split(meIni, "-")
	meIni = me1[1]
	me, err := strconv.Atoi(meIni)
	check.CheckError(err, "Main, mal numero entero de indice de nodo:")
	var nodos []rpctimeout.HostPort
	// Resto de argumento son los end points como strings
	// De todas la replicas-> pasarlos a HostPort
	for _, endPoint := range os.Args[2:] {
		nodos = append(nodos, rpctimeout.HostPort(endPoint))
	}
	chAplicar := make(chan raft.AplicaOperacion, 1000)
	//Almacen en RAM
	go gestionAlmacenamiento(chAplicar)
	// Parte Servidor
	nr := raft.NuevoNodo(nodos, me, chAplicar)
	rpc.Register(nr)

	//fmt.Println("Replica escucha en :", me, " de ", os.Args[2:])

	l, err := net.Listen("tcp", os.Args[2:][me])
	check.CheckError(err, "Main listen error:")

	rpc.Accept(l)
}

func gestionAlmacenamiento(canalAplicarOperacion chan raft.AplicaOperacion) {
	al := almacenamiento.NewAlmacen()
	for {
		data := <-canalAplicarOperacion
		fmt.Println("Me llega una op por el canal")
		if data.Operacion.Operacion == "leer" {
			al.Leer(data.Operacion.Clave)
		} else if data.Operacion.Operacion == "escribir" {
			al.Escribir(data.Operacion.Clave, data.Operacion.Valor)
		} else {
			fmt.Println("Se intenta realizar una operacion no permitida")
		}
		al.DumpData()
	}
}
