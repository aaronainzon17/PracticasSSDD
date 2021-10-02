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
	"fmt"
	"math/big"
	"net"
	"os"
	"strconv"
	"strings"
)

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}

func IntToBytes(i int) []byte {
	return big.NewInt(int64(i)).Bytes()
}
func BytesToInt(b []byte) int {
	return int(big.NewInt(0).SetBytes(b[:len(b)]).Int64())
}

func main() {
	endpoint := "localhost:2000"

	// TODO: crear el intervalo solicitando dos números por teclado
	var ini, fin string
	fmt.Println("Enter an integer value : ")

	_, err := fmt.Scanf("%s %s", &ini, &fin)
	cad := ini + "*" + fin + "*"
	fmt.Println("el buffer guarda ", cad)
	tcpAddr, err := net.ResolveTCPAddr("tcp", endpoint)
	checkError(err)

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	checkError(err)

	// la variable conn es de tipo *net.TCPconn
	fmt.Printf("Connection established between %s and localhost.\n", endpoint)
	_, err = conn.Write([]byte(cad))
	checkError(err)
	bufSizeOfSolve := make([]byte, 10)
	_, err = conn.Read(bufSizeOfSolve)
	checkError(err)
	a := string(bufSizeOfSolve)
	fmt.Println(a)
	_, err = conn.Write([]byte("ack"))
	checkError(err)
	splits := strings.Split(a, "*")
	intVar, err := strconv.Atoi(splits[0])
	checkError(err)
	fmt.Println(intVar)
	sol := make([]byte, intVar)
	_, err = conn.Read(sol)
	checkError(err)
	fmt.Println(string(sol))
}
