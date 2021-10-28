/*
* AUTOR: Rafael Tolosana Calasanz
* ASIGNATURA: 30221 Sistemas Distribuidos del Grado en Ingeniería Informática
*			Escuela de Ingeniería y Arquitectura - Universidad de Zaragoza
* FECHA: septiembre de 2021
* FICHERO: ms.go
* DESCRIPCIÓN: Implementación de un sistema de mensajería asíncrono, insipirado en el Modelo Actor
 */
package msGoVec

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"net"
	"os"
)

type Message interface{}

type MessageSystem struct {
	mbox  chan Message
	peers []string
	done  chan bool
	me    int
}

type Escribir struct {
	Fase   string
	OpType int
	Me     int
}

type Leer struct {
	Fase   string
	OpType int
	Me     int
}

type GoVecMsg struct {
	Msg []byte
}

const (
	MAXMESSAGES = 10000
)

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}

func parsePeers(path string) (lines []string) {
	file, err := os.Open(path)
	checkError(err)
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines
}

// Pre: pid en {1..n}, el conjunto de procesos del SD
// Post: envía el mensaje msg a pid
func (msGoVec *MessageSystem) Send(pid int, msg Message) {
	conn, err := net.Dial("tcp", msGoVec.peers[pid-1])
	checkError(err)
	encoder := gob.NewEncoder(conn)
	err = encoder.Encode(&msg)
	conn.Close()
}

// Pre: True
// Post: el mensaje msg de algún Proceso P_j se retira del mailbox y se devuelve
//		Si mailbox vacío, Receive bloquea hasta que llegue algún mensaje
func (msGoVec *MessageSystem) Receive() (msg Message) {
	msg = <-msGoVec.mbox
	return msg
}

func register(messageTypes []Message) {
	for _, msgTp := range messageTypes {
		gob.Register(msgTp)
	}
}

// Pre: whoIam es el pid del proceso que inicializa este ms
//		usersFile es la ruta a un fichero de texto que en cada línea contiene IP:puerto de cada participante
//		messageTypes es un slice con todos los tipos de mensajes que los procesos se pueden intercambiar a través de este ms
func New(whoIam int, usersFile string, messageTypes []Message) (msGoVec MessageSystem) {
	msGoVec.me = whoIam
	msGoVec.peers = parsePeers(usersFile)
	msGoVec.mbox = make(chan Message, MAXMESSAGES)
	msGoVec.done = make(chan bool)
	register(messageTypes)
	go func() {
		listener, err := net.Listen("tcp", msGoVec.peers[msGoVec.me-1])
		checkError(err)
		fmt.Println("Process listening at " + msGoVec.peers[msGoVec.me-1])
		defer close(msGoVec.mbox)
		for {
			select {
			case <-msGoVec.done:
				return
			default:
				conn, err := listener.Accept()
				checkError(err)
				decoder := gob.NewDecoder(conn)
				var msg Message
				err = decoder.Decode(&msg)
				conn.Close()
				msGoVec.mbox <- msg
			}
		}
	}()
	return msGoVec
}

//Pre: True
//Post: termina la ejecución de este ms
func (ms *MessageSystem) Stop() {
	ms.done <- true
}
