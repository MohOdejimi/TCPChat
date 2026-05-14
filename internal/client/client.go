package client

import (
	"bufio"
	"net"
	"strings"
	"time"

	"github.com/MohOdejimi/TCPChat/internal/commands"
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


func (c *Client) Read(broadcast chan models.Message, deregister, list chan string, dm chan models.DMMessage, newName chan models.Rename, done chan struct{}) {
	defer c.Conn.Close()
	defer close(done)
	defer close(c.Send)

	scanner := bufio.NewScanner(c.Conn)
	for scanner.Scan() {
		message := scanner.Text()

		if strings.HasPrefix(message, "/"){
			if cmd, valid := commands.Parse(message); valid {
				switch cmd.Type {

				case commands.Quit:
					c.Conn.Write([]byte("Goodbye " + c.Username + "!\n"))
					deregister <- c.Username
					return

				case commands.List:
					list <- c.Username

				case commands.DM:
					dm <- models.DMMessage{
						Sender:   c.Username,
						Receiver: cmd.Target,
						Message:  cmd.Body,
						Time:     time.Now(),
					}
				case commands.Rename: 	
					newName <- models.Rename{
						Sender: c.Username,
						Newname: cmd.Target,
					}
			}

			} else {
				c.Send <- []byte("Invalid command. Please try again with any of the accepted commands. /list, /quit, /rename <username>, /dm <username> <message>\n")
			}
		} else {
			broadcast <- models.Message{
			Sender:  c.Username,
			Message: message,
			Time: time.Now(),
		  }
		}
	}
	
	deregister <- c.Username
}


func (c *Client) Write() {
	for msg := range c.Send {
		_, err := c.Conn.Write(msg)
		if err != nil {
			return 
		}
	}
}

