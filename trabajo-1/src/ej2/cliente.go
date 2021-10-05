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
	"bufio"
	"fmt"
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

func main() {
	endpoint := "localhost:2000"

	// TODO: crear el intervalo solicitando dos números por teclado
	var ini, fin string
	fmt.Println("Enter an integer value : ")

	//Se pide por pantalla el incio y fin del intervalo
	_, err := fmt.Scanf("%s %s", &ini, &fin)
	interval := ini + "*" + fin + "*"

	tcpAddr, err := net.ResolveTCPAddr("tcp", endpoint)
	checkError(err)

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	checkError(err)

	// la variable conn es de tipo *net.TCPconn
	fmt.Printf("Connection established between %s and localhost.\n", endpoint)

	//Se envia el intervalo
	_, err = conn.Write([]byte(interval))
	checkError(err)

	//Se recibe el tamanyo del vector de primos
	bufSizeOfSolve := make([]byte, 10)
	_, err = conn.Read(bufSizeOfSolve)
	checkError(err)
	_, err = conn.Write([]byte("ack"))
	checkError(err)
	splits := strings.Split(string(bufSizeOfSolve), "*")
	intVar, err := strconv.Atoi(splits[0])

	checkError(err)

	//Recibe el vector de numeros primos calculados
	sol := make([]byte, intVar)
	fmt.Println(len(sol))
	n, err := conn.Read(sol)
	fmt.Printf("se han leido %d bytes \n", n)

	checkError(err)

	//Se crea un fichero donde se vuelca la salida
	f, err := os.Create("./primes.txt")
	checkError(err)
	defer f.Close()
	w := bufio.NewWriter(f)
	_, err = fmt.Fprintf(w, "%v\n", string(sol))

	//Se muestra la solucion por pantalla
	fmt.Println(string(sol))
}
