# Go TCP Server 
## Problem Statement

`````
Build a TCP server that can accept and hold a maximum of N clients (where N is configurable).
These clients are assigned ranks based on first-come-first-serve, i.e whoever connects first receives the next available high rank. Ranks are from 0–N, 0 being the highest rank.

Clients can send to the server commands that the server distributes among the clients. Only a client with a lower rank can execute a command of a higher rank client. Higher rank clients cannot execute commands by lower rank clients, so these commands are rejected. The command execution can be as simple as the client printing to console that command has been executed.

If a client disconnects the server should re-adjust the ranks and promote any client that needs to be promoted not to leave any gaps in the ranks.
`````

## Clone the project

```
$ git clone https://github.com/ianmuhia/ppal
$ cd cd ppal

$ cd hello
$ go build 
```

```
$ ./demo -c 5 # max connections the server will handle can be supplied using `` -c`` flag. 0 means no limits will be set, so the server will be bound by system resources.
```
## Types

### type [Cache](/cache.go#L10)

`type Cache struct { ... }`

Cache holds the structure for
holding a connected clients.

#### func [NewCache](/cache.go#L16)

`func NewCache() *Cache`

NewCache initializes the cache.

#### func (*Cache) [Del](/cache.go#L43)

`func (c *Cache) Del(key string) string`

#### func (*Cache) [Get](/cache.go#L24)

`func (c *Cache) Get(key string) (*Client, bool)`

Get retrieves the client from cache.

#### func (*Cache) [Set](/cache.go#L36)

`func (c *Cache) Set(val *Client, key string)`

Set Adds the client to cache.

### type [Client](/server.go#L30)

`type Client struct { ... }`

Client is a single connected client.

### type [Server](/server.go#L21)

`type Server struct { ... }`

#### func [NewServer](/server.go#L38)

`func NewServer(listenAddr string, maxConns int) (*Server, error)`

NewServer creates a new server instance with the provided parameters.
It also initialises the cache.

#### func (*Server) [HandleCommands](/server.go#L120)

`func (t *Server) HandleCommands(conn net.Conn)`

handleCommands listens for client commands and matches then to thier
respective ranks.If the client disconnects the remover fuctions closes the
connection and adjusts the remaing client ranks.

#### func (*Server) [HandleConn](/server.go#L97)

`func (t *Server) HandleConn(conn net.Conn)`

#### func (*Server) [Remover](/server.go#L156)

`func (t *Server) Remover(conn net.Conn, v *Client)`

Remover deletes a client from the cache and re-adjusts the remaing client ranks
to fill in the gaps left.

