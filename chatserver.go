package main

import (
	"github.com/labstack/gommon/log"
	"net"
	"fmt"
	"os"
)

// implement a chat server
func main() {
	log.EnableColor()
	transmit := make(chan string, 10)
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		// handle error
	}
	log.Info("Server started on", ln.Addr())

	clientReceiveChannels := make([]chan<- string, 0)
	clientReceiveChannelsRef := &clientReceiveChannels
	go router(transmit, clientReceiveChannelsRef)

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("error accepting: ", err)
			//todo: probably dont want to exit here... :)
			os.Exit(1)
		}
		newClientReceiveChannel := make(chan string, 10)
		clientReceiveChannels = append(clientReceiveChannels, newClientReceiveChannel)
		clientReceiveChannelsRef = &clientReceiveChannels
		go handleConnection(conn, transmit, newClientReceiveChannel)
	}
}

//TODO: need to reap disconnected clients from toClients or else we will eventually deadlock
func router(fromClients <-chan string, toClients *[]chan<- string) {
	for {
		msg := <-fromClients
		log.Info("got message from a client:", msg)
		for i := 0; i < len(*toClients); i++ {
			log.Infof("sending %s to client %d", msg, i)
			(*toClients)[i] <- msg
		}
	}
}

func handleConnection(conn net.Conn, transmit chan<- string, receive <-chan string) {
	defer conn.Close()
	log.Info("Connection established for ", conn.RemoteAddr())
	//var buf = new([]byte)
	//TODO: hmm, what's difference between new([]byte) and make([]byte, SIZE)?

	var msg string
	var incoming ReadResult
	readChan := make(chan ReadResult, 10)
	defer close(readChan)
	go readToChan(conn, readChan)
	for {

		// https://stackoverflow.com/questions/34717331/why-does-conn-read-write-nothing-into-a-byte-but-bufio-reader-readstring
		select {
		case incoming = <-readChan:
			// handle stuff received from the connected client
			if incoming.Err != nil {
				log.Error(incoming.Err)
				// TODO...could be connection closed, or something else...
				transmit <- fmt.Sprintf("%s disconnected\n", conn.RemoteAddr())
				return
			}
			msg = string(incoming.Data)
			log.Info("received: ", msg)
			// echo it back to all clients
			response := fmt.Sprintf("%s: %s", conn.RemoteAddr(), msg)
			transmit <- response

		case msg = <-receive:
			// write received stuff from other clients to *this* client
			log.Info("got:", msg)
			if _, err := conn.Write([]byte(msg)); err != nil {
				// probably dead connection?
				log.Info(conn.RemoteAddr(), "is dead, sending disconnection message")
				transmit <- fmt.Sprintf("%s disconnected\n", conn.RemoteAddr())
				return
			}
		}
	}
}

type ReadResult struct {
	Data []byte
	Err  error
}

func readToChan(conn net.Conn, rx chan<- ReadResult) {
	buf := make([]byte, 1024)

	for {
		log.Info("read loop for", conn.RemoteAddr())
		if nb, err := conn.Read(buf); err != nil {
			rx <- ReadResult{Err: err}
			break
		} else {
			tmp := make([]byte, nb)
			copy(tmp, buf[:nb])
			rx <- ReadResult{Data: tmp, Err: nil}
		}
	}
	log.Warn("read loop EXITED for ", conn.RemoteAddr())
}
