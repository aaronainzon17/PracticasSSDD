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
	"unsafe"
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
	var ini, fin int
	fmt.Println("Enter an integer value : ")

	_, err := fmt.Scanf("%d %d", &ini, &fin)
	fmt.Println(unsafe.Sizeof(int(0)))
	//interval := com.TPInterval{ini, fin}
	//req1 := com.Request{1, interval}
	tcpAddr, err := net.ResolveTCPAddr("tcp", endpoint)
	checkError(err)

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	checkError(err)

	// la variable conn es de tipo *net.TCPconn
	fmt.Printf("Connection established between %s and localhost.\n", endpoint)
	buf := make([]byte, 8)
	buf = IntToBytes(ini)
	buf = append(buf, IntToBytes(fin)...)
	fmt.Println("el buffer guarda ", buf)
	written, err := conn.Write(buf)
	fmt.Printf("El numero de bytes escrito es %d \n", written)
	checkError(err)
	bufSizeOfSolve := make([]byte, 4)
	_, err1 := conn.Read(bufSizeOfSolve)
	checkError(err1)
	size := BytesToInt(bufSizeOfSolve)
	sol := make([]byte, size)
	_, err2 := conn.Read(sol)
	checkError(err2)
	for i := 4; i < 10; i += 4 {
		fmt.Println(BytesToInt((sol[i-4 : i])))
	}

}
