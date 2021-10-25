/*
* AUTORES: Aaron Ibañez (779088), Angel Espinosa (775750)
* FECHA: octubre de 2021
* FICHERO: lector.go
* DESCRIPCIÓN: Proceso lector para el problema de los lectores, escritores
 */
package main

import (
	"fmt"
	"os"
	"practica2/ms"
	"practica2/ra"
	"strconv"
	"time"
)

var File *ra.RASharedDB

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}

func main() {
	if len(os.Args) != 4 {
		fmt.Fprint(os.Stderr, "Usage: ./escritor <me> <N> <pathToUsers> \n")
		os.Exit(1)
	}
	me, err := strconv.Atoi(os.Args[1])
	checkError(err)
	N, err := strconv.Atoi(os.Args[2])
	checkError(err)
	path := os.Args[3]

	File = ra.New(me, path, N, 1)
	go File.RecieveReqRes()

	time.Sleep(1 * time.Second) //Para dar tiempo a lanzar el resto

	for {
		File.PreProtocol()
		// SC
		File.Ms.Send(1, ms.Leer{Fase: "Leyendo", OpType: 0, Me: me})
		//FSC
		File.PostProtocol()
		time.Sleep(2000 * time.Millisecond)
	}
}
