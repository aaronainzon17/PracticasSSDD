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

	if len(os.Args) != 3 {
		fmt.Println("WRONG USAGE")
		fmt.Println("Usage: go run lanzar.go <username> <client>")
		os.Exit(1)
	}

	username := os.Args[1]
	client := os.Args[2]

	pemBytes, err := ioutil.ReadFile("/home/" + "aaron" + "/.ssh/id_rsa")
	checkError(err)
	signer, err := ssh.ParsePrivateKey(pemBytes)
	checkError(err)

	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			// use OpenSSH's known_hosts file if you care about host validation
			return nil
		},
	}

	result, err := runCmd("ls", client, config)
	checkError(err)

	log.Println(result)
}
