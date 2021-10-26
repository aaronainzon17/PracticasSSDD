/*
* AUTORES: Aaron Ibañez (779088), Angel Espinosa (775750)
* FECHA: octubre de 2021
* FICHERO: gestorfichero.go
* DESCRIPCIÓN: Proceso gestorfichero, simulando una base de datos compartida
 */
package main

import (
	"fmt"
	"os"
	"practica2/ms"
	"strconv"
)

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}

func main() {
	if len(os.Args) != 3 {
		fmt.Fprint(os.Stderr, "Usage: ./gestorfichero <pid> <pathToUsers> \n")
		os.Exit(1)
	}

	me, err := strconv.Atoi(os.Args[1])
	checkError(err)
	path := os.Args[2]

	msgs := ms.New(me, path, []ms.Message{ms.Escribir{}, ms.Leer{}})
	f, err := os.Create("sharedDatabase.txt")
	checkError(err)
	//Bucle de gestion de peticiones
	for {
		msg := msgs.Receive()
		escritor, isWr := msg.(ms.Escribir)
		lector, isRd := msg.(ms.Leer)
		fmt.Println("isWr:", isWr, ",isRd:", isRd)
		if isWr { // The process wants to write
			fmt.Println("Recibo peticion de lectura de", escritor.Me)
			_, err := f.WriteString("Operation from " + strconv.Itoa(escritor.Me) + ": " + escritor.Fase + "\n")
			checkError(err)
		} else if isRd { //The process wants to read
			fmt.Println("Recibo peticion de escritura de", lector.Me)
			_, err := f.WriteString("Operation from " + strconv.Itoa(lector.Me) + ": " + lector.Fase + "\n")
			checkError(err)
		} else {
			fmt.Println("Recibo peticion desconocida de ", escritor.Me, lector.Me)
			_, err := f.WriteString("Unknown Operation\n")
			checkError(err)
		}
	}
	msgs.Stop()
}
