/*
* AUTOR: Rafael Tolosana Calasanz
* ASIGNATURA: 30221 Sistemas Distribuidos del Grado en Ingeniería Informática
*			Escuela de Ingeniería y Arquitectura - Universidad de Zaragoza
* FECHA: septiembre de 2021
* FICHERO: ricart-agrawala.go
* DESCRIPCIÓN: Implementación del algoritmo de Ricart-Agrawala Generalizado en Go
 */
package ra

import (
	"practica2/ms"
	"sync"
)

type Request struct {
	Clock  int
	Pid    int
	OpType int
}

type Reply struct{}

type RASharedDB struct {
	OurSeqNum int  // Our sequence number
	HigSeqNum int  // Higher sequence number
	OutRepCnt int  // Outstanding reply count
	ReqCS     bool // Request critical section
	RepDefd   []bool
	Ms        *ms.MessageSystem
	Done      chan bool
	Chrep     chan bool
	Mutex     sync.Mutex // mutex para proteger concurrencia sobre las variables
	Exclude   [2][2]bool // [{read,read},{read,write}] [{write,read} {write,write}]
	N         int        // Numero de nodos en la red
	Me        int        // Identificador del proceso
	OpType    int        // 0 -> read, 1 -> write
}

func New(me int, usersFile string, N int, opType int) *RASharedDB {
	messageTypes := []ms.Message{Request{}, Reply{}}
	msgs := ms.New(me, usersFile, messageTypes)
	ra := RASharedDB{0, 0, 0, false, []bool{}, &msgs, make(chan bool), make(chan bool),
		sync.Mutex{}, [2][2]bool{{false, true}, {true, true}}, N, me, opType}
	return &ra
}

//Pre: Verdad
//Post: Realiza  el  PreProtocol  para el  algoritmo de
//      Ricart-Agrawala Generalizado
func (ra *RASharedDB) PreProtocol() {
	//Traduccion literal del algoritmo en ALGOL
	ra.Mutex.Lock()
	ra.ReqCS = true
	ra.OurSeqNum = ra.HigSeqNum + 1
	ra.Mutex.Unlock()
	ra.OutRepCnt = ra.N - 1
	for j := 1; j <= ra.N; j++ {
		if j != ra.Me {
			ra.Ms.Send(j, Request{ra.OurSeqNum, ra.Me, ra.OpType}) //Puede que falten datos para el generalizado como el tipo de op y el reloj interno :))
		}
	}
	for ra.OutRepCnt != 0 { //Duda de si tiene que llegar la reply al channel booleano chrep
		<-ra.Chrep // Se recibe respuesta por el canal de respuestas (no es necesario almacenar el valor de la respuesta en ninguna variable)
		ra.OutRepCnt--
	}
	// Despues del preprotocol se pasa a la seccion critica
}

//Pre: Verdad
//Post: Realiza  el  PostProtocol  para el  algoritmo de
//      Ricart-Agrawala Generalizado
func (ra *RASharedDB) PostProtocol() {
	ra.ReqCS = false
	for j := 1; j <= ra.N; j++ {
		if ra.RepDefd[j-1] {
			ra.RepDefd[j-1] = false
			ra.Ms.Send(j, Reply{}) // Falta bullshit to send
		}
	}
}

func (ra *RASharedDB) Stop() {
	ra.Ms.Stop()
	ra.Done <- true
}

func max(x, y int) int {
	if x > y {
		return x
	} else {
		return y
	}
}

//Process wich recieves request,response
func (ra *RASharedDB) RecieveReqRes() {
	defer_it := false
	for {
		//Se recibe la peticion
		msg := ra.Ms.Receive()
		req, ok := msg.(Request)
		if ok {
			ra.HigSeqNum = max(ra.HigSeqNum, req.Clock)
			ra.Mutex.Lock()
			defer_it = ra.ReqCS &&
				((req.Clock > ra.HigSeqNum) || (req.Clock == ra.OurSeqNum && req.Pid > ra.Me)) &&
				ra.Exclude[ra.OpType][req.OpType] // Echarle un ojo al general
			ra.Mutex.Unlock()
			if defer_it {
				ra.RepDefd[req.Pid-1] = true
			} else {
				ra.Ms.Send(req.Pid, Reply{})
			}
		} else {
			ra.Chrep <- true
		}
	}
}
