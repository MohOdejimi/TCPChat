package hub

import (
    "sync"
    "net"

    "github.com/MohOdejimi/TCPChat/internal/client"
)

type Registry struct  {
    mu sync.RWMutex
    client map[string]*client.Client
}

func NewRegistry() *Registry {
    return &Registry{
        client: make(map[string]*client.Client),
    }
}

func (r *Registry) GetUserName(username string) (string, bool) {
    r.mu.RLock()
    defer r.mu.RUnlock()

    _, exists := r.client[username]

    if !exists {
        return "", false
    }

    return username, true
}

func createNewClient(conn net.Conn, username string) *client.Client {
    return client.NewClient(conn, username)
}

func (r *Registry) SetUserName(username string, conn net.Conn) *client.Client{
    r.mu.Lock()
    defer r.mu.Unlock()

    clientStruct := createNewClient(conn, username)

    r.client[username] = clientStruct

    return clientStruct
}

func (r *Registry) ListOfConnectedClients() []*client.Client {
    r.mu.RLock()
    defer r.mu.RUnlock()

    users := make([]*client.Client, 0, len(r.client))

    for _, clientInfo := range r.client {
        users = append(users, clientInfo)
    }

    return users
}

func (r *Registry) Exists(username string) bool {
    r.mu.RLock()
    defer r.mu.RUnlock()

    clientInfo := r.client[username]
    if clientInfo == nil {
        return false
    }
    
    return true
}

func (r *Registry) IsTargetUserOnline(connectedClients []*client.Client, target string) bool {
	for _, client := range connectedClients {
		if client.Username == target {
			return true
		}
	}
	return false
}

func (r *Registry) Deregister(username string) {
    r.mu.Lock()
    defer r.mu.Unlock()

    delete(r.client, username)
}

func (r *Registry) Get(username string) (*client.Client, bool) {
    r.mu.RLock()
    defer r.mu.RUnlock()

    senderClient, exists := r.client[username]

    if exists {
        return senderClient, true
    }

    return nil, false 
}

func (r *Registry) UpdateUsername(oldName, newName string, existingClient *client.Client) {
    r.mu.Lock()
    defer r.mu.Unlock()

    delete(r.client, oldName)
    existingClient.Username = newName
    r.client[newName] = existingClient
}