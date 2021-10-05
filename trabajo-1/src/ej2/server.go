/*
* AUTOR: Rafael Tolosana Calasanz
* ASIGNATURA: 30221 Sistemas Distribuidos del Grado en Ingeniería Informática
*			Escuela de Ingeniería y Arquitectura - Universidad de Zaragoza
* FECHA: septiembre de 2021
* FICHERO: server.go
* DESCRIPCIÓN: contiene la funcionalidad esencial para realizar los servidores
*				correspondientes al trabajo 1
 */
package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"trabajo1/src/ej2/com"
)

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}

// PRE: verdad
// POST: IsPrime devuelve verdad si n es primo y falso en caso contrario
func IsPrime(n int) (foundDivisor bool) {
	foundDivisor = false
	for i := 2; (i < n) && !foundDivisor; i++ {
		foundDivisor = (n%i == 0)
	}
	return !foundDivisor
}

// PRE: interval.A < interval.B
// POST: FindPrimes devuelve todos los números primos comprendidos en el
// 		intervalo [interval.A, interval.B]
func FindPrimes(interval com.TPInterval) (primes []int) {
	for i := interval.A; i <= interval.B; i++ {
		if IsPrime(i) {
			primes = append(primes, i)
		}
	}
	return primes
}

const (
	CONN_HOST = "localhost"
	CONN_PORT = "2000"
	CONN_TYPE = "tcp"
)

func main() {

	listener, err := net.Listen(CONN_TYPE, CONN_HOST+":"+CONN_PORT)
	checkError(err)

	conn, err := listener.Accept()
	defer conn.Close()
	checkError(err)

	buf := make([]byte, 20)
	_, err = conn.Read(buf)
	checkError(err)
	cad := string(buf)

	split := strings.Split(cad, "*")

	ints := make([]int, len(split))

	for i := 0; i < len(split)-1; i++ {
		ints[i], err = strconv.Atoi(split[i])
		checkError(err)
	}

	interval := com.TPInterval{ints[0], ints[1]}
	primes := FindPrimes(interval)

	sizePrimes := len(primes)
	var primes2send string
	if sizePrimes > 0 {
		primes2send = strconv.Itoa(primes[0]) + " "
		for i := 1; i < sizePrimes; i++ {
			primes2send += strconv.Itoa(primes[i]) + " "
		}
	}

	fmt.Println(primes2send)
	_, err = conn.Write([]byte((strconv.Itoa(len(primes2send))) + "*"))
	checkError(err)

	bufAck := make([]byte, 3)
	_, err = conn.Read(bufAck)
	checkError(err)

	if string(bufAck) == "ack" {
		_, err = conn.Write([]byte(primes2send))
		checkError(err)
	}
}
