package main

import (
	"fmt"
	"github.com/labstack/gommon/log"
	"github.com/satori/go.uuid"
	"net"
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
	log.EnableColor()
	fromClients := make(chan string, 10)
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		// handle error...?
		log.Errorf("rut roh: %v", err)
	}
	log.Info("Server started on", ln.Addr())

	clientSessions := make(map[uuid.UUID]*ClientSession, 0)
	go router(fromClients, clientSessions)

	// monitor connected clients periodically for debugging
	go func(clientSessions map[uuid.UUID]*ClientSession) {
		for {
			log.Info("connected clients:", len(clientSessions))
			for id, session := range clientSessions {
				log.Infof("%s: %+v", id, session)
			}
			time.Sleep(time.Second * 5)
		}
	}(clientSessions)

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
		go handleConnection(&newClientSession)
	}
}

func router(fromClients <-chan string, clientSessions map[uuid.UUID]*ClientSession) {
	for {
		msg := <-fromClients
		log.Info("got message from a client:", msg)
		for id, session := range clientSessions {
			log.Infof("router: handling %s: %+v", id, session)
			if session.State == Disconnected {
				// reap sessions of disconnected clients to stop
				// trying to send to them. otherwise will eventually
				// deadlock when the session.Receive channel fills up
				log.Infof("%s: %+v is Disconnected, reaping", id, session)
				delete(clientSessions, id)
			} else {
				log.Infof("sending %s to client %d", msg, id)
				session.Receive <- msg
			}
		}
	}
}

func handleConnection(session *ClientSession) {
	defer session.Conn.Close()
	log.Info("Connection established for ", session.Conn.RemoteAddr())
	//var buf = new([]byte)
	//TODO: hmm, what's difference between new([]byte) and make([]byte, SIZE)?

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
				// TODO...could be connection closed, or something else...
				session.Transmit <- fmt.Sprintf("%s disconnected\n", session.Conn.RemoteAddr())
				session.State = Disconnected
				log.Infof("should reap: %+v", session)
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
				log.Infof("should reap: %+v", session)
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
