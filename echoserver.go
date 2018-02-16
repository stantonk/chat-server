package main

import (
	"github.com/labstack/gommon/log"
	"time"
	"net"
)

// Programming challenge: implement an echo server
func main() {
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		// handle error
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			// handle error
		}
		go handleConnection(conn)
	}
}
func handleConnection(conn net.Conn) {
	//var buf = new([]byte)
	//TODO: hmm, what's difference between new([]byte) and make([]byte, SIZE)?
	var buf = make([]byte, 10)
	for {
		time.Sleep(1e9)
		nb, err := conn.Read(buf)
		if err != nil {
			log.Error("error:", err)
			panic("oh well")
		}
		if nb > 0 {
			log.Info(string(buf[:nb]))
		} else {
			log.Info("nothing received yet...")
		}

	}

}
