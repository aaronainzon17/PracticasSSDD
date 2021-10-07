/*
* AUTOR: Rafael Tolosana Calasanz
* ASIGNATURA: 30221 Sistemas Distribuidos del Grado en Ingeniería Informática
*			Escuela de Ingeniería y Arquitectura - Universidad de Zaragoza
* FECHA: septiembre de 2021
* FICHERO: client.go
* DESCRIPCIÓN: cliente completo para los cuatro escenarios de la práctica 1
 */
package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"strconv"
)

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}

func main() {
	endpoint := "localhost:2000"

	if len(os.Args) != 3 {
		fmt.Println("WRONG USAGE")
		fmt.Println("Use: cliente.go <ini interval> <fin interval>")
		os.Exit(1)
	}

	ini, err := strconv.Atoi(os.Args[1])
	checkError(err)
	fin, err := strconv.Atoi(os.Args[2])
	checkError(err)
	//Buffer para almacenar cada numero en bytes
	num := make([]byte, 4)

	//Se crea un buffer para enviar el intervalo
	var interval []byte
	binary.LittleEndian.PutUint32(num, uint32(ini))
	interval = append(interval, num...)
	binary.LittleEndian.PutUint32(num, uint32(fin))
	interval = append(interval, num...)

	tcpAddr, err := net.ResolveTCPAddr("tcp", endpoint)
	checkError(err)

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	checkError(err)

	// la variable conn es de tipo *net.TCPconn
	fmt.Printf("Connection established between %s and localhost.\n", endpoint)

	//Se envia el intervalo
	_, err = conn.Write(interval)
	checkError(err)

	//Se recibe el tamanyo del vector de primos
	bufSizeOfSolve := make([]byte, 10)
	_, err = conn.Read(bufSizeOfSolve)
	checkError(err)
	intVar := int(binary.LittleEndian.Uint32(bufSizeOfSolve[0:4]))

	//Recibe el vector de numeros primos calculados
	sol := make([]byte, intVar)
	_, err = conn.Read(sol)
	checkError(err)

	var resultado []int
	var n int
	for i := 0; i < intVar; i += 4 {
		n = int(binary.LittleEndian.Uint32(sol[i : i+4]))
		resultado = append(resultado, n)
	}
	//Se muestra la solucion por pantalla
	fmt.Println(resultado)
}
