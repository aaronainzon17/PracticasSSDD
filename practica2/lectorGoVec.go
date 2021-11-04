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
	"practica2/raGoVec"
	"strconv"
	"time"

	"github.com/DistributedClocks/GoVector/govec"
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

	// Initialize GoVector logger
	logger := govec.InitGoVector("Lector"+os.Args[1], "LogFile"+os.Args[1], govec.GetDefaultConfig())

	File := raGoVec.New(me, path, N, 0, logger)
	go File.GestionReqRes()

	time.Sleep(5 * time.Second) //Para dar tiempo a lanzar el resto

	for {
		File.PreProtocol()

		// SC
		File.Log.LogLocalEvent("SC", govec.GetDefaultLogOptions())
		File.Ms.Send(PidGestorFIchero, ms.Leer{Fase: "Comienzo de lectura", OpType: File.OpType, Me: me})
		rand.Seed(time.Now().UnixNano())
		a := rand.Intn(7)
		for i := 0; i < a; i++ {
			File.Ms.Send(PidGestorFIchero, ms.Escribir{Fase: "Leyendo...", OpType: File.OpType, Me: me})
		}
		File.Ms.Send(PidGestorFIchero, ms.Escribir{Fase: "Fin de lectura", OpType: File.OpType, Me: me})
		//FSC

		File.PostProtocol()
		time.Sleep(time.Duration(a) * time.Second)
	}
}
