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

type rank int

const (
	Low rank = iota
	Medium
	High
	Print = "print"
	Hello = "hello"
)

type commands struct {
	Rank int    `json:"rank"`
	Cmd  string `json:"cmd"`
}

type Server struct {
	listenAddr string
	remoteAddr string
	maxConns   int
	users      atomic.Int32
	cmds       []commands
	ln         net.Listener
	cache      *Cache
}

type ConnectedUsers struct {
	rank int
	id   string
	conn net.Conn
}

func New(listenAddr, remoteAddr string, maxConns int) (*Server, error) {
	tcpc := &Server{
		listenAddr: listenAddr,
		remoteAddr: remoteAddr,
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

		go t.handleConn(conn)
		go t.handleCommands(conn)

	}
}

func (t *Server) handleConn(conn net.Conn) {
	t.cache.Set(&ConnectedUsers{rank: int(t.users.Load()), id: conn.RemoteAddr().String(), conn: conn}, conn.RemoteAddr().String())
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
	_, err := conn.Write(data)
	if err != nil {
		log.Print(err)
	}
}

func (t *Server) handleCommands(conn net.Conn) {
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
			v, ok := t.cache.Get(key)
			if !ok {
				log.Println("error getting client from cache")
			}
			t.remover(conn, v)
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

		if contains {
			for _, v := range t.cache.data {

				t.writeConn(v.conn, []byte(fmt.Sprintf("\t%v\n\t", cmd)))

			}
		} else if cmd == "q" {
			t.remover(conn, v)
		} else {
			t.writeConn(conn, []byte(fmt.Sprintf("\tunknown command: [%s] or rank is not allowed to execute\t\n", v.id)))

		}

	}
}

func (t *Server) remover(conn net.Conn, v *ConnectedUsers) {
	_ = t.cache.Del(conn.RemoteAddr().String())
	for _, j := range t.cache.data {
		if j.rank > v.rank {
			log.Print(strings.Repeat("*", 100))
			log.Printf("before key: [%v]  value: [%v] ", j.id, j.rank)
			t.cache.Set(&ConnectedUsers{rank: j.rank - 1, id: j.id, conn: j.conn}, j.id)
			u, _ := t.cache.Get(j.id)
			log.Printf("after key: [%v]  value: [%v] ", u.id, u.rank)
			log.Print(strings.Repeat("*", 100))
			t.writeConn(j.conn, []byte(fmt.Sprintf("\tconn: [%s] disconnected, your new rank is %v\t\n ", v.id, u.rank)))
		}
	}
	conn.Close()
}
