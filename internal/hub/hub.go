package hub

import (
	"fmt"
	"time"
	"strings"
	"net"

	"github.com/MohOdejimi/TCPChat/internal/models"
)

type Hub struct {
	Broadcast  chan models.Message
	Deregister chan string
	Register   chan string
	List 	   chan string
	DM 		   chan models.DMMessage
	Name 	   chan models.Rename
	Registry   *Registry
	Listener   net.Listener 
}

func NewHub(reg *Registry) *Hub {
	return &Hub{
		Broadcast:  make(chan models.Message),
		Deregister: make(chan string),
		Register:   make(chan string),
		List: 		make(chan string),
		DM: 		make(chan models.DMMessage),
		Name: 		make(chan models.Rename),
		Registry:   reg,
		Listener: 	nil, 
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
					h.Registry.Deregister(username)
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
			senderClient, exists := h.Registry.Get(username)
			if len(userList) == 0 {
				if exists {
					senderClient.Send <- []byte("No other users are currently connected\n")
				}
			} else {
				message := fmt.Sprintf(
					"Connected users:\n- %s\n",
					strings.Join(userList, "\n- "),
				)
				 senderClient.Send <- []byte(message)
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
				if senderClient, exists := h.Registry.Get(dmMessage.Sender); exists {
					senderClient.Send <- []byte(fmt.Sprintf("%s is either not online or does not exist. DM could not be delivered.\n", receiver))
				}	
			}	
		case newNameStruct := <-h.Name: 
			newName := newNameStruct.Newname
			oldName :=  newNameStruct.Sender

			connectedClients := h.Registry.ListOfConnectedClients()

			senderClient, exist := h.Registry.Get(oldName)

			if newName == oldName {
				if exist {
					senderClient.Send <- []byte("You are already known as " + newName + "\n")
					break
				}
			}

			if strings.TrimSpace(newName) == "" {
				if exist {
					senderClient.Conn.Write([]byte("Username cannot contain spaces. Please enter a different username: "))
				}
				continue 
			}

			if h.Registry.IsTargetUserOnline(connectedClients, newName) {
				if exist {
					senderClient.Send <- []byte(fmt.Sprintf("%s is already taken", newName))
				} 
			} else {
				h.Registry.UpdateUsername(oldName, newName, senderClient)
				connectedClients = h.Registry.ListOfConnectedClients()
				postedToServer := false

				for _, client := range connectedClients {
					if client.Username != newName {
						if !postedToServer {
							fmt.Printf("%s is now known as %s\n", oldName, newName)
							postedToServer = true
						}
						messageText := fmt.Sprintf("%s is now known as %s\n", oldName, newName)
						client.Send <- []byte(messageText)
					} else if client.Username == newName {
						messageText := fmt.Sprintf("You are now known as %s\n", newName)
						client.Send <- []byte(messageText)
					}
				}
			}
    	}
	}
}

func (h *Hub) Shutdown() {
	connectedClients := h.Registry.ListOfConnectedClients()

	for _, client := range connectedClients {
		client.Conn.Write([]byte("Server is shutting down. Goodbye.\n"))
		client.Conn.Close()
	}

	if h.Listener != nil {
		h.Listener.Close()
	}
}