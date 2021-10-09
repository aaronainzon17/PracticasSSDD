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

const (
	CONN_HOST = "localhost"
	CONN_TYPE = "tcp"
	CONN_PORT = "30000"
)

func handleConnexion(conn net.Conn) {
	// close connection on exit
	defer conn.Close()
	var reply com.Request
	encoder := gob.NewEncoder(conn)
	decoder := gob.NewDecoder(conn)
	for {
		//Se almacena en la variable reply el objeto de tipo request
		err := decoder.Decode(&reply)
		checkError(err)
		//Se calculan los primos del intervalo
		primes := FindPrimes(reply.Interval)
		//Se crea un objeto de tipo com.Reply y se envia al cliente
		solution := com.Reply{reply.Id, primes}
		encoder.Encode(solution)
	}

}

func main() {

	listener, err := net.Listen(CONN_TYPE, CONN_HOST+":"+CONN_PORT)
	checkError(err)

	for {
		conn, err := listener.Accept()
		checkError(err)
		go handleConnexion(conn)
	}

}
