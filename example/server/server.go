package main

import (
	"fmt"
	"sync"

	"github.com/lologarithm/netgen/example/models"
	"github.com/lologarithm/netgen/lib/ngen/client"
)

type server struct {
	mut   *sync.Mutex
	conns []*client.Client
}

// runClient is the server client closure.
// It holds references to the outbound/incoming
func runClient(c *client.Client, ss *server) {
	go client.Sender(c)
	go client.Reader(c, models.ParseNetMessage)

	for packet := range c.Incoming {
		if packet == nil {
			fmt.Printf("%s: Nil packet, starting shutdown of client conn.\n", c.Name)
			break
		}
		switch tmsg := packet.NetMsg.(type) {
		case *models.Message:
			fmt.Printf(" Got message: %s\n", tmsg.Message)
			c.Outgoing <- packet // ECHO FOR SOME REASON
		}
	}

	fmt.Printf("%s: Socket closed, shutting down parser.\n", c.Name)
}
