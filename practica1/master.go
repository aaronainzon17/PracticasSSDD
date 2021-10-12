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
	"bufio"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"p1/com"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
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
		//defer workerConn.Close()
		checkError(err)

		workerEnc := gob.NewEncoder(workerConn)
		workerDec := gob.NewDecoder(workerConn)

		err = workerEnc.Encode(job.interval)
		checkError(err)

		err = workerDec.Decode(&reply)
		checkError(err)

		workerConn.Close()

		end := time.Now()
		texec := end.Sub(start)
		clientEnc := gob.NewEncoder(clientConn)

		err = clientEnc.Encode(reply)
		checkError(err)

		fmt.Println("Tiempo de ejecucion: \n", texec)

		// close connection on exit
		clientConn.Close()
	}
}

func readFile(path string) []string {
	fmt.Println("entra a leer el fichero ", path)

	f, err := os.Open(path)
	checkError(err)

	var workers []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		workers = append(workers, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	f.Close()
	return workers
}

func runCmd(cmd string, client string, s *ssh.ClientConfig) (string, error) {
	// open connection
	conn, err := ssh.Dial("tcp", client+":22", s)
	checkError(err)
	defer conn.Close()

	// open session
	session, err := conn.NewSession()
	checkError(err)
	defer session.Close()

	// run command and capture stdout/stderr
	output, err := session.CombinedOutput(cmd)

	return fmt.Sprintf("%s", output), err
}

func sshWorkerUp(worker string) (string, error) {
	pemBytes, err := ioutil.ReadFile("/home/" + "aaron" + "/.ssh/id_rsa")
	checkError(err)
	signer, err := ssh.ParsePrivateKey(pemBytes)
	checkError(err)

	config := &ssh.ClientConfig{
		User: "a779088",
		Auth: []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			// use OpenSSH's known_hosts file if you care about host validation
			return nil
		},
	}
	res1 := strings.Split(worker, ":")
	res, err := runCmd("cd /home/a779088/cuarto/PracticasSSDD/practica1/ && /usr/local/go/bin/go run worker.go "+worker, res1[0], config)
	return res, err
}

func main() {
	if len(os.Args) < 3 {
		fmt.Fprint(os.Stderr, "Usage:go run master.go <ip:port> <path to workers ip file>\n")
		os.Exit(1)
	}
	//Ip y pueto del worker
	ip := os.Args[1]
	//Se leen las ip y puerto de fichero
	workers := readFile(os.Args[2])

	listener, err := net.Listen("tcp", ip)
	checkError(err)

	//Se crea un canal y se lanzan las gorutines
	ch := make(chan Params)
	for i := range workers {
		res, err := sshWorkerUp(workers[i])
		checkError(err)
		fmt.Println(res)
		go workerControl(ch, workers[i])
		//fmt.Println("connecting to", workers[i])
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
