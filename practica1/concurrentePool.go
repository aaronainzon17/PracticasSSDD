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

const (
	CONN_HOST = "localhost"
	CONN_TYPE = "tcp"
	CONN_PORT = "30000"
)

func handleConnexion(ch chan Params) {

	for {
		p := <-ch // Recube un tabajo del canal
		conn := p.conn
		encoder := gob.NewEncoder(conn)
		start := time.Now()
		//Se calculan los primos del intervalo
		primes := FindPrimes(p.interval.Interval)
		end := time.Now()
		texec := end.Sub(start)
		//Se crea un objeto de tipo com.Reply y se envia al cliente
		solution := com.Reply{p.interval.Id, primes}
		encoder.Encode(solution)
		fmt.Println("Tiempo de ejecucion: ", texec)
		// close connection on exit
		conn.Close()
	}
}

func main() {

	listener, err := net.Listen(CONN_TYPE, CONN_HOST+":"+CONN_PORT)
	checkError(err)
	//Se crea un canal y se lanzan las gorutines
	ch := make(chan Params)
	for i := 0; i < 4; i++ {
		go handleConnexion(ch)
	}

	var interval com.Request
	var hcArgs Params

	for {
		conn, err := listener.Accept()
		checkError(err)

		decoder := gob.NewDecoder(conn)

		err = decoder.Decode(&interval)
		checkError(err)

		hcArgs.conn = conn
		hcArgs.interval = interval

		ch <- hcArgs //anyade el trabajo al canal (pool de Gorutines)
	}

}
