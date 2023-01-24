package main

import (
	"encoding/gob"
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

type Server struct {
	listenAddr   string
	remoteAddr   string
	maxConns     int
	users        atomic.Int32
	Sendchan     chan string
	Recvchan     chan string
	outboundConn net.Conn
	ln           net.Listener
	cache        *Cache
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
		Sendchan:   make(chan string, 10),
		Recvchan:   make(chan string, 10),
		maxConns:   maxConns,
		cache:      NewCache(),
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

func (t *Server) loop() {
	for {
		msg := <-t.Sendchan
		log.Println("sending msg over the wire: ", msg)
		if err := gob.NewEncoder(t.outboundConn).Encode(&msg); err != nil {
			log.Println(err)
		}
	}
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
		log.Printf("number of client connected %v", t.users.Add(1))

		go t.handleConn(conn)
		go t.handleCommands(conn)
	}
}

func (t *Server) handleConn(conn net.Conn) {
	t.cache.Set(&ConnectedUsers{rank: int(t.users.Load()), id: conn.RemoteAddr().String(), conn: conn}, conn.RemoteAddr().String())
}

func (t *Server) writeConn(conn net.Conn, data []byte) {
	l, err := conn.Write(data)
	if err != nil {
		log.Print(err)
	}
	log.Printf("wrote %v bytes", l)
}

func (t *Server) handleCommands(conn net.Conn) {
	for {
		// incoming request
		buffer := make([]byte, 1024)
		g, err := conn.Read(buffer)
		if err != nil {
			log.Print(err)
			continue
		}
		log.Print(string(buffer[:g]))
		v, ok := t.cache.Get(conn.RemoteAddr().String())
		if !ok {
			log.Println("error getting client from cache")
		}

		cmd := strings.TrimSpace(string(buffer[:g]))
		switch cmd {
		case "hello":
			if v.rank != 1 {
				c := fmt.Sprintf("command %v rejected, rank not allowed  your rank is %v", cmd, v.rank)
				t.writeConn(v.conn, []byte(c))
				return
			}
			for _, v := range t.cache.data {
				t.writeConn(v.conn, []byte(cmd))
			}
		case "print":
			if v.rank != 1 {
				c := fmt.Sprintf("command %v rejected, rank not allowed  your rank is %v", cmd, v.rank)
				t.writeConn(v.conn, []byte(c))
				return
			}
			for _, v := range t.cache.data {
				t.writeConn(v.conn, []byte(cmd))
			}

		default:
			cmd := fmt.Sprintf("unknown command %v", string(buffer))
			for _, v := range t.cache.data {
				t.writeConn(v.conn, []byte(cmd))
			}

		}

	}

}
