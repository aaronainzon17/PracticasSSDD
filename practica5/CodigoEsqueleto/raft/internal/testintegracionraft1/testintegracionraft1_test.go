//scp -r practica5 a779088@central.cps.unizar.es:~/cuarto
package testintegracionraft1

import (
	"fmt"
	"io/ioutil"
	"net"
	"os/exec"
	"raft/internal/comun/check"

	//"log"
	//"crypto/rand"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"raft/internal/comun/rpctimeout"
	"raft/internal/raft"

	"golang.org/x/crypto/ssh"
)

const (
	//hosts
	MAQUINA1 = "155.210.154.201"
	MAQUINA2 = "155.210.154.202"
	MAQUINA3 = "155.210.154.205"

	//puertos
	PUERTOREPLICA1 = "29060"
	PUERTOREPLICA2 = "29060"
	PUERTOREPLICA3 = "29060"

	//hosts
	/*MAQUINA1 = "127.0.0.1"
	MAQUINA2 = "127.0.0.1"
	MAQUINA3 = "127.0.0.1"

	//puertos
	PUERTOREPLICA1 = "29060"
	PUERTOREPLICA2 = "29061"
	PUERTOREPLICA3 = "29062"
	*/
	//nodos replicas
	REPLICA1 = MAQUINA1 + ":" + PUERTOREPLICA1
	REPLICA2 = MAQUINA2 + ":" + PUERTOREPLICA2
	REPLICA3 = MAQUINA3 + ":" + PUERTOREPLICA3

	// paquete main de ejecutables relativos a PATH previo
	EXECREPLICA = "cmd/srvraft/main.go"

	// comandos completo a ejecutar en máquinas remota con ssh. Ejemplo :
	// 				cd $HOME/raft; go run cmd/srvraft/main.go 127.0.0.1:29001

	// Ubicar, en esta constante, nombre de fichero de vuestra clave privada local
	// emparejada con la clave pública en authorized_keys de máquinas remotas

	PRIVKEYFILE = "id_rsa"
)

// PATH de los ejecutables de modulo golang de servicio Raft
var PATH string = filepath.Join(os.Getenv("HOME"), "tmp", "p5", "raft")

// go run cmd/srvraft/main.go 0 127.0.0.1:29001 127.0.0.1:29002 127.0.0.1:29003
var EXECREPLICACMD string = "cd " + PATH + "; go run " + EXECREPLICA

// TEST primer rango
func TestPrimerasPruebas(t *testing.T) { // (m *testing.M) {
	// <setup code>
	// Crear canal de resultados de ejecuciones ssh en maquinas remotas
	cfg := makeCfgDespliegue(t,
		3,
		[]string{REPLICA1, REPLICA2, REPLICA3},
		[]bool{true, true, true})

	// tear down code
	// eliminar procesos en máquinas remotas
	defer cfg.stop()

	// Run test sequence

	// Test1 : No debería haber ningun primario, si SV no ha recibido aún latidos
	t.Run("T1:soloArranqueYparada",
		func(t *testing.T) { cfg.soloArranqueYparadaTest1(t) })
	time.Sleep(2 * time.Second)
	// Test2 : No debería haber ningun primario, si SV no ha recibido aún latidos
	t.Run("T2:ElegirPrimerLider",
		func(t *testing.T) { cfg.elegirPrimerLiderTest2(t) })
	time.Sleep(2 * time.Second)
	// Test3: tenemos el primer primario correcto
	t.Run("T3:FalloAnteriorElegirNuevoLider",
		func(t *testing.T) { cfg.falloAnteriorElegirNuevoLiderTest3(t) })
	time.Sleep(2 * time.Second)
	// Test4: Tres operaciones comprometidas en configuración estable
	t.Run("T4:tresOperacionesComprometidasEstable",
		func(t *testing.T) { cfg.tresOperacionesComprometidasEstable(t) })
	time.Sleep(2 * time.Second)
}

// TEST primer rango
func TestAcuerdosConFallos(t *testing.T) { // (m *testing.M) {
	// <setup code>
	// Crear canal de resultados de ejecuciones ssh en maquinas remotas
	cfg := makeCfgDespliegue(t,
		3,
		[]string{REPLICA1, REPLICA2, REPLICA3},
		[]bool{true, true, true})

	// tear down code
	// eliminar procesos en máquinas remotas
	defer cfg.stop()
	time.Sleep(2 * time.Second)
	// Test5: Se consigue acuerdo a pesar de desconexiones de seguidor
	t.Run("T5:AcuerdoAPesarDeDesconexionesDeSeguidor ",
		func(t *testing.T) { cfg.AcuerdoApesarDeSeguidor(t) })
	time.Sleep(2 * time.Second)
	t.Run("T5:SinAcuerdoPorFallos ",
		func(t *testing.T) { cfg.SinAcuerdoPorFallos(t) })
	time.Sleep(2 * time.Second)
	t.Run("T5:SometerConcurrentementeOperaciones ",
		func(t *testing.T) { cfg.SometerConcurrentementeOperaciones(t) })
}

// ---------------------------------------------------------------------
//
// Canal de resultados de ejecución de comandos ssh remotos
type canalResultados chan string

func (cr canalResultados) stop() {
	close(cr)

	// Leer las salidas obtenidos de los comandos ssh ejecutados
	for s := range cr {
		fmt.Println(s)
	}
}

// ---------------------------------------------------------------------
// Operativa en configuracion de despliegue y pruebas asociadas
type configDespliegue struct {
	t           *testing.T
	conectados  []bool
	numReplicas int
	nodosRaft   []rpctimeout.HostPort
	cr          canalResultados
}

// Crear una configuracion de despliegue
func makeCfgDespliegue(t *testing.T, n int, nodosraft []string,
	conectados []bool) *configDespliegue {
	cfg := &configDespliegue{}
	cfg.t = t
	cfg.conectados = conectados
	cfg.numReplicas = n
	cfg.nodosRaft = rpctimeout.StringArrayToHostPortArray(nodosraft)
	cfg.cr = make(canalResultados, 2000)

	return cfg
}

func (cfg *configDespliegue) stop() {
	//cfg.stopDistributedProcesses()

	time.Sleep(50 * time.Millisecond)

	cfg.cr.stop()
}

// --------------------------------------------------------------------------
// FUNCIONES DE SUBTESTS

// Se pone en marcha una replica ?? - 3 NODOS RAFT
func (cfg *configDespliegue) soloArranqueYparadaTest1(t *testing.T) {
	//t.Skip("SKIPPED soloArranqueYparadaTest1")

	fmt.Println(t.Name(), ".....................")

	cfg.t = t // Actualizar la estructura de datos de tests para errores

	// Poner en marcha replicas en remoto con un tiempo de espera incluido
	cfg.startDistributedProcesses()
	time.Sleep(6 * time.Second)

	// Comprobar estado replica 0
	cfg.comprobarEstadoRemoto(0, 1, false, -1)

	// Comprobar estado replica 1
	cfg.comprobarEstadoRemoto(1, 1, false, -1)

	// Comprobar estado replica 2
	cfg.comprobarEstadoRemoto(2, 1, false, -1)

	// Parar réplicas almacenamiento en remoto
	cfg.stopDistributedProcesses()

	fmt.Println(".............", t.Name(), "Superado")
}

// Primer lider en marcha - 3 NODOS RAFT
func (cfg *configDespliegue) elegirPrimerLiderTest2(t *testing.T) {
	//t.Skip("SKIPPED ElegirPrimerLiderTest2")

	fmt.Println(t.Name(), ".....................")
	cfg.t = t // Actualizar la estructura de datos de tests para errores

	cfg.startDistributedProcesses()
	time.Sleep(6 * time.Second)
	// Se ha elegido lider ?
	fmt.Printf("Probando lider en curso\n")
	cfg.pruebaUnLider(3)

	// Parar réplicas alamcenamiento en remoto
	cfg.stopDistributedProcesses() // Parametros

	fmt.Println(".............", t.Name(), "Superado")
}

// Fallo de un primer lider y reeleccion de uno nuevo - 3 NODOS RAFT
func (cfg *configDespliegue) falloAnteriorElegirNuevoLiderTest3(t *testing.T) {
	//t.Skip("SKIPPED FalloAnteriorElegirNuevoLiderTest3")

	fmt.Println(t.Name(), ".....................")
	cfg.t = t // Actualizar la estructura de datos de tests para errores

	cfg.startDistributedProcesses()
	time.Sleep(6 * time.Second)
	fmt.Printf("Lider inicial\n")
	cfg.pruebaUnLider(3)

	// Desconectar lider
	i := cfg.desconectaLider()
	time.Sleep(1 * time.Second)
	//Se veulve a levantar el nodo
	cfg.startOneNodeDistributed(i)
	time.Sleep(6 * time.Second)

	fmt.Printf("Comprobar nuevo lider\n")
	cfg.pruebaUnLider(3)

	// Parar réplicas almacenamiento en remoto
	cfg.stopDistributedProcesses() //parametros

	fmt.Println(".............", t.Name(), "Superado")
}

// 3 operaciones comprometidas con situacion estable y sin fallos - 3 NODOS RAFT
func (cfg *configDespliegue) tresOperacionesComprometidasEstable(t *testing.T) {
	//t.Skip("SKIPPED tresOperacionesComprometidasEstable")

	fmt.Println(t.Name(), ".....................")
	cfg.t = t // Actualizar la estructura de datos de tests para errores

	cfg.startDistributedProcesses()
	time.Sleep(6 * time.Second)

	cmds := []raft.TipoOperacion{
		{Operacion: "leer", Clave: "1"},
		{Operacion: "leer", Clave: "2"},
		{Operacion: "escribir", Clave: "3", Valor: "mensaje1"}}
	cfg.someterOperaciones(cmds)
	time.Sleep(100 * time.Millisecond)
	// Se comprueba que el commitIndex es 2 ya que se han comprometido 3 operaciones(0,1,2)
	cfg.comrprobarComprometidas(2)
	// Parar réplicas alamcenamiento en remoto
	cfg.stopDistributedProcesses() // Parametros

	fmt.Println(".............", t.Name(), "Superado")
}

// Se consigue acuerdo a pesar de desconexiones de seguidor -- 3 NODOS RAFT
func (cfg *configDespliegue) AcuerdoApesarDeSeguidor(t *testing.T) {
	//t.Skip("SKIPPED AcuerdoApesarDeSeguidor")

	fmt.Println(t.Name(), ".....................")
	cfg.t = t // Actualizar la estructura de datos de tests para errores

	// Obtener un lider
	cfg.startDistributedProcesses()
	time.Sleep(6 * time.Second)
	// Desconectar una de los nodos Raft
	index := cfg.desconectaSeguidor()
	time.Sleep(1000 * time.Millisecond)
	cmds := []raft.TipoOperacion{
		{Operacion: "escribir", Clave: "1", Valor: "mensaje1"},
		{Operacion: "escribir", Clave: "2", Valor: "mensaje2"},
		{Operacion: "escribir", Clave: "3", Valor: "mensaje3"}}
	cfg.someterOperaciones(cmds)
	time.Sleep(100 * time.Millisecond)
	// Se comprueba que el commitIndex es 2 ya que se han comprometido 3 operaciones(0,1,2)
	cfg.comrprobarComprometidas(2)
	// reconectar nodo Raft previamente desconectado y comprobar varios acuerdos

	cfg.startOneNodeDistributed(index)
	time.Sleep(6 * time.Second)

	cmds = []raft.TipoOperacion{
		{Operacion: "escribir", Clave: "4", Valor: "mensaje4"},
		{Operacion: "leer", Clave: "3"},
		{Operacion: "escribir", Clave: "5", Valor: "mensaje5"}}
	cfg.someterOperaciones(cmds)
	time.Sleep(100 * time.Millisecond)
	cfg.comrprobarComprometidas(5)
	// Parar réplicas alamcenamiento en remoto
	cfg.stopDistributedProcesses()

	fmt.Println(".............", t.Name(), "Superado")
}

// NO se consigue acuerdo al desconectarse mayoría de seguidores -- 3 NODOS RAFT
func (cfg *configDespliegue) SinAcuerdoPorFallos(t *testing.T) {
	//t.Skip("SKIPPED SinAcuerdoPorFallos")

	fmt.Println(t.Name(), ".....................")
	cfg.t = t // Actualizar la estructura de datos de tests para errores

	// Obtener un lider
	cfg.startDistributedProcesses()
	time.Sleep(6 * time.Second)
	// Desconectar una de los nodos Raft
	nodo1 := cfg.desconectaSeguidor()
	nodo2 := cfg.desconectaSeguidor()
	time.Sleep(1000 * time.Millisecond)
	cmds := []raft.TipoOperacion{
		{Operacion: "escribir", Clave: "1", Valor: "mensaje1"},
		{Operacion: "escribir", Clave: "2", Valor: "mensaje2"},
		{Operacion: "escribir", Clave: "3", Valor: "mensaje3"}}
	cfg.someterOperaciones(cmds)
	time.Sleep(100 * time.Millisecond)
	// Se comprueba que el commitIndex es 2 ya que se han comprometido 3 operaciones(0,1,2)
	cfg.comrprobarComprometidas(2)
	// reconectar nodo Raft previamente desconectado y comprobar varios acuerdos
	cfg.startOneNodeDistributed(nodo1)
	cfg.startOneNodeDistributed(nodo2)
	time.Sleep(6 * time.Second)
	cmds = []raft.TipoOperacion{
		{Operacion: "escribir", Clave: "4", Valor: "mensaje4"},
		{Operacion: "leer", Clave: "3"},
		{Operacion: "escribir", Clave: "5", Valor: "mensaje5"}}
	cfg.someterOperaciones(cmds)
	time.Sleep(100 * time.Millisecond)
	cfg.comrprobarComprometidas(5)
	// Parar réplicas alamcenamiento en remoto
	cfg.stopDistributedProcesses()

	fmt.Println(".............", t.Name(), "Superado")

	// Comprometer una entrada

	//  Obtener un lider y, a continuación desconectar 2 de los nodos Raft

	// Comprobar varios acuerdos con 2 réplicas desconectada

	// reconectar lo2 nodos Raft  desconectados y probar varios acuerdos
}

// Se somete 5 operaciones de forma concurrente -- 3 NODOS RAFT
func (cfg *configDespliegue) SometerConcurrentementeOperaciones(t *testing.T) {
	//t.Skip("SKIPPED SometerConcurrentementeOperaciones")

	fmt.Println(t.Name(), ".....................")
	cfg.t = t // Actualizar la estructura de datos de tests para errores

	// Obtener un lider y, a continuación someter una operacion
	cfg.startDistributedProcesses()
	time.Sleep(6 * time.Second)
	cmds := []raft.TipoOperacion{
		{Operacion: "escribir", Clave: "1", Valor: "mensaje1"},
		{Operacion: "escribir", Clave: "2", Valor: "mensaje2"},
		{Operacion: "escribir", Clave: "3", Valor: "mensaje3"},
		{Operacion: "escribir", Clave: "4", Valor: "mensaje4"},
		{Operacion: "escribir", Clave: "5", Valor: "mensaje5"}}

	// Someter 5  operaciones concurrentes, un bucle para estabilizar la ejecucion
	for _, cmd := range cmds {
		go cfg.someterOperaciones([]raft.TipoOperacion{cmd})
	}
	time.Sleep(1 * time.Second)
	cfg.comrprobarComprometidas(4)
	// Parar réplicas alamcenamiento en remoto
	cfg.stopDistributedProcesses()

	fmt.Println(".............", t.Name(), "Superado")
}

// --------------------------------------------------------------------------
// FUNCIONES DE APOYO
// Comprobar que hay un solo lider
// probar varias veces si se necesitan reelecciones
func (cfg *configDespliegue) pruebaUnLider(numreplicas int) int {
	for iters := 0; iters < 10; iters++ {
		time.Sleep(500 * time.Millisecond)
		mapaLideres := make(map[int][]int)
		for i := 0; i < numreplicas; i++ {
			if cfg.conectados[i] {
				if _, mandato, eslider, _ := cfg.obtenerEstadoRemoto(i); eslider {
					mapaLideres[mandato] = append(mapaLideres[mandato], i)
				}
			}
		}

		ultimoMandatoConLider := -1
		for mandato, lideres := range mapaLideres {
			if len(lideres) > 1 {
				cfg.t.Fatalf("mandato %d tiene %d (>1) lideres",
					mandato, len(lideres))
			}
			if mandato > ultimoMandatoConLider {
				ultimoMandatoConLider = mandato
			}
		}

		if len(mapaLideres) != 0 {

			return mapaLideres[ultimoMandatoConLider][0] // Termina

		}
	}
	cfg.t.Fatalf("un lider esperado, ninguno obtenido")

	return -1 // Termina
}

func (cfg *configDespliegue) desconectaLider() int {
	lider := -1
	for i, endPoint := range cfg.nodosRaft {
		_, _, esLider, _ := cfg.obtenerEstadoRemoto(i)
		if esLider {
			lider = i
			var reply raft.Vacio
			err := endPoint.CallTimeout("NodoRaft.ParaNodo",
				raft.Vacio{}, &reply, 50*time.Millisecond)
			check.CheckError(err, "Error en llamada RPC Para nodo")
		}
	}
	return lider
}

func (cfg *configDespliegue) desconectaSeguidor() int {
	desconectado := -1
	done := false
	for i, endPoint := range cfg.nodosRaft {
		if cfg.conectados[i] {
			_, _, esLider, _ := cfg.obtenerEstadoRemoto(i)
			if !esLider && !done {
				fmt.Println("Desconecto a:", i)
				desconectado = i
				cfg.conectados[i] = false
				var reply raft.Vacio
				err := endPoint.CallTimeout("NodoRaft.ParaNodo",
					raft.Vacio{}, &reply, 50*time.Millisecond)
				check.CheckError(err, "Error en llamada RPC Para nodo")
				done = true
			}
		}
	}
	return desconectado
}

func (cfg *configDespliegue) encontrarLider() int {
	lider := -1
	for i := range cfg.nodosRaft {
		if cfg.conectados[i] {
			_, _, esLider, _ := cfg.obtenerEstadoRemoto(i)
			if esLider {
				lider = i
				fmt.Println("He encontrado al lider: ", lider)
				break
			}
		}
	}
	return lider
}

func (cfg *configDespliegue) comrprobarComprometidas(n int) bool {
	var reply bool
	for i, endPoint := range cfg.nodosRaft {
		if cfg.conectados[i] {
			reply = false
			err := endPoint.CallTimeout("NodoRaft.CheckCommits",
				n, &reply, 50*time.Millisecond)
			check.CheckError(err, "Error en llamada RPC CheckCommits nodo")
			if !reply {
				fmt.Println("El nodo ", i, " no tiene todas las entradas comprometidas")
			} else {
				fmt.Println("BIEN:", i)
			}
		}
	}
	return reply
}

func (cfg *configDespliegue) obtenerEstadoRemoto(
	indiceNodo int) (int, int, bool, int) {
	var reply raft.EstadoRemoto
	cfg.nodosRaft[indiceNodo].CallTimeout("NodoRaft.ObtenerEstadoNodo",
		raft.Vacio{}, &reply, 50*time.Millisecond)
	//check.CheckError(err, "Error en llamada RPC ObtenerEstadoRemoto")

	return reply.IdNodo, reply.Mandato, reply.EsLider, reply.IdLider
}

// start  gestor de vistas; mapa de replicas y maquinas donde ubicarlos;
// y lista clientes (host:puerto)
/*func (cfg *configDespliegue) startDistributedProcesses() {
	//cfg.t.Log("Before starting following distributed processes: ", cfg.nodosRaft)

	for i, endPoint := range cfg.nodosRaft {
		despliegue.ExecMutipleHosts(EXECREPLICACMD+
			" "+strconv.Itoa(i)+" "+
			rpctimeout.HostPortArrayToString(cfg.nodosRaft),
			[]string{endPoint.Host()}, cfg.cr, PRIVKEYFILE)

		// dar tiempo para se establezcan las replicas
		//time.Sleep(500 * time.Millisecond)
	}

	// aproximadamente 500 ms para cada arranque por ssh en portatil
	time.Sleep(2500 * time.Millisecond)
}*/

//
func (cfg *configDespliegue) stopDistributedProcesses() {
	var reply raft.Vacio

	for _, endPoint := range cfg.nodosRaft {
		err := endPoint.CallTimeout("NodoRaft.ParaNodo",
			raft.Vacio{}, &reply, 50*time.Millisecond)
		check.CheckError(err, "Error en llamada RPC PararNodo nodo")
	}
}

// Comprobar estado remoto de un nodo con respecto a un estado prefijado
func (cfg *configDespliegue) comprobarEstadoRemoto(idNodoDeseado int,
	mandatoDeseado int, esLiderDeseado bool, IdLiderDeseado int) {
	idNodo, mandato, _, idLider := cfg.obtenerEstadoRemoto(idNodoDeseado)

	//cfg.t.Log("Estado replica 0: ", idNodo, mandato, esLider, idLider, "\n")

	if idNodo != idNodoDeseado || mandato != mandatoDeseado || idLider != IdLiderDeseado {
		cfg.t.Fatalf("Estado incorrecto en replica %d en subtest %s",
			idNodoDeseado, cfg.t.Name())
	}

}

func (cfg *configDespliegue) someterOperaciones(cmds []raft.TipoOperacion) {
	lider := cfg.encontrarLider()
	for _, cmd := range cmds {
		var reply raft.ResultadoRemoto
		err := cfg.nodosRaft[lider].CallTimeout("NodoRaft.SometerOperacionRaft",
			cmd, &reply, 50*time.Millisecond)
		check.CheckError(err, "Error en llamada RPC SometerOperacion")
	}
}

func (cfg *configDespliegue) startLocalProcesses() {
	for i := range cfg.nodosRaft {
		route := "cd /home/aaron/Documents/GitHub/PracticasSSDD/practica5/CodigoEsqueleto/raft/cmd/srvraft"
		gorun := "go run main.go " + strconv.Itoa(i) + " " + rpctimeout.HostPortArrayToString(cfg.nodosRaft) + " > /dev/null 2>&1 &"
		cmd := exec.Command("/bin/bash", "-c", route+";"+gorun)
		err := cmd.Run()
		if err != nil {
			panic(err.Error())
		}
	}
	time.Sleep(2500 * time.Millisecond)
}

func (cfg *configDespliegue) startOneNode(i int) {
	route := "cd /home/aaron/Documents/GitHub/PracticasSSDD/practica5/CodigoEsqueleto/raft/cmd/srvraft"
	gorun := "go run main.go " + strconv.Itoa(i) + " " + rpctimeout.HostPortArrayToString(cfg.nodosRaft) + " > /dev/null 2>&1 &"
	fmt.Println(gorun)
	cmd := exec.Command("/bin/bash", "-c", route+";"+gorun)
	err := cmd.Run()
	if err != nil {
		panic(err.Error())
	}
	cfg.conectados[i] = true
	fmt.Println("Lanzando el nodo: ", i)
	time.Sleep(2500 * time.Millisecond)
}

/*EJECUCION EN DISTRIBUIDO*/

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}

func runCmd(cmd string, client string, s *ssh.ClientConfig) error {
	// open connection
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
	//fmt.Println(string(salida))
	return err
}

func (cfg *configDespliegue) sshUp(node rpctimeout.HostPort, i int, hostUser string, remoteUser string) {
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

	cmd := "cd /home/a779088/cuarto/practica5/CodigoEsqueleto/raft/cmd/srvraft &&" +
		" ./main " + strconv.Itoa(i) + " " + rpctimeout.HostPortArrayToString(cfg.nodosRaft) + " &" //> /dev/null 2>&1
	//fmt.Println(cmd)
	err = runCmd(cmd, node.Host(), config)
	checkError(err)
}

// start  gestor de vistas; mapa de replicas y maquinas donde ubicarlos;
// y lista clientes (host:puerto)
func (cfg *configDespliegue) startDistributedProcesses() {
	for i, endPoint := range cfg.nodosRaft {
		go cfg.sshUp(endPoint, i, "a779088", "a779088")
	}
}

// start  gestor de vistas; mapa de replicas y maquinas donde ubicarlos;
// y lista clientes (host:puerto)
func (cfg *configDespliegue) startOneNodeDistributed(i int) {
	go cfg.sshUp(cfg.nodosRaft[i], i, "a779088", "a779088")
	cfg.conectados[i] = true
}
