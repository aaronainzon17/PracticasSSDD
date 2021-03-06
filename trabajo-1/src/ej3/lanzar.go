package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"

	"golang.org/x/crypto/ssh"
)

// La funcion check error se ha copaido de los ficheros proporcionados
// por el profesor para falicitar la legibilidad del codigo
func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}

func runCmd(cmd string, client string, s *ssh.ClientConfig) (string, error) {
	// open connection
	conn, err := ssh.Dial("tcp", client+":22", s)
	checkError(err)
	defer conn.Close()

	// open session
	session, err := conn.NewSession()
	checkError(err)
	defer session.Close()

	// run command and capture stdout/stderr
	output, err := session.CombinedOutput(cmd)

	return fmt.Sprintf("%s", output), err
}

func main() {

	if (len(os.Args) != 5) && (len(os.Args) != 7) {
		fmt.Println("WRONG USAGE")
		fmt.Println("Usage: go run lanzar.go <client/server> <hostUser> <remoteUser> <server>")
		os.Exit(1)
	}

	opt := os.Args[1]
	hostUser := os.Args[2]
	remoteUser := os.Args[3]
	server := os.Args[4]

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
	var result string
	fmt.Println(opt)
	if opt == "cliente" {
		fmt.Println("Entra cliente")
		ini := os.Args[5]
		fin := os.Args[6]
		result, err = runCmd("cd /home/a779088/cuarto/PracticasSSDD/trabajo-1/src/ej2/ && /usr/local/go/bin/go run cliente.go "+ini+" "+fin, server, config)
	} else if opt == "server" {
		fmt.Println("Entra server")
		result, err = runCmd("cd /home/a779088/cuarto/PracticasSSDD/trabajo-1/src/ej2/ && /usr/local/go/bin/go run server.go &", server, config)
	}
	checkError(err)
	log.Println(result)
}
