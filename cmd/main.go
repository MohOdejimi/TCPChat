package main

import (


	"github.com/MohOdejimi/TCPChat/internal/cli"
	"github.com/MohOdejimi/TCPChat/internal/server"
)

func main() {
	port, maxConnections := cli.Flags()
	server.TCPServer(port, maxConnections)
}	
