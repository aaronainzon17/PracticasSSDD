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
	"strconv"
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

func workerControl(ch chan Params, workerIp string) {
	for {
		var reply com.Reply
		job := <-ch            // Recibe un tabajo del canal
		clientConn := job.conn // Conexion al cliente

		start := time.Now() // Se inicia el contador de tiempo

		// Se establece una conexion TCP con el worker
		tcpAddr, err := net.ResolveTCPAddr("tcp", workerIp)
		checkError(err)
		workerConn, err := net.DialTCP("tcp", nil, tcpAddr)
		defer workerConn.Close()
		checkError(err)

		workerEnc := gob.NewEncoder(workerConn)
		workerDec := gob.NewDecoder(workerConn)

		err = workerEnc.Encode(job.interval)
		checkError(err)

		err = workerDec.Decode(&reply)
		checkError(err)

		end := time.Now()
		texec := end.Sub(start)
		clientEnc := gob.NewEncoder(clientConn)

		err = clientEnc.Encode(reply)
		checkError(err)

		fmt.Println("Tiempo de ejecucion: ", texec)

		// close connection on exit
		clientConn.Close()
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, "Usage:go run master.go <ip:port> <nWorkers> [workeriIP:porti]\n")
		os.Exit(1)
	}
	ip := os.Args[1]
	nWorkers, err := strconv.Atoi(os.Args[2])
	checkError(err)
	listener, err := net.Listen("tcp", ip)
	checkError(err)

	//Se crea un canal y se lanzan las gorutines
	ch := make(chan Params)
	for i := 0; i < nWorkers; i++ {
		a := "localhost:" + strconv.Itoa((30000 + i))
		go workerControl(ch, a)
		fmt.Println("connecting to", a)
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
