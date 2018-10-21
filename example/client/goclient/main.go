package main

import (
	"fmt"
	"log"
	"os"

	"github.com/lologarithm/netgen/example/client/clientnet"
	"github.com/lologarithm/netgen/example/newmodels"
	"github.com/lologarithm/netgen/lib/ngen"
)

func main() {
	c := &clientnet.Client{
		CEvents: &clientnet.Events{},
	}
	conn := make(chan struct{})
	c.CEvents.OnConnected(func(v bool) {
		if !v {
			log.Fatalf("Failed to connect.")
		}
		conn <- struct{}{}
	})
	c.Dial("")
	<-conn
	c.Outgoing <- ngen.NewPacket(newmodels.Message{
		Message: "HELLOOOO",
	})

	for packet := range c.Incoming {
		if packet == nil {
			break
		}
		switch packet.Header.MsgType {
		case newmodels.MessageMsgType:
			fmt.Printf("Got message: %s\n", packet.NetMsg.(*newmodels.Message).Message)
			os.Exit(0)
		}
	}
}
