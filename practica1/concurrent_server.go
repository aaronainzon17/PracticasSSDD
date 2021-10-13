/*
* AUTOR: Angel Espinosa (775750), Aaron Ibañez (779088)
*
*
 */
package main

import (
	"encoding/gob"
	"fmt"
	"net"
	"os"
	"p1/com"
	"time"
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
	CONN_TYPE = "tcp"
	CONN_PORT = "30000"
)

func handleConnexion(conn net.Conn) {

	defer conn.Close()

	var reply com.Request
	encoder := gob.NewEncoder(conn)
	decoder := gob.NewDecoder(conn)

	//Se almacena en la variable reply el objeto de tipo request
	err := decoder.Decode(&reply)
	checkError(err)
	start := time.Now()
	//Se calculan los primos del intervalo
	primes := FindPrimes(reply.Interval)
	end := time.Now()
	texec := end.Sub(start)
	//Se crea un objeto de tipo com.Reply y se envia al cliente
	solution := com.Reply{reply.Id, primes}
	fmt.Println(reply.Id)
	encoder.Encode(solution)
	fmt.Println("Tiempo de ejecucion: ", texec)

}

func main() {

	if len(os.Args) != 2 {
		fmt.Fprint(os.Stderr, "Usage:go run concurrent_server.go <ip:port> \n")
		os.Exit(1)
	}
	ipPort := os.Args[1]

	listener, err := net.Listen("tcp", ipPort)
	checkError(err)

	fmt.Println("Listening on:", ipPort)

	for {
		conn, err := listener.Accept()
		checkError(err)
		go handleConnexion(conn)
	}
}
