/*
* AUTOR: Rafael Tolosana Calasanz
* ASIGNATURA: 30221 Sistemas Distribuidos del Grado en Ingeniería Informática
*			Escuela de Ingeniería y Arquitectura - Universidad de Zaragoza
* FECHA: septiembre de 2021
* FICHERO: server.go
* DESCRIPCIÓN: contiene la funcionalidad esencial para realizar los servidores
*				correspondientes a la práctica 1

* La arquitectura cliente servidor secuencial consiste en un servidor que atien-
* de peticiones de forma secuencial, de una en una, de manera que cuando llegan
* varias peticiones, atiende una de ellas (a menudo la primera en llegar) y, una vez
* terminada, atiende la siguiente. Para reducir el tiempo de espera de los clientes,
* siempre que haya recursos hardware suficientes en el servidor y siempre que la apli-
* caci ́on lo permita, sse puede utilizar la arquitectura cliente-servidor concurrente
 */
package main

import (
	"encoding/gob"
	"fmt"
	"net"
	"os"

	"p1/com"
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

func main() {

	CONN_TYPE := "tcp"
	CONN_HOST := "localhost"
	CONN_PORT := "30000"

	listener, err := net.Listen(CONN_TYPE, CONN_HOST+":"+CONN_PORT)
	checkError(err)

	for {
		conn, err := listener.Accept()
		defer conn.Close()
		checkError(err)
		encoder := gob.NewEncoder(conn)
		decoder := gob.NewDecoder(conn)
		var interval com.Request
		decoder.Decode(&interval)
		id := interval.Id
		var primes com.Reply
		primes.Id = id
		primes.Primes = FindPrimes(interval.Interval)
		fmt.Println(primes.Id)
		encoder.Encode(primes)
	}
}