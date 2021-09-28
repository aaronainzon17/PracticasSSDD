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
	"net"
	"os"
	"trabajo1/src/com"
)

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}

func main() {
	endpoint := "localhost:2000"

	// TODO: crear el intervalo solicitando dos números por teclado
	var ini, fin int
	fmt.Println("Enter an integer value : ")

	_, err := fmt.Scanf("%d %d", &ini, &fin)
	interval := com.TPInterval{ini, fin}
	req1 := com.Request{1, interval}
	fmt.Printf("El valor de ini es %d y el de fin %d \n", interval.A, interval.B)
	tcpAddr, err := net.ResolveTCPAddr("tcp", endpoint)
	checkError(err)

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	checkError(err)

	// la variable conn es de tipo *net.TCPconn
	fmt.Printf("Connection established between %s and localhost.\n", endpoint)

	buf := []byte{byte(req1.Id), byte(req1.Interval.A), byte(req1.Interval.B)}
	fmt.Println("el buffer guarda ", buf)
	_, err = conn.Write(buf)
	checkError(err)
	readingBuf := make([]byte, 1024)
	_, err1 := conn.Read(readingBuf)
	checkError(err1)
	fmt.Println(string(readingBuf))
}
