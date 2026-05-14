package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/MohOdejimi/TCPChat/internal/cli"
	"github.com/MohOdejimi/TCPChat/internal/hub"
	"github.com/MohOdejimi/TCPChat/internal/server"
)

func main() {
	port, maxConnections := cli.Flags()

	chatHub := hub.NewHub(server.Registry)
	go chatHub.Run()
	
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-signalChan
		log.Println("Shutdown signal received. Notifying clients...")
		chatHub.Shutdown()
		log.Println("Server shut down cleanly.")
		os.Exit(0)
	}()

	err := server.TCPServer(port, maxConnections, chatHub)
	if err != nil {
		log.Fatalln(err)
	}	
}
