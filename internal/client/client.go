package client

import (
	"bufio"
	"net"

	//"strings"

	"github.com/MohOdejimi/TCPChat/internal/models"
	"github.com/google/uuid"
)


type Client struct {
	ID       string      
	Conn     net.Conn    
	Username string       
	Send     chan []byte  
}

func NewClient(conn net.Conn, username string) *Client {
	return &Client{
		ID:       uuid.New().String(),
		Conn:     conn,
		Username: username,
		Send:     make(chan []byte, 256),
	}
}


func (c *Client) Read(broadcast chan models.Message, deregister chan string, done chan struct{}) {
	defer c.Conn.Close()
	defer close(done)

	scanner := bufio.NewScanner(c.Conn)
	for scanner.Scan() {
		message := scanner.Text()
		broadcast <- models.Message{
			Sender:  c.Username,
			Message: message,
		}
	}
	deregister <- c.Username
}

