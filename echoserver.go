package main

import (
	"github.com/labstack/gommon/log"
	"time"
	"net"
	"fmt"
)

// Programming challenge: implement an echo server
func main() {
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		// handle error
	}
	log.Info("Server started on", ln.Addr())

	for {
		conn, err := ln.Accept()
		if err != nil {
			// handle error
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	log.Info("Connection established for ", conn.RemoteAddr())
	//var buf = new([]byte)
	//TODO: hmm, what's difference between new([]byte) and make([]byte, SIZE)?
	var buf = make([]byte, 1024)
	for {
		time.Sleep(1e9)
		// https://stackoverflow.com/questions/34717331/why-does-conn-read-write-nothing-into-a-byte-but-bufio-reader-readstring
		if nb, err := conn.Read(buf); err != nil {
			log.Warn("error:", err)
			// TODO...could be connection closed, or something else...
			break
		} else {
			log.Info(string(buf[:nb]))
			// echo it back
			response := []byte(fmt.Sprintf("%s: %s", "echo", buf[:nb]))
			if _, err := conn.Write(response); err != nil {
				log.Error("unable to write data back to client: ", err)
				// probably dead connection?
				break
			}
		}
	}
	log.Info("Done handling connection for ", conn.RemoteAddr())
}
