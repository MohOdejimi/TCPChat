package cli 

import (
	"flag"
)

func Flags() (int, int) {
	port := flag.Int("port", 8080, "Port to Listen On")
	maxConnections := flag.Int("max-connections", 100, "Maximum Number of Concurrent Connections")

	flag.Parse()

	if *port <= 0 || *port > 65535 {
		panic("Invalid port number. Must be between 1 and 65535.")
	}
	if *maxConnections <= 0 {
		panic("Invalid max connections. Must be a positive integer.")
	}
	return *port, *maxConnections
}