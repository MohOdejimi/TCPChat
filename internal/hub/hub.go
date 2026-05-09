package hub

import (
	"fmt"

	"github.com/MohOdejimi/TCPChat/internal/models"
)

type Hub struct {
	Broadcast  chan models.Message
	Deregister chan string
	Register   chan string
	Registry   *Registry
}

func NewHub(reg *Registry) *Hub {
	return &Hub{
		Broadcast:  make(chan models.Message),
		Deregister: make(chan string),
		Register:   make(chan string),
		Registry:   reg,
	}
}

func (h *Hub) Run() {
    for {
        select {

        case username := <-h.Register:
            fmt.Println(username + " Joined the Chat")

        case username := <-h.Deregister:
            fmt.Println(username + " Left the Chat")

        case message := <-h.Broadcast:
            fmt.Println("broadcasting:", message)
        }
    }
}