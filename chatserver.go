package main

import (
	"fmt"
	"github.com/labstack/gommon/log"
	"github.com/satori/go.uuid"
	"net"
	"os"
	"strings"
	"time"
)

const (
	Disconnected = 0
	Connected    = 1
)

type ClientSession struct {
	Id       uuid.UUID
	Conn     net.Conn
	Receive  chan string
	Transmit chan<- string
	State    uint
}

// implement a chat server
func main() {
	//log.EnableColor()

	// startup
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Printf("failed to start server: %v", err)
		os.Exit(1)
	}
	log.Info("Server started on", ln.Addr())

	// channel to share received messages from clients to all other clients
	fromClients := make(chan string, 10)
	// session management
	clientSessions := make(map[uuid.UUID]*ClientSession, 0)
	go sessionManager(fromClients, clientSessions)

	// monitor connected clients periodically for debugging
	go func(clientSessions map[uuid.UUID]*ClientSession) {
		for {
			b := strings.Builder{}
			fmt.Fprintf(&b, "connected clients: %v", len(clientSessions))
			for _, session := range clientSessions {
				fmt.Fprintf(&b, "\n%+v", session)
			}
			log.Info(b.String())
			b.Reset()
			time.Sleep(time.Second * 5)
		}
	}(clientSessions)

	// handle incoming connections from clients and
	// create a new session and goroutine for each
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("error accepting: ", err)
			continue
		}
		id := uuid.NewV4()
		newClientSession := ClientSession{
			Id:       id,
			Conn:     conn,
			Receive:  make(chan string, 10),
			Transmit: fromClients,
			State:    Connected}
		clientSessions[id] = &newClientSession
		go clientHandler(&newClientSession)
	}
}

// manages client sessions:
// - routing messages between clients
// - reaping disconnected clients
func sessionManager(fromClients <-chan string, clientSessions map[uuid.UUID]*ClientSession) {
	for {
		msg := <-fromClients
		log.Info("got message from a client:", msg)

		for id, session := range clientSessions {
			log.Infof("sessionManager: handling %+v", session)
			if session.State == Disconnected {
				// reap sessions of disconnected clients to stop
				// trying to send to them. otherwise will eventually
				// deadlock when the session.Receive channel fills up
				log.Infof("%+v is Disconnected, reaping", session)
				delete(clientSessions, id)
			} else {
				log.Infof("sending %s to client %v", msg, session.Id)
				session.Receive <- msg
			}
		}
	}
}

// handles a single client:
// - read data sent from client and pass to session manager for distribution
//   to other clients
// - write data out to client from the session manager (usually messages from
//   other connected clients)
func clientHandler(session *ClientSession) {
	defer session.Conn.Close()
	log.Info("Connection established for", session.Conn.RemoteAddr())

	var msg string
	var incoming ReadResult
	readChan := make(chan ReadResult, 10)
	defer close(readChan)
	go readToChan(session.Conn, readChan)
	for {
		// https://stackoverflow.com/questions/34717331/why-does-conn-read-write-nothing-into-a-byte-but-bufio-reader-readstring
		select {
		case incoming = <-readChan:
			// handle stuff received from the connected client

			if incoming.Err != nil {
				log.Error(incoming.Err)
				session.State = Disconnected
				session.Transmit <- fmt.Sprintf("%s disconnected\n", session.Conn.RemoteAddr())
				return
			}

			msg = string(incoming.Data)
			log.Info("received: ", msg)
			// echo it back to all clients
			response := fmt.Sprintf("%s: %s", session.Conn.RemoteAddr(), msg)
			session.Transmit <- response

		case msg = <-session.Receive:
			// write received stuff from other clients to *this* client
			log.Info("got:", msg)
			if _, err := session.Conn.Write([]byte(msg)); err != nil {
				// probably dead connection?
				session.State = Disconnected
				session.Transmit <- fmt.Sprintf("%s disconnected\n", session.Conn.RemoteAddr())
				return
			}
		}
	}
}

type ReadResult struct {
	Data []byte
	Err  error
}

// abstraction over reading on net.Conn to avoid blocking
// in the clientHandler
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
	log.Info("read loop EXITED for ", conn.RemoteAddr())
}
