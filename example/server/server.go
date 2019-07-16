package main

import (
	"fmt"
	"sync"

	"github.com/lologarithm/netgen/example/models"
	"github.com/lologarithm/netgen/lib/ngservice/client"
)

type server struct {
	mut   *sync.Mutex
	conns []*client.Client
}

// runClient is the server client closure.
// It holds references to the outbound/incoming
func runClient(c *client.Client, ss *server) {
	client.ManageClient(models.Context, c)
	for msg := range c.Incoming {
		if msg == nil {
			fmt.Printf("%s: Nil packet, starting shutdown of client conn.\n", c.Name)
			break
		}
		switch tmsg := msg.(type) {
		case *models.Message:
			fmt.Printf(" Got message: %s\n", tmsg.Message)
			c.Outgoing <- msg // ECHO
		case *models.VersionedMessage:
			fmt.Printf(" Got versioned message: %#v\n", tmsg)
			c.Outgoing <- &models.VersionedMessage{Message: tmsg.Message + "(ECHO)", From: "The Server", UselessData: 7}
		default:
			fmt.Printf("Got a message from the client: %#v\n", msg)
		}
	}

	fmt.Printf("%s: Socket closed, shutting down parser.\n", c.Name)
}
