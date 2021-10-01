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
	"math/big"
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

func IntToBytes(i int) []byte {
	return big.NewInt(int64(i)).Bytes()
}
func BytesToInt(b []byte) int {
	return int(big.NewInt(0).SetBytes(b[:len(b)]).Int64())
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

	// TO DO
	// Make a buffer to hold incoming data.
	buf := make([]byte, 8)

	n, err1 := conn.Read(buf)
	fmt.Println(n)
	checkError(err1)
	ini := BytesToInt(buf[0:])
	fin := BytesToInt(buf[4:])
	fmt.Printf("El valor de ini es %d y el de fin %d \n", ini, fin)
	interval := com.TPInterval{ini, fin}
	primes := FindPrimes(interval)

	sizeOfSolve := 4 * len(primes)
	_, _ = conn.Write(IntToBytes(sizeOfSolve))
	checkError(err1)

}
