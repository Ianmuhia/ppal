package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strings"
	"sync/atomic"

	"golang.org/x/net/netutil"
)

// commands list of commands the clients
// can send to the server.
type commands struct {
	Rank int    `json:"rank"`
	Cmd  string `json:"cmd"`
}

type Server struct {
	maxConns int
	users    atomic.Int32
	cmds     []commands
	ln       net.Listener
	cache    *Cache
}

// Client is a single connected client.
type Client struct {
	rank int
	id   string
	conn net.Conn
}

// NewServer creates a new server instance with the provided parameters.
// It also initialises the cache.
func NewServer(listenAddr string, maxConns int) (*Server, error) {
	tcpc := &Server{
		cmds: []commands{
			{
				Rank: 0,
				Cmd:  "print",
			}, {
				Rank: 1,
				Cmd:  "hello",
			}, {
				Rank: 2,
				Cmd:  "hi",
			}, {
				Rank: 3,
				Cmd:  "hi",
			}},
		maxConns: maxConns,
		cache:    NewCache(),
	}

	ln, err := net.Listen("tcp", listenAddr)
	log.Print(maxConns)
	listener := netutil.LimitListener(ln, maxConns)

	if err != nil {
		return nil, err
	}
	log.Printf("tcp server running %s", listener.Addr().String())
	tcpc.ln = listener

	return tcpc, nil
}

// acceptLoop listens for connections runs 2 groutines.
// one to handle the commands and the other to write data to the cache.
func (t *Server) acceptLoop() {
	defer func() {
		t.ln.Close()
	}()

	for {
		conn, err := t.ln.Accept()
		if err != nil {
			log.Println("accept error:", err)
			return
		}

		log.Printf("client connected %s", conn.RemoteAddr())
		if len(t.cache.data) == 0 {
			log.Printf("number of client connected %v", t.users.Add(0))
		} else {
			log.Printf("number of client connected %v", t.users.Add(1))
		}

		go t.HandleConn(conn)
		go t.HandleCommands(conn)
	}
}

func (t *Server) HandleConn(conn net.Conn) {
	t.cache.Set(&Client{rank: int(t.users.Load()), id: conn.RemoteAddr().String(), conn: conn}, conn.RemoteAddr().String())
	k, err := json.MarshalIndent(t.cmds, "", "    ")
	if err != nil {
		log.Print(err)
	}
	d, ok := t.cache.Get(conn.RemoteAddr().String())
	if !ok {
		log.Println("error getting client from cache")
	}
	u := fmt.Sprintf("connection sucessful. your rank is %v\n Avilable commands and ranks are You can only execute command below you rank: \n\t\t\t%v\n", d.rank, string(k))
	t.writeConn(conn, []byte(u))
}

func (t *Server) writeConn(conn net.Conn, data []byte) {
	if _, err := conn.Write(data); err != nil {
		log.Print(err)
	}
}

// handleCommands listens for client commands and matches then to thier
// respective ranks.If the client disconnects the remover fuctions closes the
// connection and adjusts the remaing client ranks.
func (t *Server) HandleCommands(conn net.Conn) {
	for {
		key := conn.RemoteAddr().String()
		v, ok := t.cache.Get(key)
		if !ok {
			log.Println("error getting client from cache")
		}
		// incoming request
		buffer := make([]byte, 1024)
		g, err := conn.Read(buffer)
		if err != nil {
			t.Remover(conn, v)
			return
		}

		cmd := strings.TrimSpace(string(buffer[:g]))
		contains := func(item string, i int) bool {
			for _, k := range t.cmds {
				if k.Cmd == item && i > k.Rank {
					return true
				}
			}
			return false
		}(cmd, v.rank)
		// check if the command and ranks are a match
		if !contains {
			t.writeConn(conn, []byte(fmt.Sprintf("\tunknown command: [%s] or rank is not allowed to execute\t\n", v.id)))
		}
		for _, v := range t.cache.data {
			t.writeConn(v.conn, []byte(fmt.Sprintf("\t%v\n\t", cmd)))
		}
	}
}

// Remover deletes a client from the cache and re-adjusts the remaing client ranks
// to fill in the gaps left.
func (t *Server) Remover(conn net.Conn, v *Client) {
	_ = t.cache.Del(conn.RemoteAddr().String())
	for _, j := range t.cache.data {
		if j.rank > v.rank {
			log.Print(strings.Repeat("*", 100))
			log.Printf("before key: [%v]  value: [%v] ", j.id, j.rank)
			t.cache.Set(&Client{rank: j.rank - 1, id: j.id, conn: j.conn}, j.id)
			u, _ := t.cache.Get(j.id)
			log.Printf("after key: [%v]  value: [%v] ", u.id, u.rank)
			log.Print(strings.Repeat("*", 100))
			t.writeConn(j.conn, []byte(fmt.Sprintf("\tconn: [%s] disconnected, your new rank is %v\t\n ", v.id, u.rank)))
		}
	}
	conn.Close()
}
