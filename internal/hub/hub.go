package hub

import (
	"fmt"
	"time"
	"strings"

	"github.com/MohOdejimi/TCPChat/internal/models"
)

type Hub struct {
	Broadcast  chan models.Message
	Deregister chan string
	Register   chan string
	List 	   chan string
	DM 		   chan models.DMMessage
	Registry   *Registry
}

func NewHub(reg *Registry) *Hub {
	return &Hub{
		Broadcast:  make(chan models.Message),
		Deregister: make(chan string),
		Register:   make(chan string),
		List: 		make(chan string),
		DM: 		make(chan models.DMMessage),
		Registry:   reg,
	}
}



func (h *Hub) Run() { 
    for {
        select {
        case username := <-h.Register:
            currentTime := time.Now().Format("15:04:05")
            fmt.Printf("%s", fmt.Sprintf("[%s] %s has joined the chat\n", currentTime, username))
			connectedClients := h.Registry.ListOfConnectedClients()

			for _, client := range connectedClients {
				if client.Username != username {
					select {
					case client.Send <- ([]byte(fmt.Sprintf("[%s] %s has joined the chat\n", currentTime, username))):
					default:
						fmt.Printf("Failed to send message to %s\n", client.Username)
					}
				}
			}

        case username := <-h.Deregister:
			currentTime := time.Now().Format("15:04:05")
            fmt.Printf("%s", fmt.Sprintf("[%s] %s has left the chat\n", currentTime, username))
			connectedClients := h.Registry.ListOfConnectedClients()

			for _, client := range connectedClients {
				if client.Username == username {
					delete(h.Registry.client, username)
				} else{
					select {
					case client.Send <- []byte(fmt.Sprintf("[%s] %s has left the chat\n", currentTime, username)):
					default:
						fmt.Printf("Failed to send message to %s\n", client.Username)
					}
				}
			}

		case username := <-h.List: 
			connectedClients := h.Registry.ListOfConnectedClients()
			var userList []string	

			for _, client := range connectedClients {
				if client.Username != username {
					userList = append(userList, client.Username)
				}
			}
			if len(userList) == 0 {
				h.Registry.client[username].Send <- []byte("No other users are currently connected\n")
			} else {
				message := fmt.Sprintf(
					"Connected users:\n- %s\n",
					strings.Join(userList, "\n- "),
				)
				h.Registry.client[username].Send <- []byte(message)
			}
        case message := <-h.Broadcast:
			connectedClients := h.Registry.ListOfConnectedClients()
			currentTime := time.Now().Format("15:04:05")
			fmt.Printf("%s", fmt.Sprintf("[%s] Incoming message from %s: %s\n", currentTime, message.Sender, message.Message))
			messageText := fmt.Sprintf("[%s] %s: %s\n", currentTime, message.Sender, message.Message)

			for _, client := range connectedClients {
				if client.Username != message.Sender {
					select {
					case client.Send <- []byte(messageText):
					default:
						fmt.Printf("Failed to send message to %s\n", client.Username)
					}
				}
        	}
		case dmMessage := <-h.DM:
			connectedClients := h.Registry.ListOfConnectedClients()
			currentTime := time.Now().Format("15:04:05")
			receiver := dmMessage.Receiver 

			if h.Registry.IsTargetUserOnline(connectedClients, receiver) {
				for _, client := range connectedClients {
					if client.Username == receiver {
						messageText := fmt.Sprintf("[%s] DM from %s: %s\n", currentTime, dmMessage.Sender, dmMessage.Message)
						client.Send <- []byte(messageText)
					}
				}
			} else {
				if senderClient, exists := h.Registry.client[dmMessage.Sender]; exists {
					senderClient.Send <- []byte(fmt.Sprintf("%s is either not online or does not exist. DM could not be delivered.\n", receiver))
				}	
			}		
    	}
	}
}