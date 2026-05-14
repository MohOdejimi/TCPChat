# TCPChat

A real-time TCP chat server built in Go. TCPChat allows multiple clients to connect over raw TCP, register a username, and communicate in real time through broadcast messaging, private direct messages, and server-side commands — no frameworks, no WebSockets, no message brokers.

---

## Features

- **Real-time broadcast messaging** — messages sent by any client are instantly forwarded to all other connected clients
- **Private direct messaging** — clients can send messages to a specific user using `/dm`, visible only to the recipient
- **Username registration** — every client registers a unique username on connection, duplicate usernames are rejected
- **Command support** — four built-in commands for interacting with the server:
  - `/list` — returns a list of all currently connected usernames
  - `/dm <username> <message>` — sends a private message to a specific user
  - `/rename <newname>` — updates your username mid-session and notifies all connected clients
  - `/quit` — disconnects cleanly from the server with a goodbye notification to remaining clients
- **Join and leave notifications** — all connected clients are notified when someone joins or leaves the chat
- **Maximum connection limit** — the server rejects new connections beyond a configurable limit with an informative message
- **Graceful shutdown** — the server catches interrupt signals, notifies all connected clients, and exits cleanly

---

## Architecture & Design

![TCPChat Architecture](assets/TCPChat.png)

TCPChat is organised into four internal packages, each owning a single responsibility:

```
.
├── cmd/
│   └── main.go                  # Entry point — flag parsing, hub creation, server startup
├── internal/
│   ├── server/
│   │   └── server.go            # TCP listener, connection handling, goroutine spawning
│   ├── hub/
│   │   ├── hub.go               # Central hub — Run and Shutdown goroutines, channel routing
│   │   └── registry.go          # In-memory client registry protected by sync.RWMutex
│   ├── client/
│   │   └── client.go            # Client struct, Read goroutine, Write goroutine
│   ├── commands/
│   │   └── commands.go          # Command parsing — /dm, /list, /rename, /quit
│   └── models/
│       └── models.go            # Shared types — Message, DMMessage, Rename
```

### How the flows work

**Client Connection and Registration**

When a client connects, a TCP three-way handshake is completed and the server spawns a dedicated `handleConnection` goroutine. The client is prompted for a username, which is validated and checked against the registry for uniqueness. On successful registration, the client is stored in the registry and two goroutines are spawned — `Read` and `Write` — which run for the entire lifetime of the connection. All existing clients are notified of the new arrival.

```
Client → TCP Listener → handleConnection goroutine
       ← username prompt
Client → username → registry validation → client registered
       ← go client.Read + go client.Write spawned
       ← join notification broadcast to all existing clients
```

**Broadcast Message**

When a client sends a plain message, the `Read` goroutine forwards it to the hub's broadcast channel. The hub's `Run` goroutine receives it, queries the registry for all connected clients, and deposits the formatted message into each client's `Send` channel — skipping the sender. Each client's `Write` goroutine picks it up and writes it to their TCP connection.

```
Client → Read goroutine → broadcast channel → Hub.Run
                                            → registry.List (skip sender)
                                            → each client.Send channel
                                            → Write goroutine → terminal
```

**Private Message (/dm)**

When a client sends a `/dm` command, the `Read` goroutine parses it and sends a `DMMessage` struct to the hub's DM channel. The hub looks up the target username in the registry — if online, it deposits the message into only that client's `Send` channel. No other client sees it. If the target is offline, an error is sent back to the sender only.

```
Client → Read goroutine → DM channel → Hub.Run
                                     → registry lookup (target only)
                                     → recipient.Send channel
                                     → Write goroutine → recipient terminal
```

**Client Disconnection**

When a client disconnects — gracefully via `/quit` or abruptly by closing their terminal — the `Read` goroutine's scanner detects EOF and exits its loop. The client's username is sent to the hub's deregister channel. The hub removes the client from the registry and broadcasts a leave notification to all remaining clients. The `done` channel is closed, unblocking `handleConnection` and triggering the deferred `conn.Close`.

```
Client disconnects → scanner.Scan returns false (EOF)
                   → deregister channel ← username
                   → Hub.Run → registry.Deregister
                             → leave notification to remaining clients
                   → close(done) → close(client.Send)
                   → handleConnection unblocks → conn.Close
```

### Concurrency

Each connected client runs two goroutines — `Read` and `Write` — concurrently on the same TCP connection. `Read` is permanently blocked waiting for input from the client. `Write` is permanently blocked waiting on the `Send` channel for outgoing messages. Because they run in separate goroutines, neither blocks the other — a client can receive a broadcast while simultaneously typing a message.

The hub runs a single `Run` goroutine that processes all events through a `select` loop across six channels — `broadcast`, `DM`, `register`, `deregister`, `list`, and `rename`. Every concurrent event in the system flows through one place, eliminating the need for shared state across goroutines.

The client registry uses a `sync.RWMutex` — allowing multiple goroutines to read concurrently while writes remain exclusive. In a chat server where reads happen far more frequently than writes, this keeps operations non-blocking under high concurrency.

---

## Installation & Usage

### Prerequisites

- Go 1.21 or higher

### Clone and build

```bash
git clone https://github.com/MohOdejimi/TCPChat.git
cd TCPChat
go build -o tcpchat ./cmd
```

### Start the server

```bash
./tcpchat --port 8080 --max-connections 10
```

The server will start on port `8080` and accept up to `10` concurrent connections.

### Connect as a client

```bash
nc localhost 8080
```

You will see:

```
Welcome to the TCP Chat Server!
Please Enter Your Username:
```

Enter a username and start chatting.

### Send a broadcast message

```
hello everyone
```

All other connected clients will see:

```
[14:23:45] Mohammed: hello everyone
```

### Send a private message

```
/dm John hey, are you there?
```

Only John will see:

```
[14:23:50] DM from Mohammed: hey, are you there?
```

### List connected users

```
/list
```

You will see:

```
Connected users:
- John
- Sarah
```

### Rename yourself

```
/rename Mo
```

All clients will see:

```
Mohammed is now known as Mo
```

### Disconnect

```
/quit
```

All remaining clients will see:

```
[14:25:00] Mo has left the chat
```

### CLI flags

| Flag | Description | Default |
|------|-------------|---------|
| `--port` | Port for the server to listen on | `8080` |
| `--max-connections` | Maximum number of concurrent client connections | `10` |