/*
* AUTOR: Rafael Tolosana Calasanz
* ASIGNATURA: 30221 Sistemas Distribuidos del Grado en Ingeniería Informática
*			Escuela de Ingeniería y Arquitectura - Universidad de Zaragoza
* FECHA: septiembre de 2021
* FICHERO: ricart-agrawala.go
* DESCRIPCIÓN: Implementación del algoritmo de Ricart-Agrawala Generalizado en Go
 */
package raGoVec2ms

import (
	"fmt"
	"practica2/ms"
	"sync"

	"github.com/DistributedClocks/GoVector/govec"
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
	MsLog     *ms.MessageSystem
	Done      chan bool
	Chrep     chan bool
	Mutex     sync.Mutex // mutex para proteger concurrencia sobre las variables
	Exclude   [2][2]bool // [{read,read},{read,write}] [{write,read} {write,write}]
	N         int        // Numero de nodos en la red
	Me        int        // Identificador del proceso
	OpType    int        // 0 -> read, 1 -> write
	Log       *govec.GoLog
}

func New(me int, usersFile string, N int, opType int, Logger *govec.GoLog) *RASharedDB {
	messageTypes := []ms.Message{Request{}, Reply{}, ms.Escribir{}, ms.Leer{}}
	msgs := ms.New(me, usersFile, messageTypes)
	messageTypesLog := []ms.Message{ms.GoVecMsg{}}
	msgsLog := ms.New(me, "./ms/usersLog.txt", messageTypesLog)
	var ExcludeAux [2][2]bool
	ExcludeAux[0][0] = false
	ExcludeAux[0][1] = true
	ExcludeAux[1][0] = true
	ExcludeAux[1][1] = true
	ra := RASharedDB{0, 0, 0, false, []bool{}, &msgs, &msgsLog, make(chan bool), make(chan bool),
		sync.Mutex{}, ExcludeAux, N, me, opType, Logger}
	for i := 0; i < ra.N; i++ {
		ra.RepDefd = append(ra.RepDefd, false)
	}
	return &ra
}

//Pre: Verdad
//Post: Realiza  el  PreProtocol  para el  algoritmo de
//      Ricart-Agrawala Generalizado
func (ra *RASharedDB) PreProtocol() {
	//Traduccion literal del algoritmo en ALGOL
	//fmt.Println("Entra al PREprotocol")
	ra.Mutex.Lock()
	ra.ReqCS = true
	ra.OurSeqNum = ra.HigSeqNum + 1
	ra.Mutex.Unlock()
	ra.OutRepCnt = ra.N - 1
	for j := 1; j <= ra.N; j++ {
		if j != ra.Me {
			ra.Ms.Send(j, Request{ra.OurSeqNum, ra.Me, ra.OpType})
			msgBytes := []byte("Request")
			logMsg := ra.Log.PrepareSend("Sending message REQ", msgBytes, govec.GetDefaultLogOptions())
			ra.MsLog.Send(j, ms.GoVecMsg{Msg: logMsg})
		}
	}
	for ra.OutRepCnt != 0 {
		//fmt.Println("Esperando respuestas de todos")
		<-ra.Chrep // Se recibe respuesta por el canal de respuestas (no es necesario almacenar el valor de la respuesta en ninguna variable)
		ra.OutRepCnt--
	}
	//fmt.Println("Todas las respuestas recibidas")
}

//Pre: Verdad
//Post: Realiza  el  PostProtocol  para el  algoritmo de
//      Ricart-Agrawala Generalizado
func (ra *RASharedDB) PostProtocol() {
	//fmt.Println("Entra al POSTprotocol")
	ra.ReqCS = false
	for j := 1; j <= ra.N; j++ {
		if ra.RepDefd[j-1] {
			ra.RepDefd[j-1] = false
			ra.Ms.Send(j, Reply{})
			msgBytes := []byte("Reply")
			logMsg := ra.Log.PrepareSend("Sending message REP", msgBytes, govec.GetDefaultLogOptions())
			ra.MsLog.Send(j, ms.GoVecMsg{Msg: logMsg})
		}
	}
}

func (ra *RASharedDB) Stop() {
	ra.Ms.Stop()
	ra.Done <- true // Despues del preprotocol se pasa a la seccion critica
}

func max(x, y int) int {
	if x > y {
		return x
	} else {
		return y
	}
}

//Process wich recieves request,response
func (ra *RASharedDB) GestionReqRes() {
	defer_it := false
	for {
		//Se recibe la peticion
		msg := ra.Ms.Receive()
		req, ok := msg.(Request)

		//Se recibe por el canal de Log
		msgLog := ra.MsLog.Receive()
		reqLog, okLog := msgLog.(ms.GoVecMsg)
		if okLog {
			//fmt.Println("Se ha recibido peticion LOG")
			ra.Log.UnpackReceive("Received Message ", reqLog.Msg, nil, govec.GetDefaultLogOptions())
		}

		if ok {
			//fmt.Println("Se ha recibido peticion REQUEST")
			ra.HigSeqNum = max(ra.HigSeqNum, req.Clock)
			ra.Mutex.Lock()
			defer_it = ra.ReqCS &&
				((req.Clock > ra.OurSeqNum) || (req.Clock == ra.OurSeqNum && req.Pid > ra.Me)) &&
				ra.Exclude[ra.OpType][req.OpType]
			ra.Mutex.Unlock()
			if defer_it {
				fmt.Println("DEFER IT")
				ra.RepDefd[req.Pid-1] = true
			} else {
				ra.Ms.Send(req.Pid, Reply{})
				msgBytes := []byte("Reply")
				logMsg := ra.Log.PrepareSend("Reply request", msgBytes, govec.GetDefaultLogOptions())
				ra.MsLog.Send(req.Pid, ms.GoVecMsg{Msg: logMsg})
			}
		} else {
			//fmt.Println("Se ha recibido peticion REPLY")
			ra.Chrep <- true
		}
	}
}
