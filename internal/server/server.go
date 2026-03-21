package server

import (
	"fmt"
	"net"
	"strconv"

	"golang.org/x/text/message"
)

func handleConnection(conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, 1024)
	n, err := conn.Read(buf)

	if err != nil {
		panic(err)
	}

	msg := string(buf[:n])
	p := message.NewPrinter(message.MatchLanguage("en"))
	p.Printf("Received message: %s\n", msg)

}

func TCPServer(port, maxConnections int) {
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	currentConnections := 0

	for range maxConnections {
		conn, err := listener.Accept()
		currentConnections++
		fmt.Printf("New connection accepted. Current connections: %d\n", currentConnections)
		if err != nil {
			defer conn.Close()
			panic(err)
		}
		go handleConnection(conn)
	}
}