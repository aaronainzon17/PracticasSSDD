// Escribir vuestro código de funcionalidad Raft en este fichero
//

package raft

//
// API
// ===
// Este es el API que vuestra implementación debe exportar
//
// nodoRaft = NuevoNodo(...)
//   Crear un nuevo servidor del grupo de elección.
//
// nodoRaft.Para()
//   Solicitar la parado de un servidor
//
// nodo.ObtenerEstado() (yo, mandato, esLider)
//   Solicitar a un nodo de elección por "yo", su mandato en curso,
//   y si piensa que es el msmo el lider
//
// nodoRaft.SometerOperacion(operacion interface()) (indice, mandato, esLider)

// type AplicaOperacion

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"

	//"crypto/rand"
	"sync"
	"time"

	//"net/rpc"
	"raft/internal/comun/rpctimeout"
)

const (
	C = "candidate"
	L = "leader"
	F = "follower"
)

const (
	// Constante para fijar valor entero no inicializado
	IntNOINICIALIZADO = -1

	//  false deshabilita por completo los logs de depuracion
	// Aseguraros de poner kEnableDebugLogs a false antes de la entrega
	kEnableDebugLogs = true

	// Poner a true para logear a stdout en lugar de a fichero
	kLogToStdout = false

	// Cambiar esto para salida de logs en un directorio diferente
	kLogOutputDir = "./logs_raft/"
)

type TipoOperacion struct {
	Operacion string // La operaciones posibles son "leer" y "escribir"
	Clave     string
	Valor     string // en el caso de la lectura Valor = ""
}

// A medida que el nodo Raft conoce las operaciones de las  entradas de registro
// comprometidas, envía un AplicaOperacion, con cada una de ellas, al canal
// "canalAplicar" (funcion NuevoNodo) de la maquina de estados
type AplicaOperacion struct {
	Indice    int // en la entrada de registro
	Operacion TipoOperacion
}

// Tipo de dato Go que representa un solo nodo (réplica) de raft
//
type NodoRaft struct {
	Mux sync.Mutex // Mutex para proteger acceso a estado compartido

	// Host:Port de todos los nodos (réplicas) Raft, en mismo orden
	Nodos   []rpctimeout.HostPort
	Yo      int // indice de este nodos en campo array "nodos"
	IdLider int
	// Utilización opcional de este logger para depuración
	// Cada nodo Raft tiene su propio registro de trazas (logs)
	Logger *log.Logger

	// Vuestros datos aqui.
	State
	// mirar figura 2 para descripción del estado que debe mantenre un nodo Raft
}

type State struct {
	CurrentTerm int //Mandato actual
	VotedFor    int //Candidato que ha recibido el voto en el mandto actual
	log         []LogEntry

	//For all servers
	CommitIndex int //Índice de la última entrada cometida
	LastApplied int //Ultima entrada del log aplicada en la máquina de estados

	// Datos auxiliares de cada nodo
	StateNode          string    // Leader, Follower, Candidate
	electionResetEvent time.Time // Last heart beat

	//Only for leaders
	CanalAplicar chan AplicaOperacion
	NextIndex    []int
	MatchIndex   []int
}

type LogEntry struct {
	Command interface{}
	Term    int
}

// Creacion de un nuevo nodo de eleccion
//
// Tabla de <Direccion IP:puerto> de cada nodo incluido a si mismo.
//
// <Direccion IP:puerto> de este nodo esta en nodos[yo]
//
// Todos los arrays nodos[] de los nodos tienen el mismo orden

// canalAplicar es un canal donde, en la practica 5, se recogerán las
// operaciones a aplicar a la máquina de estados. Se puede asumir que
// este canal se consumira de forma continúa.
//
// NuevoNodo() debe devolver resultado rápido, por lo que se deberían
// poner en marcha Gorutinas para trabajos de larga duracion
func NuevoNodo(nodos []rpctimeout.HostPort, yo int,
	canalAplicarOperacion chan AplicaOperacion) *NodoRaft {
	nr := &NodoRaft{}

	if kEnableDebugLogs {
		nombreNodo := nodos[yo].Host() + "_" + nodos[yo].Port()
		logPrefix := fmt.Sprintf("%s", nombreNodo)

		fmt.Println("LogPrefix: ", logPrefix)

		if kLogToStdout {
			nr.Logger = log.New(os.Stdout, nombreNodo+" -->> ",
				log.Lmicroseconds|log.Lshortfile)
		} else {
			err := os.MkdirAll(kLogOutputDir, os.ModePerm)
			if err != nil {
				panic(err.Error())
			}
			logOutputFile, err := os.OpenFile(fmt.Sprintf("%s/%s.txt",
				kLogOutputDir, logPrefix), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
			if err != nil {
				panic(err.Error())
			}
			nr.Logger = log.New(logOutputFile,
				logPrefix+" -> ", log.Lmicroseconds|log.Lshortfile)
		}
		nr.Logger.Println("logger initialized")
	} else {
		nr.Logger = log.New(ioutil.Discard, "", 0)
	}

	// Añadir codigo de inicialización
	nr.Nodos = nodos
	nr.Yo = yo
	nr.IdLider = IntNOINICIALIZADO
	nr.CurrentTerm = 0
	nr.VotedFor = IntNOINICIALIZADO
	nr.CommitIndex = IntNOINICIALIZADO
	nr.LastApplied = IntNOINICIALIZADO
	nr.StateNode = F
	nr.CanalAplicar = canalAplicarOperacion
	nr.NextIndex = make([]int, len(nodos))
	nr.MatchIndex = make([]int, len(nodos))
	nr.electionResetEvent = time.Now()
	go func() { time.Sleep(2000 * time.Millisecond); go nr.gestionDeLider() }()
	return nr
}

// Metodo Para() utilizado cuando no se necesita mas al nodo
//
// Quizas interesante desactivar la salida de depuracion
// de este nodo
//
func (nr *NodoRaft) para() {
	go func() { time.Sleep(5 * time.Millisecond); os.Exit(0) }()
}

// Devuelve "yo", mandato en curso y si este nodo cree ser lider
//
// Primer valor devuelto es el indice de este  nodo Raft el el conjunto de nodos
// la operacion si consigue comprometerse.
// El segundo valor es el mandato en curso
// El tercer valor es true si el nodo cree ser el lider
// Cuarto valor es el lider, es el indice del líder si no es él
func (nr *NodoRaft) obtenerEstado() (int, int, bool, int) {
	nr.Mux.Lock()
	yo := nr.Yo
	mandato := nr.CurrentTerm
	idLider := nr.IdLider
	esLider := false
	if nr.StateNode == L {
		esLider = true
	}
	nr.Mux.Unlock()
	return yo, mandato, esLider, idLider
}

// El servicio que utilice Raft (base de datos clave/valor, por ejemplo)
// Quiere buscar un acuerdo de posicion en registro para siguiente operacion
// solicitada por cliente.

// Si el nodo no es el lider, devolver falso
// Sino, comenzar la operacion de consenso sobre la operacion y devolver en
// cuanto se consiga
//
// No hay garantia que esta operacion consiga comprometerse en una entrada de
// de registro, dado que el lider puede fallar y la entrada ser reemplazada
// en el futuro.
// Primer valor devuelto es el indice del registro donde se va a colocar
// la operacion si consigue comprometerse.
// El segundo valor es el mandato en curso
// El tercer valor es true si el nodo cree ser el lider
// Cuarto valor es el lider, es el indice del líder si no es él
func (nr *NodoRaft) someterOperacion(operacion TipoOperacion) (int, int,
	bool, int, string) {
	indice := IntNOINICIALIZADO
	mandato := IntNOINICIALIZADO
	EsLider := false
	idLider := nr.IdLider
	valorADevolver := ""

	// Vuestro codigo aqui
	nr.Mux.Lock()
	if nr.StateNode == L {
		nr.log = append(nr.log, LogEntry{operacion.Operacion, nr.CurrentTerm})
		fmt.Println(nr.log)
		indice = nr.NextIndex[nr.Yo]
		mandato = nr.CurrentTerm
		EsLider = true
		for i := range nr.Nodos {
			if i != nr.Yo {
				go nr.submit(i)
			}
		}
		idLider = IntNOINICIALIZADO
	}
	nr.Mux.Unlock()
	return indice, mandato, EsLider, idLider, valorADevolver
}

// -----------------------------------------------------------------------
// LLAMADAS RPC al API
//
// Si no tenemos argumentos o respuesta estructura vacia (tamaño cero)
type Vacio struct{}

func (nr *NodoRaft) ParaNodo(args Vacio, reply *Vacio) error {
	defer nr.para()
	return nil
}

type EstadoParcial struct {
	Mandato int
	EsLider bool
	IdLider int
}

type EstadoRemoto struct {
	IdNodo int
	EstadoParcial
}

func (nr *NodoRaft) ObtenerEstadoNodo(args Vacio, reply *EstadoRemoto) error {
	reply.IdNodo, reply.Mandato, reply.EsLider, reply.IdLider = nr.obtenerEstado()
	return nil
}

type ResultadoRemoto struct {
	ValorADevolver string
	IndiceRegistro int
	EstadoParcial
}

func (nr *NodoRaft) SometerOperacionRaft(operacion TipoOperacion,
	reply *ResultadoRemoto) error {
	reply.IndiceRegistro, reply.Mandato, reply.EsLider,
		reply.IdLider, reply.ValorADevolver = nr.someterOperacion(operacion)
	return nil
}

// -----------------------------------------------------------------------
// LLAMADAS RPC protocolo RAFT
//
// Structura de ejemplo de argumentos de RPC PedirVoto.
//
// Recordar
// -----------
// Nombres de campos deben comenzar con letra mayuscula !
//
type ArgsPeticionVoto struct {
	Term         int
	CandidateId  int
	LastLogIndex int
	LastLogTerm  int
}

// Structura de ejemplo de respuesta de RPC PedirVoto,
//
// Recordar
// -----------
// Nombres de campos deben comenzar con letra mayuscula !
//
//
type RespuestaPeticionVoto struct {
	Term        int
	VoteGranted bool
}

// Metodo para RPC PedirVoto
//
func (nr *NodoRaft) PedirVoto(peticion *ArgsPeticionVoto,
	reply *RespuestaPeticionVoto) error {
	nr.Mux.Lock()
	lastLogIndex, lastLogTerm := nr.getLastLogData()
	if peticion.Term > nr.CurrentTerm {
		nr.becomeFollower(peticion.Term)
	}
	if peticion.Term < nr.CurrentTerm {
		reply.VoteGranted = false
	}
	if nr.CurrentTerm == peticion.Term &&
		(nr.VotedFor == IntNOINICIALIZADO || nr.VotedFor == peticion.CandidateId) &&
		(peticion.LastLogIndex >= lastLogIndex &&
			peticion.LastLogTerm == lastLogTerm) {
		reply.VoteGranted = true
		nr.VotedFor = peticion.CandidateId
		nr.electionResetEvent = time.Now()
	} else {
		reply.VoteGranted = false
	}
	reply.Term = nr.CurrentTerm
	nr.Mux.Unlock()

	return nil
}

type ArgAppendEntries struct {
	Term     int
	LeaderId int

	PrevLogIndex int
	PrevLogTerm  int
	Entries      []LogEntry
	LeaderCommit int
}

type Results struct {
	Term    int
	Success bool
}

// Metodo de tratamiento de llamadas RPC AppendEntries
func (nr *NodoRaft) AppendEntries(args *ArgAppendEntries,
	results *Results) error {
	nr.Mux.Lock()
	//Si el mandato esta desactualizado (elecciones)
	if args.Term > nr.CurrentTerm {
		nr.becomeFollower(args.Term)
	}
	//Si el mandato es menor rechaza la peticion
	results.Success = false
	if args.Term == nr.CurrentTerm {
		nr.electionResetEvent = time.Now()
		if args.PrevLogIndex == IntNOINICIALIZADO || (len(nr.log) > args.PrevLogIndex &&
			nr.log[args.PrevLogIndex].Term == args.PrevLogTerm) {
			results.Success = true
			if len(args.Entries) > 0 {
				//Se puede hacer un bucle para comprobar el punto de insercion
				fmt.Println("Voy a ver que hay hasta args.PrevLogIndex+1", args.PrevLogIndex+1)
				fmt.Println(nr.log[:args.PrevLogIndex+1])
				nr.log = append(nr.log[:args.PrevLogIndex+1], args.Entries...)
				fmt.Println("Se ha almacenado una op en el LOG")
				fmt.Println(nr.log)
			}
		}
		//Si se ha actualizado el CommitIndex del lider y mi logger tambien => actualizo mi commit index
		if args.LeaderCommit > nr.CommitIndex && args.LeaderCommit == len(nr.log) {
			nr.CommitIndex = args.LeaderCommit
		}
		results.Term = nr.CurrentTerm
	}
	nr.Mux.Unlock()

	return nil
}

// ----- Metodos/Funciones a utilizar como clientes
//
//

// Ejemplo de código enviarPeticionVoto
//
// nodo int -- indice del servidor destino en nr.nodos[]
//
// args *RequestVoteArgs -- argumentos par la llamada RPC
//
// reply *RequestVoteReply -- respuesta RPC
//
// Los tipos de argumentos y respuesta pasados a CallTimeout deben ser
// los mismos que los argumentos declarados en el metodo de tratamiento
// de la llamada (incluido si son punteros
//
// Si en la llamada RPC, la respuesta llega en un intervalo de tiempo,
// la funcion devuelve true, sino devuelve false
//
// la llamada RPC deberia tener un timout adecuado.
//
// Un resultado falso podria ser causado por una replica caida,
// un servidor vivo que no es alcanzable (por problemas de red ?),
// una petición perdida, o una respuesta perdida
//
// Para problemas con funcionamiento de RPC, comprobar que la primera letra
// del nombre  todo los campos de la estructura (y sus subestructuras)
// pasadas como parametros en las llamadas RPC es una mayuscula,
// Y que la estructura de recuperacion de resultado sea un puntero a estructura
// y no la estructura misma.
//
func (nr *NodoRaft) enviarPeticionVoto(nodo int, args *ArgsPeticionVoto,
	reply *RespuestaPeticionVoto) bool {
	var ok bool
	err := nr.Nodos[nodo].CallTimeout("NodoRaft.PedirVoto", args,
		&reply, 50*time.Millisecond)

	if err != nil {
		ok = false
	} else {
		ok = true
	}

	return ok
}

//----------------------------------------------------------------------------
//FUNCIONES ASOCIADAS A ESTADOS

/********************			 SEGUIDOR			***************************
* Funcion encargada de la gestion del lider y de iniciar un proceso de eleccion
* si no se reciben mensajes de nadie durante un tiempo.
**/
func (nr *NodoRaft) gestionDeLider() {
	// Uso una semilla porque sino genera la misma secuencia de
	// numeros aleatorios para todos
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	timeout := time.Duration(150+r1.Intn(150)) * time.Millisecond
	tick := time.NewTicker(20 * time.Millisecond)
	eleccion := false
	// Mientras no se haya iniciado una eleccion se comprueba en cada tick si el
	// tiempo transcurrido desde el ultimo latido es mayor que el timeout, si lo
	// es, se incia una votacion.
	for !eleccion {
		<-tick.C
		//nr.mux.Lock()
		//Se comprueba si el tiempo que ha pasado es > timeout
		elapsed := time.Since(nr.electionResetEvent)
		if elapsed >= timeout && nr.StateNode == F {
			nr.elecciones()
			//nr.mux.Unlock()
			eleccion = true
		}
	}
}

/*
* Funcion encargada de convertir a un nodo seguidor del mandato que se envia
* por parametro
**/
func (nr *NodoRaft) becomeFollower(term int) {
	nr.StateNode = F
	nr.CurrentTerm = term
	nr.VotedFor = IntNOINICIALIZADO
	nr.electionResetEvent = time.Now()

	go nr.gestionDeLider()
}

/********************			  LIDER				***************************
* Funcion encargada de inicializar un lider y enviar latidos periodicamente
**/
func (nr *NodoRaft) becomeLeader(term int) {
	fmt.Println("SOY EL LIDER DEL MANDATO ", term)
	nr.Mux.Lock()
	nr.StateNode = L
	//Se inicializan NextIndex y MatchIndex
	for i := 0; i < len(nr.Nodos); i++ {
		nr.NextIndex[i] = len(nr.log)
		nr.MatchIndex[i] = -1
	}
	nr.Mux.Unlock()
	tick := time.NewTicker(50 * time.Millisecond)
	defer tick.Stop()
	for nr.StateNode == L {
		<-tick.C
		nr.sendHeartBeat()
	}
}

/*
* Funcion encargada de enviar latidos a todos los nodos
**/
func (nr *NodoRaft) sendHeartBeat() {
	for i := range nr.Nodos {
		if i != nr.Yo {
			args := ArgAppendEntries{Term: nr.CurrentTerm, LeaderId: nr.Yo}
			var reply Results
			err := nr.Nodos[i].CallTimeout("NodoRaft.AppendEntries", args,
				&reply, 5*time.Second)
			if err != nil {
				fmt.Println("Cant reach node ", i)
			}
		}
	}
}

/********************    CANDIDATO(ELECCIONES)    ***************************
* Funcion encargada de comenzar una eleccion de lider en el momento en el que
* a un seguidor le ha saltado el timeout del lider.
**/
func (nr *NodoRaft) elecciones() {
	nr.CurrentTerm = nr.CurrentTerm + 1
	nr.StateNode = C
	nr.VotedFor = nr.Yo
	nr.electionResetEvent = time.Now()

	votos, exito := nr.hacerElecciones(nr.CurrentTerm)
	if exito {
		// votos >= (N + 1)/2
		if votos*2 >= len(nr.Nodos)+1 {
			// Gana la eleccion y se convierte en lider
			fmt.Println("HE GANADO CON ", votos, " VOTOS")
			go nr.becomeLeader(nr.CurrentTerm)
		} else {
			go nr.gestionDeLider()
		}
	}
}

/*
* Funcion auxiliar de elecciones, encargada de enviar las peticiones rpc a las
* replicas y contabilizar los votos
**/
func (nr *NodoRaft) hacerElecciones(storedTerm int) (int, bool) {
	votos := 1
	exito := true
	for i := range nr.Nodos {
		if i != nr.Yo {
			// Si ha llegado una llamada de un lider y ha cambiado a seguidor
			// se detiene la votacion
			if nr.StateNode != C {
				fmt.Println("Estoy en elecciones y no soy candidato ME VOY")
				exito = false
				break
			}
			nr.Mux.Lock()
			lastLogIndex, lastLogTerm := nr.getLastLogData()
			nr.Mux.Unlock()
			args := ArgsPeticionVoto{Term: storedTerm, CandidateId: nr.Yo,
				LastLogIndex: lastLogIndex, LastLogTerm: lastLogTerm}
			var reply RespuestaPeticionVoto

			if res := nr.enviarPeticionVoto(i, &args, &reply); res {
				if storedTerm < reply.Term {
					nr.becomeFollower(reply.Term)
					exito = false
					break
				}
				if (reply.Term == storedTerm) && reply.VoteGranted {
					votos += 1
				}
			}
		}
	}
	return votos, exito
}

/***********************     FUNCIONES AUXILIARES     ***********************/

/*
* Esta funcion es necesaria porque en el caso de que el logger este vacio,
* se daria un error al intentar obtener el LastLogTerm.
**/
func (nr *NodoRaft) getLastLogData() (int, int) {
	if len(nr.log) > 0 {
		return len(nr.log) - 1, nr.log[len(nr.log)-1].Term
	} else {
		return -1, -1
	}
}

//Funcion encargada de enviar latidos a todos los nodos
func (nr *NodoRaft) submit(i int) {
	done := false
	for !done {
		args := nr.makeAppendEntriesArgs(i)
		var reply Results
		err := nr.Nodos[i].CallTimeout("NodoRaft.AppendEntries", args,
			&reply, 40*time.Millisecond)
		if err == nil {
			nr.Mux.Lock()
			if reply.Success {
				nr.checkReply(reply, i, len(args.Entries))
				done = true
			} else {
				if nr.NextIndex[i] > 0 {
					fmt.Println("ALGO NO COINCIDE CAPITAN")
					nr.NextIndex[i] = nr.NextIndex[i] - 1
				}
			}
			nr.Mux.Unlock()

		} else {
			fmt.Println("Cant reach node ", i, " ", err)
		}
	}
}

func (nr *NodoRaft) checkReply(reply Results, i int, lenEntries int) {
	fmt.Println("Voy a actualizar nextIndex de ", i)
	fmt.Println("NextIndex ini: ", nr.NextIndex[i])
	nr.NextIndex[i] = nr.NextIndex[i] + lenEntries
	fmt.Println("NextIndex fin: ", nr.NextIndex[i])
	nr.MatchIndex[i] = nr.NextIndex[i] - 1
	index := nr.CommitIndex + 1
	match := true
	for index < len(nr.log) && match {
		if nr.log[index].Term == nr.CurrentTerm {
			matchLog := 1
			for j := range nr.Nodos {
				if nr.MatchIndex[j] >= index {
					matchLog++
				}
			}
			if matchLog*2 >= len(nr.Nodos)+1 {
				nr.CommitIndex = index
				/*NO SE SI HACE FALTA
				aplica := AplicaOperacion{
					Indice:    index,
					Operacion: nr.log[index].Command,
				}
				fmt.Println("Se va a someter ", aplica)
				nr.CanalAplicar <- aplica*/
			} else {
				match = false
			}
		}
		index++
	}
}

func (nr *NodoRaft) makeAppendEntriesArgs(i int) ArgAppendEntries {
	nr.Mux.Lock()
	var prevLogIndex, prevLogTerm int
	nrIndex := nr.NextIndex[i]
	if nrIndex-1 >= 0 {
		prevLogIndex = nrIndex - 1
		prevLogTerm = nr.log[prevLogIndex].Term
	} else {
		prevLogTerm = -1
		prevLogIndex = -1
	}
	entries := nr.log[nrIndex:]
	args := ArgAppendEntries{
		Term: nr.CurrentTerm, LeaderId: nr.Yo, PrevLogIndex: prevLogIndex,
		PrevLogTerm: prevLogTerm, Entries: entries, LeaderCommit: nr.CommitIndex,
	}
	nr.Mux.Unlock()
	return args
}
