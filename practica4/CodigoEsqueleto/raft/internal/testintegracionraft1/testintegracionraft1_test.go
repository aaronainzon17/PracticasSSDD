package testintegracionraft1_test

import (
	"fmt"
	"net/rpc"
	"os"
	"os/exec"
	"path/filepath"
	"raft/internal/despliegue"
	"testing"
	"time"
)

// PATH de los ejecutables de modulo golang de servicio de vistas
var PATH = filepath.Join(os.Getenv("HOME"), "tmp", "P4", "raft")

// go run testcltvts/main.go 127.0.0.1:29003 127.0.0.1:29001 127.0.0.1:29000
//var REPLICACMD = "cd " + PATH + "; go run " + EXECREPLICA
var REPLICACMD = "cd ~/cuarto/practica4/CodigoEsqueleto/raft/cmd/srvraft; go run " + EXECREPLICA

const (
	//hosts
	MAQUINA_LOCAL = "127.0.0.1"
	MAQUINA1      = "127.0.0.1"
	MAQUINA2      = "127.0.0.1"
	MAQUINA3      = "127.0.0.1"

	//puertos
	PUERTOREPLICA1 = "29001"
	PUERTOREPLICA2 = "29002"
	PUERTOREPLICA3 = "29003"

	//nodos replicas
	REPLICA1 = MAQUINA1 + ":" + PUERTOREPLICA1
	REPLICA2 = MAQUINA2 + ":" + PUERTOREPLICA2
	REPLICA3 = MAQUINA3 + ":" + PUERTOREPLICA3

	// paquete main de ejecutables relativos a PATH previo
	//EXECREPLICA = "cmd/srvraft/main.go "
	EXECREPLICA = "main.go "

	// comandos completo a ejecutar en máquinas remota con ssh. Ejemplo :
	// 				cd $HOME/raft; go run cmd/srvraft/main.go 127.0.0.1:29001

	// Ubicar, en esta constante, nombre de fichero de vuestra clave privada local
	// emparejada con la clave pública en authorized_keys de máquinas remotas

	PRIVKEYFILE = "id_rsa"
)

// TEST primer rango
func TestPrimerasPruebas(t *testing.T) { // (m *testing.M) {
	// <setup code>
	// Crear canal de resultados de ejecuciones ssh en maquinas remotas
	cr := make(CanalResultados, 2000)

	// Run test sequence

	// Test1 : No debería haber ningun primario, si SV no ha recibido aún latidos
	t.Run("T1:ElegirPrimerLider",
		func(t *testing.T) { cr.soloArranqueYparadaTest1(t) })

	// Test2 : No debería haber ningun primario, si SV no ha recibido aún latidos
	t.Run("T1:ElegirPrimerLider",
		func(t *testing.T) { cr.ElegirPrimerLiderTest2(t) })

	// Test3: tenemos el primer primario correcto
	t.Run("T2:FalloAnteriorElegirNuevoLider",
		func(t *testing.T) { cr.FalloAnteriorElegirNuevoLiderTest3(t) })

	// Test4: Primer nodo copia
	t.Run("T3:EscriturasConcurrentes",
		func(t *testing.T) { cr.tresOperacionesComprometidasEstable(t) })

	// tear down code
	// eliminar procesos en máquinas remotas
	//cr.stop()
}

// ---------------------------------------------------------------------
//
// Canal de resultados de ejecución de comandos ssh remotos
type CanalResultados chan string

/*func (cr *CanalResultados) stop() {
	//close(ts.cmdOutput)

	// Leer las salidas obtenidos de los comandos ssh ejecutados
	for s := range cr {
		fmt.Println(s)
	}
}*/

// start  gestor de vistas; mapa de replicas y maquinas donde ubicarlos;
// y lista clientes (host:puerto)
func (cr *CanalResultados) startDistributedProcesses(
	replicasMaquinas map[string]string) {

	for replica, maquina := range replicasMaquinas {
		despliegue.ExecMutipleHosts(
			REPLICACMD+" "+replica+" > /dev/null 2>&1 &",
			[]string{maquina}, make(chan string), PRIVKEYFILE)

		// dar tiempo para se establezcan las replicas
		time.Sleep(1000 * time.Millisecond)
	}
}

type NrArgs struct {
	Operacion interface{}
}

type NrReply struct {
	Indice  int
	Mandato int
	EsLider bool
}

// ~/Documents/GitHub/PracticasSSDD/practica4/CodigoEsqueleto/raft/cmd/srvraft
// ~/cuarto/practica4/CodigoEsqueleto/raft/cmd/srvraft
func (cr *CanalResultados) startLocalProcesses(
	replicasMaquinas map[string]string) {

	for replica := range replicasMaquinas {
		route := "cd ~/Documents/GitHub/PracticasSSDD/practica4/CodigoEsqueleto/raft/cmd/srvraft"
		gorun := "go run main.go " + replica + " > /dev/null 2>&1 &"
		cmd := exec.Command("/bin/bash", "-c", route+";"+gorun)
		err := cmd.Run()
		if err != nil {
			panic(err.Error())
		}
	}
}

func (cr *CanalResultados) stopDistributedProcesses(
	replicasMaquinas map[string]string) {

	// Parar procesos que han sido distribuidos con ssh ??
	for replica := range replicasMaquinas {
		rpcConn, err := rpc.DialHTTP("tcp", replica)
		if err == nil {
			var reply int
			err = rpcConn.Call("OpsServer.StopNode", NrArgs{}, &reply)
			if err != nil {
				fmt.Println("Unable to exit\n", err)
				os.Exit(1)
			}
		} else {
			fmt.Println("stopDistributedProcesses conn error", err)
		}
	}
}

// --------------------------------------------------------------------------
// FUNCIONES DE SUBTESTS

// Se pone en marcha una replica ??
func (cr *CanalResultados) soloArranqueYparadaTest1(t *testing.T) {
	t.Skip("SKIPPED soloArranqueYparadaTest1")

	fmt.Println(t.Name(), ".....................")

	// Poner en marcha replicas en remoto
	cr.startLocalProcesses(map[string]string{REPLICA1: MAQUINA1})
	time.Sleep(2 * time.Second)
	// Parar réplicas alamcenamiento en remoto
	cr.stopDistributedProcesses(map[string]string{REPLICA1: MAQUINA1})

	fmt.Println(".............", t.Name(), "Superado")
}

// Primer lider en marcha
func (cr *CanalResultados) ElegirPrimerLiderTest2(t *testing.T) {
	t.Skip("SKIPPED ElegirPrimerLiderTest2")

	fmt.Println(t.Name(), ".....................")

	// Poner en marcha  3 réplicas Raft
	replicasMaquinas :=
		map[string]string{REPLICA1: MAQUINA1, REPLICA2: MAQUINA2, REPLICA3: MAQUINA3}
	cr.startLocalProcesses(replicasMaquinas)
	time.Sleep(2 * time.Second)
	// Se ha elegido lider ?
	fmt.Printf("Probando lider en curso\n")
	pruebaUnLider(replicasMaquinas)

	// Parar réplicas alamcenamiento en remoto
	cr.stopDistributedProcesses(replicasMaquinas)

	fmt.Println(".............", t.Name(), "Superado")
}

// Fallo de un primer lider y reeleccion de uno nuevo
func (cr *CanalResultados) FalloAnteriorElegirNuevoLiderTest3(t *testing.T) {
	t.Skip("SKIPPED FalloAnteriorElegirNuevoLiderTest3")

	fmt.Println(t.Name(), ".....................")

	// Poner en marcha  3 réplicas Raft
	replicasMaquinas :=
		map[string]string{REPLICA1: MAQUINA1, REPLICA2: MAQUINA2, REPLICA3: MAQUINA3}
	cr.startLocalProcesses(replicasMaquinas)
	time.Sleep(2 * time.Second)
	fmt.Printf("Lider inicial\n")
	pruebaUnLider(replicasMaquinas)

	// Desconectar lider
	desconectaLider(replicasMaquinas)
	time.Sleep(2 * time.Second)
	fmt.Printf("Comprobar nuevo lider\n")
	pruebaUnLider(replicasMaquinas)

	// Parar réplicas alamcenamiento en remoto
	cr.stopDistributedProcesses(replicasMaquinas)

	fmt.Println(".............", t.Name(), "Superado")
}

// 3 operaciones comprometidas con situacion estable y sin fallos
func (cr *CanalResultados) tresOperacionesComprometidasEstable(t *testing.T) {
	//t.Skip("SKIPPED tresOperacionesComprometidasEstable")

	fmt.Println(t.Name(), ".....................")

	// Poner en marcha  3 réplicas Raft
	replicasMaquinas :=
		map[string]string{REPLICA1: MAQUINA1, REPLICA2: MAQUINA2, REPLICA3: MAQUINA3}
	//cr.startLocalProcesses(replicasMaquinas)
	time.Sleep(2 * time.Second)
	cmds := []string{"op1", "op2", "op3"}
	for _, cmd := range cmds {
		for replica := range replicasMaquinas {
			rpcConn, err := rpc.DialHTTP("tcp", replica)
			if err == nil {
				var reply string
				err = rpcConn.Call("OpsServer.NodeState", NrArgs{}, &reply)
				if err != nil {
					fmt.Println("No se puede obtener el estado\n", err)
				}
				if reply == "leader" {
					replyOp := someterOperacionTest(NrArgs{Operacion: cmd}, rpcConn)
					if replyOp.EsLider {
						fmt.Println("a")
					} else {
						fmt.Println("b")
					}
				}

			} else {
				fmt.Println("Connexion error ComprometerOperacionesTest: ", err)
			}
		}
		time.Sleep(2000 * time.Millisecond)
	}

	// Parar réplicas alamcenamiento en remoto
	cr.stopDistributedProcesses(replicasMaquinas)

	fmt.Println(".............", t.Name(), "Superado")
}

// --------------------------------------------------------------------------
// FUNCIONES DE APOYO
// Comprobar que hay un solo lider
// probar varias veces si se necesitan reelecciones
func pruebaUnLider(replicasMaquinas map[string]string) int {
	for iters := 0; iters < 10; iters++ {
		time.Sleep(500 * time.Millisecond)
		mapaLideres := make(map[int][]int)
		i := 0
		for replica := range replicasMaquinas {
			rpcConn, err := rpc.DialHTTP("tcp", replica)
			if err == nil {
				var reply string
				err = rpcConn.Call("OpsServer.NodeState", NrArgs{}, &reply)
				if err != nil {
					fmt.Println("No se puede obtener el estado\n", err)
				}
				if reply == "leader" {
					mapaLideres[iters] = append(mapaLideres[iters], i)
				}

			} else {
				fmt.Println("pruebaUnLider conn err: ", err)
			}
			i++
		}

		ultimoMandatoConLider := -1
		for t, lideres := range mapaLideres {
			if len(lideres) > 1 {
				fmt.Println("mandato ", t, "tiene ", len(lideres), " (>1) lideres")
			}
			if t > ultimoMandatoConLider {
				ultimoMandatoConLider = t
			}
		}

		if len(mapaLideres) != 0 {

			return mapaLideres[ultimoMandatoConLider][0] // Termina

		}
	}
	fmt.Println("un lider esperado, ninguno obtenido")

	return -1 // Termina
}

func desconectaLider(replicasMaquinas map[string]string) {

	i := 0
	for replica := range replicasMaquinas {
		rpcConn, err := rpc.DialHTTP("tcp", replica)
		if err == nil {
			var reply string
			err = rpcConn.Call("OpsServer.NodeState", NrArgs{}, &reply)
			if err != nil {
				fmt.Println("No se puede obtener el estado\n", err)
			}
			if reply == "leader" {
				var reply int
				err = rpcConn.Call("OpsServer.StopNode", NrArgs{}, &reply)
				if err != nil {
					fmt.Println("Unable to exit\n", err)
					os.Exit(1)
				} else {
					fmt.Println("Leader ", replica, "stopped")
				}
			}
		} else {
			fmt.Println("desconectaLider conn err: ", err)
			os.Exit(1)
		}
		i++
	}
}

func someterOperacionTest(args NrArgs, rpcConn *rpc.Client) NrReply {
	var reply NrReply
	err := rpcConn.Call("OpsServer.Submit", args, &reply)
	if err != nil {
		fmt.Println("Couldn't submit operation\n", err)
		os.Exit(1)
	}
	return reply
}
