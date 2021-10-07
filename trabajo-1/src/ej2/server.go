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
	"encoding/binary"
	"fmt"
	"net"
	"os"
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

	intervalRead := make([]byte, 8)
	_, err = conn.Read(intervalRead)
	checkError(err)

	ini := int(binary.LittleEndian.Uint32(intervalRead[0:4]))
	fin := int(binary.LittleEndian.Uint32(intervalRead[4:8]))
	fmt.Printf("El intervalo recibido es %d, %d \n", ini, fin)
	interval := com.TPInterval{ini, fin}
	primes := FindPrimes(interval)
	fmt.Println(primes)

	//Tamanyo del array de primos
	sizePrimes := len(primes)

	//Buffer para almacenar cada numero en bytes
	num := make([]byte, 4)
	binary.LittleEndian.PutUint32(num, uint32(sizePrimes*4))

	//Se manda el tamaño del vector de bytes
	_, err = conn.Write(num)
	checkError(err)

	//Se crea un buffer para enviar el intervalo
	var sol []byte
	if sizePrimes > 0 {
		for i := 0; i < sizePrimes; i++ {
			binary.LittleEndian.PutUint32(num, uint32(primes[i]))
			sol = append(sol, num...)
		}
	}
	fmt.Println(sol)
	n1, err := conn.Write(sol)
	fmt.Printf("se escriben %d bytes \n", n1)
	checkError(err)
}
