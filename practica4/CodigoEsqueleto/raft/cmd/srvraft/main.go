package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"raft/internal/raft"
	"strconv"
	"time"
)

type NrArgs struct {
	Operacion interface{}
}

type NrReply struct {
	Indice  int
	Mandato int
	EsLider bool
}

type OpsServer int

var nr *raft.NodoRaft

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}

func readFile(path string) ([]string, int) {
	fmt.Println("entra a leer el fichero ", path)

	f, err := os.Open(path)
	checkError(err)
	nWorkers := 0
	var workers []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		workers = append(workers, scanner.Text())
		nWorkers++
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	f.Close()
	return workers, nWorkers
}

//func (os *OpsServer) StartNode() {}
func parar() {
	time.Sleep(100 * time.Millisecond)
	nr.Para()
}

func (os *OpsServer) StopNode(args NrArgs, reply *NrReply) error {
	fmt.Println("Stopping node")
	time.Sleep(2 * time.Second)
	go parar()
	return nil
}

func (os *OpsServer) Submit(args NrArgs, reply *NrReply) error {
	reply.Indice, reply.Mandato, reply.EsLider = nr.SometerOperacion(args)
	return nil
}

func main() {
	//Se lee el valor por parametro
	rpcDir, _ := strconv.Atoi(os.Args[1])

	//Parte del servidor
	os := new(OpsServer)
	rpc.Register(os)
	nodos, _ := readFile("nodos.txt")
	nr = raft.NuevoNodo(nodos, rpcDir, nil)
	time.Sleep(3 * time.Second)
	go nr.ConnectNodes(nodos)
	rpc.Register(nr)
	rpc.HandleHTTP()
	l, err := net.Listen("tcp", nodos[rpcDir])
	if err != nil {
		fmt.Println("Listen error:", err)
	}
	http.Serve(l, nil)

}

/*func main() {
	rpcDir := os.Args[1]
	fmt.Println("LA DIR DEL MAIN:", rpcDir)
	raft := new(NodoRaft)

	// Parte Servidor
	rpc.Register(raft)
	rpc.HandleHTTP()
	go func() {
		l, err := net.Listen("tcp", rpcDir)
		if err != nil {
			fmt.Println("Listen error:", err)
		}
		http.Serve(l, nil)
	}()

	go CallMethod(rpcDir)
	// Quitar el lanzamiento de la gorutina, pero no el código interno.
	// Solo se necesita para esta prueba dado que cliente y servidor estan,
	// aqui, juntos
	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				continue
			}

			go rpc.ServeConn(conn)
		}
	}()

	time.Sleep(100 * time.Millisecond)

	// Parte Cliente
	client, err := rpc.Dial("tcp", rpcDir)
	if err != nil {
		log.Fatal("dialing:", err)
	}

	var reply int
	err = rpctimeout.CallTimeout(client, "NodoRaft.SometerOperacion", &Args{"rep"},
		&reply, 5*time.Millisecond)

	if err != nil {
		log.Fatal("arith error:", err)
	}

	//fmt.Println("Arith: %d*%d=%d", args.A, args.B, reply)
}*/
