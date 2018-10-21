package main

import (
	"github.com/gopherjs/gopherjs/js"
	"github.com/lologarithm/netgen/example/client/clientnet"
	"github.com/lologarithm/netgen/example/newmodels"
)

func main() {
	print("loaded")
	js.Global.Set("ngex", map[string]interface{}{
		"newClient": newClient,
	})
}

func newClient() *js.Object {
	c := &clientnet.Client{
		CEvents: &clientnet.Events{},
	}
	go runClient(c)
	return js.MakeWrapper(c)
}

func runClient(c *clientnet.Client) {
	for packet := range c.Incoming {
		if packet == nil {
			break
		}
		switch packet.Header.MsgType {
		case newmodels.MessageMsgType:
			print("Got message: ", packet.NetMsg.(*newmodels.Message).Message, "\n")
		}
	}
	print("Got nil packet, shutting down client reader.\n")
}
