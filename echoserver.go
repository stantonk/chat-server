package main

import (
	"github.com/labstack/gommon/log"
	"net"
	"fmt"
	"time"
)

// Programming challenge: implement an echo server
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
			// handle error
		}
		newClientReceiveChannel := make(chan string, 10)
		clientReceiveChannels = append(clientReceiveChannels, newClientReceiveChannel)
		clientReceiveChannelsRef = &clientReceiveChannels
		go handleConnection(conn, transmit, newClientReceiveChannel)
	}
}

func router(fromClients <-chan string, toClients *[]chan<- string) {
	for {
		msg := <-fromClients
		log.Info("got message from a client:", msg)
		for i:=0; i < len(*toClients); i++ {
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

	var buf = make([]byte, 1024)
	var msg string
	for {

		// https://stackoverflow.com/questions/34717331/why-does-conn-read-write-nothing-into-a-byte-but-bufio-reader-readstring
		conn.SetReadDeadline(time.Now().Add(time.Second*1))
		conn.SetWriteDeadline(time.Now().Add(time.Second*5))

		if nb, err := conn.Read(buf); err != nil {
			if err, ok := err.(net.Error); ok && err.Timeout() {
				// if we timeout for the client to send, lets check if we got anything
				// from any other clients
				select {
				case msg = <-receive:
					log.Info("got:", msg)
					if _, err := conn.Write([]byte(msg)); err != nil {
						log.Error("unable to write data back to client: ", err)
						// probably dead connection?
						transmit <- fmt.Sprintf("%s disconnected", conn.RemoteAddr())
						break
					}
				case <-time.After(time.Second):
					//log.Info("timedout")
				}
			} else {
				log.Error(err)
				// TODO...could be connection closed, or something else...
				transmit <- fmt.Sprintf("%s disconnected", conn.RemoteAddr())
				break
			}
		} else {
			msg = string(buf[:nb])
			log.Info(msg)
			// echo it back
			response := fmt.Sprintf("%s: %s", conn.RemoteAddr(), msg)
			transmit <- response
		}
	}
	log.Info("Done handling connection for ", conn.RemoteAddr())
}
