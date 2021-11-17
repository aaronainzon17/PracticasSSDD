/*
*
* AUTOR: Angel Espinosa (775750), Aaron Ibañez (779088)
* FICHERO: master.go
*
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
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

/*type Master struct {
	//mutex sync.Mutex
}*/

type Reply struct {
	primes []int
	err    error
}

type PrimesImpl struct {
	//Op2ex     string
	Interval  com.TPInterval
	ReplyChan chan Reply
}

var requestChan = make(chan PrimesImpl, 100) //canal para los trabajos
var IPWORKERS = make(chan string)            //canal para que workermanager envíe ips de workers
var MAXWORKERS = 0                           // Numero maximo de workers del sistema
var WORKERS []string                         //Ips de los workers

var NWORKERSUP = 0 // Numero de workers activos
var IPWORKERSUP []string

var DELAYED = 0
var CRASHED = 0
var REQUESTS = 0

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}

func isUp(ipW string) {

}
func getAvailableDirs() []string {
	mb := make(map[string]struct{}, len(WORKERS))
	for _, x := range WORKERS {
		mb[x] = struct{}{}
	}
	var diff []string
	for _, x := range IPWORKERSUP {
		if _, found := mb[x]; !found {
			diff = append(diff, x)
		}
	}
	return diff
}

func (p *PrimesImpl) FindPrimes(interval com.TPInterval, primeList *[]int) error {
	res := make(chan Reply, 1)
	requestChan <- PrimesImpl{interval, res}
	result := <-res
	if result.err != nil {
		fmt.Println("Tarea fallida: ", result.err)
		return result.err
	}
	fmt.Println("Tarea completada: ", interval)
	*primeList = result.primes
	return nil
}
func resourceManager(hostUser string, remoteUser string) {
	start := time.Now()
	for {
		time.Sleep(10 * time.Second)
		fmt.Println("Número de peticiones: " + strconv.Itoa(REQUESTS))
		fmt.Println("Número de workers: " + strconv.Itoa(NWORKERSUP))
		fmt.Println("Numero de delay/omission: " + strconv.Itoa(DELAYED))
		fmt.Println("Numero de crash: " + strconv.Itoa(CRASHED))
		fmt.Println(time.Since(start))
		REQUESTS = 0
		nEnqueuedReq := len(requestChan)
		// Si el numero de peticiones en el canal es 0 se elimina un worker, ya que se considera
		// que el numero de peticiones restantes por atender se pueden atender garantizando el QoS
		// con un worker menos
		if nEnqueuedReq == 0 {
			if NWORKERSUP > 1 {
				workerDir := IPWORKERSUP[len(IPWORKERSUP)-1]   //Lee el ultimo elemento del slice
				IPWORKERSUP = IPWORKERSUP[:len(IPWORKERSUP)-1] //Elimina el ultimo elemento del slice

				workerConn, err := rpc.DialHTTP("tcp", workerDir)
				if err == nil {
					var res int
					workerConn.Go("PrimesImpl.Stop", 1, &res, nil)
				}
				NWORKERSUP--
			}
			// Si hay peticiones en el canal se entiende que el numero de workers activos no son suficientes
			// para sartisfacer la demanda por lo que se levanta un worker mas
		} else {
			if NWORKERSUP != MAXWORKERS {

				availableDirs := getAvailableDirs()
				dir := availableDirs[len(availableDirs)-1]

				go sshWorkerUp(dir, hostUser, remoteUser)
				time.Sleep(5000 * time.Millisecond)
				go workerControl(dir)
				NWORKERSUP++                           // Se anyade un worker mas al contador
				IPWORKERSUP = append(IPWORKERSUP, dir) //Se anyade la ip del worker a la lista de workers up
			}
		}
	}
}

//Si te caes te levantas
func workerManager(hostUser string, remoteUser string) {
	for {
		workerIp := <-IPWORKERS
		go sshWorkerUp(workerIp, hostUser, remoteUser)
		time.Sleep(5000 * time.Millisecond)
		go workerControl(workerIp)
		fmt.Println("Reconnecting to", workerIp)
	}
}

func workerControl(workerIp string) {
	var reply []int
	fin := false
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
						job.ReplyChan <- Reply{primes: reply, err: rep.Error}
					} else {
						//Se guarda fallo
						job.ReplyChan <- Reply{reply, fmt.Errorf("Worker crashed")}
						CRASHED++
						//fmt.Println("Goroutine exiting on worker ")
						//Se envia la direccion del worker caido por el canal para intentar levantarlo
						IPWORKERS <- workerIp
						fin = true
					}
				case <-time.After(3 * time.Second):
					job.ReplyChan <- Reply{reply, fmt.Errorf("Worker fail: delay/omision")}
					DELAYED++
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

	var ips []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		ips = append(ips, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	f.Close()
	return ips
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
	WORKERS = readFile(os.Args[2])
	hostUser := os.Args[3]
	remoteUser := os.Args[4]

	for i := range WORKERS {
		go sshWorkerUp(WORKERS[i], hostUser, remoteUser)
		time.Sleep(5000 * time.Millisecond)
		go workerControl(WORKERS[i])
		MAXWORKERS++
		NWORKERSUP++
		fmt.Println("connecting to", WORKERS[i])
	}

	go workerManager(hostUser, remoteUser)
	go resourceManager(hostUser, remoteUser)

	fmt.Println("SERVING ...")

	primesImpl := new(PrimesImpl)
	rpc.Register(primesImpl)
	rpc.HandleHTTP()
	l, err := net.Listen("tcp", ipPort)
	checkError(err)
	http.Serve(l, nil)

}
