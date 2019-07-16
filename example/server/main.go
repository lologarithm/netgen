package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"

	"github.com/lologarithm/netgen/lib/ngservice/client"
	"github.com/lologarithm/netgen/lib/ngservice/client/ngwebsocket"
	"golang.org/x/net/websocket"
)

func main() {
	server := &server{
		mut:   &sync.Mutex{},
		conns: []*client.Client{},
	}

	http.Handle("/ws", websocket.Handler(func(conn *websocket.Conn) {
		log.Printf("Accepting socket from %#v.", conn.RemoteAddr())
		client := ngwebsocket.AcceptConn(conn)
		runClient(client, server)
	}))

	go func() {
		err := http.ListenAndServe(":4567", nil)
		if err != nil {
			fmt.Printf("Error starting server: %s\n", err)
		}
	}()

	fmt.Printf("Started. Press CTRL+C to exit.\n")
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}
