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
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"practica3/com"
	"strings"

	"golang.org/x/crypto/ssh"
)

type Master struct {
	//mutex sync.Mutex
}

type Reply struct {
	primes []int
	err    error
}

type Params struct {
	//Op2ex     string
	Interval  com.TPInterval
	ReplyChan chan Reply
}

var requestChan = make(chan Params, 100) //canal para los trabajos

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}

func (p *Master) FindPrimes(interval com.TPInterval, primeList *[]int) error {
	fmt.Println("LLEGQA UNA PETICION A FIND PRIMES: ", interval)
	res := make(chan Reply, 1)
	requestChan <- Params{interval, res}
	fmt.Println("NUEVA PETICION REGISTRADA: ", interval)
	result := <-res

	if result.err != nil {
		fmt.Println("Tarea fallida: ", result.err)
		return result.err
	}
	fmt.Println("Tarea completada: ", interval)
	*primeList = result.primes
	return nil
}

func (P *Master) workerControl(workerIp string) {

	for {
		job := <-requestChan // Recibe un tabajo del canal
		fmt.Println(job)

		// Se establece una conexion TCP con el worker
		workerCon, err := rpc.DialHTTP("tcp", workerIp)
		checkError(err)

		// Asynchronous call
		var reply []int
		divCall := workerCon.Go("PrimesImpl.FindPrimes", job.Interval, &reply, nil)
		select {
		//Caso en el que el worker acaba correctamente
		case rep := <-divCall.Done:
			if rep.Error == nil { //Si no hay error se guarda la respuesta en el tipo Reply
				job.ReplyChan <- Reply{primes: reply, err: rep.Error}
			}
		}
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

func runCmd(cmd string, client string, s *ssh.ClientConfig) error {
	// open connection
	fmt.Println("Client: ", client)
	conn, err := ssh.Dial("tcp", client+":22", s)
	checkError(err)
	defer conn.Close()

	// open session
	session, err := conn.NewSession()
	checkError(err)
	defer session.Close()

	// run command and capture stdout/stderr
	_, err = session.CombinedOutput(cmd)
	session.Close()
	conn.Close()

	return err
}

func sshWorkerUp(worker string, hostUser string, remoteUser string) {
	pemBytes, err := ioutil.ReadFile("/home/" + hostUser + "/.ssh/id_rsa")
	checkError(err)
	signer, err := ssh.ParsePrivateKey(pemBytes)
	checkError(err)

	config := &ssh.ClientConfig{
		User: remoteUser,
		Auth: []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			// use OpenSSH's known_hosts file if you care about host validation
			return nil
		},
	}
	res1 := strings.Split(worker, ":")
	cmd := "./worker " + worker + " &"
	fmt.Println("Comando:", cmd)
	err = runCmd(cmd, res1[0], config)
	checkError(err)
}

func main() {
	if len(os.Args) < 5 {
		fmt.Fprint(os.Stderr, "Usage:go run master.go <ip:port> <path to workers ip file> <hostUser> <remoteUser>\n")
		os.Exit(1)
	}
	//Ip y pueto del worker
	ipPort := os.Args[1]
	//Se leen las ip y puerto de fichero
	workers := readFile(os.Args[2])
	hostUser := os.Args[3]
	remoteUser := os.Args[4]

	master := new(Master)

	l, err := net.Listen("tcp", ipPort)
	checkError(err)

	fmt.Println("SERVING ...")

	//rpc.Accept(l)
	rpc.Register(master)
	fmt.Println("Registro una peticion")
	rpc.HandleHTTP()
	http.Serve(l, nil)

}