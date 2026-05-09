package server

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"

	"github.com/MohOdejimi/TCPChat/internal/hub"
)

type Server struct {
	wg sync.WaitGroup;
	connMutex sync.Mutex;
	connectionCount int
}

var serverInstance = &Server{}

var Registry = hub.NewRegistry()

func handleConnection(conn net.Conn, hubInstance *hub.Hub) {
	defer conn.Close()
	defer serverInstance.wg.Done()

	serverInstance.connMutex.Lock()
	serverInstance.connectionCount++
	serverInstance.connMutex.Unlock()

	defer func() {
		serverInstance.connMutex.Lock()
		serverInstance.connectionCount--
		serverInstance.connMutex.Unlock()
	}()

	conn.Write([]byte("Welcome to the TCP Chat Server!\n"))
	conn.Write([]byte("Please Enter Your Username: "))

	reader := bufio.NewReader(conn)
	var username string
	for {
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error reading username: %v\n", err)
			return
		}

		username = strings.TrimSpace(input)
		if username == "" {
			conn.Write([]byte("Username cannot be empty. Please enter a username: "))
			continue
		}

		_, exists  := Registry.GetUserName(username)
		if !exists {
			client := Registry.SetUserName(username, conn)
			hubInstance.Register <- client.Username
			conn.Write([]byte(fmt.Sprintf("Hello, %s! You can start chatting now.\n", username)))

		done := make(chan struct{})

		go client.Read(hubInstance.Broadcast, hubInstance.Deregister, done)

			<-done
		}

		if exists {
			conn.Write([]byte("Username already taken. Please choose a different username: "))
			continue
		}

		break 
	}	
	
}

func TCPServer(port, maxConnections int, hubInstance *hub.Hub) error{
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		fmt.Printf("Error from TCP Listener: %v\n", err)
		return err
	}
	defer listener.Close()


	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Accept error: %v\n", err)
			return err
		}

		serverInstance.connMutex.Lock()
		if serverInstance.connectionCount >= maxConnections {
			fmt.Println("Max connections reached, rejecting.")
			conn.Close()
			serverInstance.connMutex.Unlock()
			continue
		}
		serverInstance.connMutex.Unlock()

		serverInstance.wg.Add(1)
		go handleConnection(conn, hubInstance)
	}
}
