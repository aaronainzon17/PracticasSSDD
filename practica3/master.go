/*
*
* AUTOR: Angel Espinosa (775750), Aaron Ibañez (779088)
* FICHERO: master.go
* RAMA: DEVELOPMENTS
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
	"time"

	"golang.org/x/crypto/ssh"
)

type Reply struct {
	primes []int
	err    error
	worker string
}

type PrimesImpl struct {
	//Op2ex     string
	Interval  com.TPInterval
	ReplyChan chan Reply
}

var requestChan = make(chan PrimesImpl, 100) //canal para los trabajos

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}

func (p *PrimesImpl) FindPrimes(interval com.TPInterval, primeList *[]int) error {
	res := make(chan Reply, 1)
	requestChan <- PrimesImpl{interval, res}
	result := <-res
	if result.err != nil {
		fmt.Println("Tarea fallida: ", result.err, " por ", result.worker)
		//Se envia que la tarea ha fallado y se vuelve a encolar
		requestChan <- PrimesImpl{interval, res}
		return result.err
	}
	fmt.Println("Tarea completada: ", interval, " por ", result.worker)
	*primeList = result.primes
	return nil
}

func workerControl(workerIp string) {
	var reply []int
	fin := false
	w := strings.Split(workerIp, ".")
	for !fin {
		select {
		// Recibe un tabajo del canal
		case job := <-requestChan:
			// Se establece una conexion TCP con el worker
			workerCon, err := rpc.DialHTTP("tcp", workerIp)
			if err == nil { // Si no hay error
				// Asynchronous call
				divCall := workerCon.Go("PrimesImpl.FindPrimes", job.Interval, &reply, nil)
				select {
				//Caso en el que el worker acaba correctamente
				case rep := <-divCall.Done:
					//Si no hay error se guarda la respuesta en el tipo Reply
					if rep.Error == nil {
						job.ReplyChan <- Reply{reply, nil, w[len(w)-1]}
					} else {
						//Se guarda fallo
						job.ReplyChan <- Reply{reply, fmt.Errorf("Crash"), w[len(w)-1]}
						fmt.Println("Goroutine exiting on worker ")
						fin = true
					}
				case <-time.After(3 * time.Second):
					//Caso en el que salta la alarma programada por el time.After
					job.ReplyChan <- Reply{reply, fmt.Errorf("Worker fail: delay/omision"), w[len(w)-1]}
				}
			} else {
				fmt.Errorf("No se ha podido establecer conexion con: ", workerIp)
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
	cmd := "cd /home/a779088/cuarto/PracticasSSDD/practica3 && /usr/local/go/bin/go run worker.go " + worker + " &"
	//fmt.Println("Comando:", cmd)
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

	for i := range workers {
		go sshWorkerUp(workers[i], hostUser, remoteUser)
		time.Sleep(5000 * time.Millisecond)
		go workerControl(workers[i])
		fmt.Println("connecting to", workers[i])
	}
	fmt.Println("SERVING ...")

	primesImpl := new(PrimesImpl)
	rpc.Register(primesImpl)
	rpc.HandleHTTP()
	l, err := net.Listen("tcp", ipPort)
	if err != nil {
		fmt.Println("Listen error:", err)
	}
	http.Serve(l, nil)

}
