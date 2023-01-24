package main

import (
	"flag"
	"log"
)

const (
	// defaultMaxConn is the default number of max connections the
	// server will handle. 0 means no limits will be set, so the
	// server will be bound by system resources.
	defaultMaxConn = 0
)

func main() {

	c := flag.Int("c", defaultMaxConn, "maximum number of client connections the server will accept, 0 means unlimited")
	flag.Parse()

	channelLocal, err := New("localhost:3000", ":4000", *c)
	if err != nil {
		log.Fatal(err)
	}

	channelLocal.acceptLoop()

	msg := <-channelLocal.Recvchan
	log.Print(msg)

}
