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

func (r *Registry) List() []string {
    r.mu.RLock()
    defer r.mu.RUnlock()

    users := make([]string, 0, len(r.client))

    for username := range r.client {
        users = append(users, username)
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
