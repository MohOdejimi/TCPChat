package main

import (
	"log"

	"github.com/MohOdejimi/TCPChat/internal/cli"
	"github.com/MohOdejimi/TCPChat/internal/hub"
	"github.com/MohOdejimi/TCPChat/internal/server"
)

func main() {
	port, maxConnections := cli.Flags()

	chatHub := hub.NewHub(server.Registry)
	go chatHub.Run()
	err := server.TCPServer(port, maxConnections, chatHub)
	if err != nil {
		log.Fatalln(err)
	}
	select {}

}
