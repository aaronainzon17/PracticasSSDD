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
    "ms"
    "sync"
)

type Request struct{
    Clock   int
    Pid     int
}

type Reply struct{}

type RASharedDB struct {
    OurSeqNum   int     // Our sequence number
    HigSeqNum   int     // Higher sequence number
    OutRepCnt   int     // Outstanding reply count 
    ReqCS       boolean // Entiendo que es request critical section 
    RepDefd     int[]
    ms          *MessageSystem
    done        chan bool
    chrep       chan bool
    Mutex       sync.Mutex  // mutex para proteger concurrencia sobre las variables
    Exclude     bool[2][2]  // [{read,read},{read,write}] [{write,read} {write,write}]
    N           int         // Numero de nodos en la red
    Pid         int
}


func New(me int, usersFile string, N int, Pid int) (*RASharedDB) {
    messageTypes := []Message{Request, Reply}
    msgs = ms.New(me, usersFile string, messageTypes)
    ra := RASharedDB{0, 0, 0, false, []int{}, &msgs,  make(chan bool),  make(chan bool), &sync.Mutex{}, [false,true][false,false], N, Pid}
    // TODO completar
    return &ra
}

//Pre: Verdad
//Post: Realiza  el  PreProtocol  para el  algoritmo de
//      Ricart-Agrawala Generalizado
func (ra *RASharedDB) PreProtocol(){
    //Traduccion literal del algoritmo en ALGOL
    ra.Mutex.Lock()
    ra.ReqCS = true
    ra.OurSeqNum = ra.HigSeqNum + 1
    ra.Mutex.Unlock()
    ra.OutRepCnt = ra.N - 1
    for i := 1; i <= ra.N; i++ {
        ra.ms.Send(ra.Pid, "REQUEST")
    } 


}

//Pre: Verdad
//Post: Realiza  el  PostProtocol  para el  algoritmo de
//      Ricart-Agrawala Generalizado
func (ra *RASharedDB) PostProtocol(){
    // TODO completar
}

func (ra *RASharedDB) Stop(){
    ra.ms.Stop()
    ra.done <- true
}
