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
	"net/rpc"
	"os"
	"raft/internal/comun/rpctimeout"
	"sync"
	"time"
)

const (
	C = "candidate"
	L = "leader"
	F = "follower"
)

//  false deshabilita por completo los logs de depuracion
// Aseguraros de poner kEnableDebugLogs a false antes de la entrega
const kEnableDebugLogs = true

// Poner a true para logear a stdout en lugar de a fichero
const kLogToStdout = true

// Cambiar esto para salida de logs en un directorio diferente
const kLogOutputDir = "./logs_raft/"

// A medida que el nodo Raft conoce las operaciones de las  entradas de registro
// comprometidas, envía un AplicaOperacion, con cada una de ellas, al canal
// "canalAplicar" (funcion NuevoNodo) de la maquina de estados
type AplicaOperacion struct {
	Indice    int // en la entrada de registro
	Operacion interface{}
}

// Tipo de dato Go que representa un solo nodo (réplica) de raft
//
type NodoRaft struct {
	mux sync.Mutex // Mutex para proteger acceso a estado compartido

	nodos []*rpc.Client //Conexiones RPC a todos los nodos (réplicas) Raft
	yo    int           // this peer's index into peers[]

	// Utilización opcional de este logger para depuración
	// Cada nodo Raft tiene su propio registro de trazas (logs)
	logger *log.Logger

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
	NextIndex  []int
	MatchIndex []int
}

type LogEntry struct {
	Command interface{}
	Term    int
}

type CommitEntry struct {
	Command interface{}
	Term    int
	Index   int
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
func NuevoNodo(nodos []string, yo int, canalAplicar chan AplicaOperacion) *NodoRaft {
	nr := &NodoRaft{}
	// /nr.nodos = nodos
	nr.yo = yo

	if kEnableDebugLogs {
		nombreNodo := nodos[yo]
		logPrefix := fmt.Sprintf("%s ", nombreNodo)
		if kLogToStdout {
			nr.logger = log.New(os.Stdout, nombreNodo,
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
			nr.logger = log.New(logOutputFile, logPrefix, log.Lmicroseconds|log.Lshortfile)
		}
		nr.logger.Println("logger initialized")
	} else {
		nr.logger = log.New(ioutil.Discard, "", 0)
	}

	// Your initialization code here (2A, 2B)

	//Inicializamos los valores necesarios para las votaciones
	nr.CurrentTerm = 0
	nr.VotedFor = -1
	nr.StateNode = F
	nr.electionResetEvent = time.Now()
	fmt.Println("Soy el nodo: ", nr.yo, "y soy ", nr.StateNode)

	return nr
}

func (nr *NodoRaft) ConnectNodes(nodes []string) {
	for i := 0; i < len(nodes); i++ {
		if i != nr.yo {
			rpcConn, err := rpc.DialHTTP("tcp", nodes[i])
			if err != nil {
				panic(err.Error())
			}
			nr.nodos = append(nr.nodos, rpcConn)
		}
	}

	for _, node := range nr.nodos {
		fmt.Println("CONOZCO A:", node)
	}

	go nr.gestionDeLider()
}

// Metodo Para() utilizado cuando no se necesita mas al nodo
//
// Quizas interesante desactivar la salida de depuracion
// de este nodo
//
func (nr *NodoRaft) Para() {
	// Vuestro codigo aqui
	os.Exit(1)
}

// Devuelve "yo", mandato en curso y si este nodo cree ser lider
//
func (nr *NodoRaft) ObtenerEstado() (int, int, bool) {
	var yo int
	var mandato int
	var esLider bool

	// Vuestro codigo aqui
	nr.mux.Lock()
	yo = nr.yo
	mandato = nr.CurrentTerm
	if nr.StateNode == L {
		esLider = true
	} else {
		esLider = false
	}
	nr.mux.Unlock()

	return yo, mandato, esLider
}

// El servicio que utilice Raft (base de datos clave/valor, por ejemplo)
// Quiere buscar un acuerdo de posicion en registro para siguiente operacion
// solicitada por cliente.

// Si el nodo no es el lider, devolver falso
// Sino, comenzar la operacion de consenso sobre la operacion y devolver con
// rapidez
//
// No hay garantia que esta operacion consiga comprometerse n una entrada de
// de registro, dado que el lider puede fallar y la entrada ser reemplazada
// en el futuro.
// Primer valor devuelto es el indice del registro donde se va a colocar
// la operacion si consigue comprometerse.
// El segundo valor es el mandato en curso
// El tercer valor es true si el nodo cree ser el lider
func (nr *NodoRaft) SometerOperacion(operacion interface{}) (int, int, bool) {
	indice := -1
	mandato := -1
	EsLider := false

	// Vuestro codigo aqui
	if nr.StateNode == L {
		nr.log = append(nr.log, LogEntry{operacion, nr.CurrentTerm})
		indice = nr.NextIndex[nr.yo]
		mandato = nr.CurrentTerm
		EsLider = true
		//REVISAR QUE SOBRA (CREO QUE TODA LA OPERACION)
		for i, nodo := range nr.nodos {
			go nr.comprometerOperacion(i, nodo)
		}
	}

	return indice, mandato, EsLider
}

func (nr *NodoRaft) comprometerOperacion(index int, nodo *rpc.Client) bool {
	commited := false
	nr.mux.Lock()
	var prevLogIndex, prevLogTerm int
	nrIndex := nr.NextIndex[index]
	if len(nr.log) > 0 {
		prevLogIndex = nrIndex - 1
		prevLogTerm = nr.log[prevLogIndex].Term
	} else {
		prevLogTerm = -1
		prevLogIndex = -1
	}
	entries := nr.log[nrIndex:]
	args := ArgsAppendEntries{
		Term:         nr.CurrentTerm,
		LeaderId:     nr.yo,
		PrevLogIndex: prevLogIndex,
		PrevLogTerm:  prevLogTerm,
		Entries:      entries,
		LeaderCommit: nr.CommitIndex,
	}
	nr.mux.Unlock()
	for !commited {
		var reply RespuestaAppendEntries
		err := rpctimeout.CallTimeout(nodo, "NodoRaft.AppendEntries", args,
			&reply, 50*time.Millisecond)
		if err != nil {
			fmt.Println("Cant reach node ", index)
		}
	}
	return commited
}

//
// ArgsPeticionVoto
// ===============
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

//
// RespuestaPeticionVoto
// ================
//
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

//
// PedirVoto
// ===========
//
// Metodo para RPC PedirVoto
//
func (nr *NodoRaft) PedirVoto(args *ArgsPeticionVoto,
	reply *RespuestaPeticionVoto) error {
	nr.mux.Lock()
	lastLogIndex, lastLogTerm := nr.getLastLogData()
	if args.Term > nr.CurrentTerm {
		nr.becomeFollower(args.Term)
	}
	if args.Term < nr.CurrentTerm {
		reply.VoteGranted = false
	}
	if nr.CurrentTerm == args.Term &&
		(nr.VotedFor == -1 || nr.VotedFor == args.CandidateId) &&
		(args.LastLogIndex >= lastLogIndex &&
			args.LastLogTerm == lastLogTerm) {
		reply.VoteGranted = true
		nr.VotedFor = args.CandidateId
		nr.electionResetEvent = time.Now()
	} else {
		reply.VoteGranted = false
	}
	reply.Term = nr.CurrentTerm
	nr.mux.Unlock()
	return nil
}

// Ejemplo de código enviarPeticionVoto
//
// nodo int -- indice del servidor destino en nr.nodos[]
//
// args *RequestVoteArgs -- argumetnos par la llamada RPC
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
// una petiión perdida, o una respuesta perdida
//
// Para problemas con funcionamiento de RPC, comprobar que la primera letra
// del nombre  todo los campos de la estructura (y sus subestructuras)
// pasadas como parametros en las llamadas RPC es una mayuscula,
// Y que la estructura de recuperacion de resultado sea un puntero a estructura
// y no la estructura misma.
//

func (nr *NodoRaft) enviarPeticionVoto(nodo *rpc.Client, args *ArgsPeticionVoto,
	reply *RespuestaPeticionVoto) bool {
	var ok bool
	err := rpctimeout.CallTimeout(nodo, "NodoRaft.PedirVoto", args,
		&reply, 50*time.Millisecond)

	if err != nil {
		ok = false
	} else {
		ok = true
	}

	return ok
}

type ArgsAppendEntries struct {
	Term     int
	LeaderId int

	PrevLogIndex int
	PrevLogTerm  int
	Entries      []LogEntry
	LeaderCommit int
}

type RespuestaAppendEntries struct {
	Term    int
	Success bool
}

func (nr *NodoRaft) AppendEntries(args *ArgsAppendEntries,
	reply *RespuestaAppendEntries) error {
	nr.mux.Lock()
	fmt.Println("hb")
	//Actualizo el momento del ultimo latido
	//Si el mandato esta desactualizado (elecciones)
	if args.Term > nr.CurrentTerm {
		nr.becomeFollower(args.Term)
	}

	//Si el mandato es menor rechaza la peticion
	reply.Success = false
	if args.Term == nr.CurrentTerm {
		reply.Success = true
	} else if args.PrevLogIndex > -1 && len(nr.log) > args.PrevLogIndex &&
		nr.log[args.PrevLogIndex].Term == args.PrevLogTerm {
		reply.Success = true
		// Aqui va todo de locos asi que mete las entradas
		//Faltan comprobaciones aunque son P5
		nr.log = append(nr.log[:args.PrevLogIndex+1], args.Entries...)
	}

	nr.electionResetEvent = time.Now()
	reply.Term = nr.CurrentTerm
	nr.mux.Unlock()
	return nil
}

// Crear una gorutina concurrente que se responsabilice de la gestión del líder,
// y ponga en marcha un proceso de elección si no recibe mensajes de nadie
// durante un tiempo. De esta forma podrá saber quien es el líder, si ya lo hay,
// o convertirse el mismo en líder.

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
		// Start an election if we haven't heard from a leader or haven't
		// voted for someone for the duration of the timeout.
		elapsed := time.Since(nr.electionResetEvent)
		if elapsed >= timeout && nr.StateNode == F {
			nr.elecciones()
			eleccion = true
		}
	}
}

func (nr *NodoRaft) elecciones() {
	// El nodo cambia a Candidato e incrementa el mandato
	nr.mux.Lock() //Creo que el mutex no hace falta porque no se lanza una gorutina
	nr.CurrentTerm = nr.CurrentTerm + 1
	storedTerm := nr.CurrentTerm
	nr.StateNode = C
	// Se vota a si mismo
	nr.VotedFor = nr.yo

	nr.electionResetEvent = time.Now()
	nr.mux.Unlock()
	votos := 1

	//RequestVote RPCs in parallel to each of the other servers in the cluster.
	for _, nodo := range nr.nodos {
		// Si ha llegado una llamada de un lider y ha cambiado a seguidor
		// se detiene la votacion
		if nr.StateNode != C {
			fmt.Println("Estoy en elecciones y no soy candidato ME VOY")
			break
		}
		args := ArgsPeticionVoto{Term: storedTerm, CandidateId: nr.yo,
			LastLogIndex: len(nr.log) - 1, LastLogTerm: nr.log[len(nr.log)-1].Term}
		var reply RespuestaPeticionVoto

		if res := nr.enviarPeticionVoto(nodo, &args, &reply); res {
			if storedTerm < reply.Term {
				nr.becomeFollower(reply.Term)
				break
			}
			if (reply.Term == storedTerm) && reply.VoteGranted {
				votos += 1
			}
		}
	}
	// votos >= (N + 1)/2
	if votos*2 > len(nr.nodos)+1 {
		// Gana la eleccion y se convierte en lider
		fmt.Println("HE GANADO CON ", votos, " VOTOS")
		go nr.becomeLeader(storedTerm)
	} else {
		//Si no se ganan las elecciones
		fmt.Println("SE REINCIA LA RUTINA DE GESTION LIDER")
		nr.VotedFor = -1
		nr.StateNode = F
		nr.CurrentTerm = nr.CurrentTerm - 1
		go nr.gestionDeLider()
	}
}

// Function to change a node's state to follower
func (nr *NodoRaft) becomeFollower(term int) {
	nr.StateNode = F
	nr.CurrentTerm = term
	nr.VotedFor = -1 // Valor que no pueda adquirir ningin nodo
	nr.electionResetEvent = time.Now()

	go nr.gestionDeLider()
}

// Funcion que inicializa un lider
func (nr *NodoRaft) becomeLeader(term int) {
	fmt.Println("SOY EL LIDER DEL MANDATO ", term)
	nr.StateNode = L
	//Se inicializan NextIndex y MatchIndex
	for i := range nr.nodos {
		nr.NextIndex[i] = len(nr.log)
		nr.MatchIndex[i] = -1
	}
	//go nr.checkLogs()
	tick := time.NewTicker(50 * time.Millisecond)
	defer tick.Stop()
	for nr.StateNode == L {
		<-tick.C
		nr.sendHeartBeat()
	}
}

//PUEDE SER MALA IDEA USARLO PORQUE SI EL LOG ES MUY LARGO SE PASA ENTERO
// checkLogs, tras un nodo hacerse lider, se hace una llamada appendEntries a
// cada seguidor pasandole todos los logs del lider, para el el propio seguidor
// compruebe donde difieren y copie a partir de ahi
func (nr *NodoRaft) checkLogs() {
	for i, nodo := range nr.nodos {
		args := nr.makeAppendEntriesArgs(i)
		var reply RespuestaAppendEntries
		err := rpctimeout.CallTimeout(nodo, "NodoRaft.AppendEntries", args,
			&reply, 50*time.Millisecond)
		if err != nil {
			fmt.Println("Cant reach node ", i)
		}
	}
	fmt.Println("SE HA PASADO CHECK LOGS CON EXITO")
}

//Funcion encargada de enviar latidos a todos los nodos
func (nr *NodoRaft) sendHeartBeat() {
	for i, nodo := range nr.nodos {
		args := nr.makeAppendEntriesArgs(i)
		var reply RespuestaAppendEntries
		err := rpctimeout.CallTimeout(nodo, "NodoRaft.AppendEntries", args,
			&reply, 50*time.Millisecond)
		if err != nil {
			if reply.Success {
				//Actualizar valores de todo
				nr.NextIndex[i] = nr.NextIndex[i] + len(args.Entries)
				nr.MatchIndex[i] = nr.NextIndex[i] - 1
			} else {
				if nr.NextIndex[i] > 0 {
					nr.NextIndex[i] = nr.NextIndex[i] - 1
				}
			}

		} else {
			fmt.Println("Cant reach node ", i)
		}
	}
}

// Esta funcion es necesaria porque en el caso de que el logger este vacio,
// se daria un error al intentar obtener el LastLogTerm.
// Se hace la funcion para lo alargar el codigo
func (nr *NodoRaft) getLastLogData() (int, int) {
	if len(nr.log) > 0 {
		return len(nr.log) - 1, nr.log[len(nr.log)-1].Term
	} else {
		return -1, -1
	}
}

func (nr *NodoRaft) makeAppendEntriesArgs(i int) ArgsAppendEntries {
	nr.mux.Lock()
	var prevLogIndex, prevLogTerm int
	nrIndex := nr.NextIndex[i]
	if len(nr.log) > 0 {
		prevLogIndex = nrIndex - 1
		prevLogTerm = nr.log[prevLogIndex].Term
	} else {
		prevLogTerm = -1
		prevLogIndex = -1
	}
	entries := nr.log[nrIndex:]
	args := ArgsAppendEntries{
		Term:         nr.CurrentTerm,
		LeaderId:     nr.yo,
		PrevLogIndex: prevLogIndex,
		PrevLogTerm:  prevLogTerm,
		Entries:      entries,
		LeaderCommit: nr.CommitIndex,
	}
	nr.mux.Unlock()
	return args
}
