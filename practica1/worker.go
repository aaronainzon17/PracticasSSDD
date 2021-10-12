/*
* AUTOR: Angel Espinosa (775750), Aaron Ibañez (779088)
*
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

//Objeto con el intervalo de primos y la conexion
type Params struct {
	interval com.Request
	conn     net.Conn
}

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

	if len(os.Args) != 2 {
		fmt.Fprint(os.Stderr, "Usage go run worker.go ip:port\n")
		os.Exit(1)
	}

	worker_dir := os.Args[1]

	listener, err := net.Listen("tcp", worker_dir)
	checkError(err)

	var interval com.Request
	fmt.Println("Launching worker at ", worker_dir)
	for {
		conn, err := listener.Accept()
		checkError(err)

		encoder := gob.NewEncoder(conn)
		decoder := gob.NewDecoder(conn)

		err = decoder.Decode(&interval)
		checkError(err)

		fmt.Println("Se atiende peticion")
		var respuesta com.Reply
		respuesta.Primes = FindPrimes(interval.Interval)
		respuesta.Id = interval.Id
		err = encoder.Encode(respuesta)
		checkError(err)

		conn.Close()

	}
}
