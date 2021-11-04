/*
* AUTORES: Aaron Ibañez (779088), Angel Espinosa (775750)
* FECHA: octubre de 2021
* FICHERO: lector.go
* DESCRIPCIÓN: Proceso lector para el problema de los lectores, escritores
 */
package main

import (
	"fmt"
	"math/rand"
	"os"
	"practica2/ms"
	"practica2/ra"
	"strconv"
	"time"
)

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
	PidGestorFIchero := N + 1
	path := os.Args[3]

	File := ra.New(me, path, N, 0)
	go File.GestionReqRes()

	time.Sleep(5 * time.Second) //Para dar tiempo a lanzar el resto

	for {
		File.PreProtocol()

		// SC
		fmt.Println("Leyendo...")
		File.Ms.Send(PidGestorFIchero, ms.Leer{Fase: "Comienzo de lectura", OpType: File.OpType, Me: me})
		rand.Seed(time.Now().UnixNano())
		a := rand.Intn(7)
		for i := 0; i < a; i++ {
			fmt.Println("Leyendo", a, "lineas")
			File.Ms.Send(PidGestorFIchero, ms.Escribir{Fase: "Leyendo...", OpType: File.OpType, Me: me})
		}
		File.Ms.Send(PidGestorFIchero, ms.Escribir{Fase: "Fin de lectura", OpType: File.OpType, Me: me})
		//FSC

		File.PostProtocol()
		time.Sleep(time.Duration(a) * time.Second)
	}
}
